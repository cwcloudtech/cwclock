package externalconn

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitSSH "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"golang.org/x/crypto/ssh"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

// gitTarget pushes invoice PDFs to a git repository (see ai-instruct-68),
// authenticating with either username/password (over HTTP(S)) or an SSH
// private key. Unlike s3Target/driveTarget it has no persistent connection
// to reuse: each Upload/Delete does its own shallow clone into an in-memory
// filesystem, commits the change and pushes it, since invoices are
// generated/removed infrequently enough that a persistent local clone isn't
// worth the complexity.
type gitTarget struct {
	repoURL string
	auth    transport.AuthMethod
	flat    bool
}

func newGitTarget(conn models.ExternalConnection) (*gitTarget, error) {
	auth, err := gitAuthMethod(conn)
	if err != nil {
		return nil, fmt.Errorf("external connection git: %w", err)
	}
	return &gitTarget{repoURL: conn.RepoURL, auth: auth, flat: conn.FlatDirectory}, nil
}

// gitAuthMethod picks SSH key auth when an SSH private key is configured,
// otherwise HTTP basic auth - validateExternalConnections (ai-instruct-68)
// enforces that exactly one of the two is ever set on a saved connection.
func gitAuthMethod(conn models.ExternalConnection) (transport.AuthMethod, error) {
	if utils.IsNotBlank(conn.SSHPrivateKey) {
		keys, err := gitSSH.NewPublicKeys("git", []byte(conn.SSHPrivateKey), conn.SSHPrivateKeyPassphrase)
		if err != nil {
			return nil, fmt.Errorf("invalid SSH private key: %w", err)
		}
		// The server has no interactive prompt to accept an unknown host key
		// and no pre-seeded known_hosts file, so host key verification is
		// skipped here - the connection is still authenticated by the
		// private key itself.
		keys.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		return keys, nil
	}
	return &gitHTTP.BasicAuth{Username: conn.Username, Password: conn.Password}, nil
}

func (g *gitTarget) Upload(ctx context.Context, year string, months []string, filename string, data []byte) error {
	repo, wt, err := g.clone(ctx)
	if err != nil {
		return err
	}

	dir := g.resolveDir(wt, year, months)
	if dir != "" {
		if err := wt.Filesystem.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("external connection git: could not create %s: %w", dir, err)
		}
	}

	filePath := path.Join(dir, filename)
	f, err := wt.Filesystem.Create(filePath)
	if err != nil {
		return fmt.Errorf("external connection git: could not create %s: %w", filePath, err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("external connection git: could not write %s: %w", filePath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("external connection git: could not write %s: %w", filePath, err)
	}

	if _, err := wt.Add(filePath); err != nil {
		return fmt.Errorf("external connection git: could not stage %s: %w", filePath, err)
	}

	return g.commitAndPush(ctx, repo, wt, "Add invoice "+invoiceID(filename))
}

func (g *gitTarget) Delete(ctx context.Context, year string, months []string, filename string) error {
	repo, wt, err := g.clone(ctx)
	if err != nil {
		return err
	}

	filePath, found := g.findExisting(wt, year, months, filename)
	if !found {
		return nil
	}

	if _, err := wt.Remove(filePath); err != nil {
		return fmt.Errorf("external connection git: could not remove %s: %w", filePath, err)
	}

	return g.commitAndPush(ctx, repo, wt, "Remove invoice "+invoiceID(filename))
}

// clone shallow-clones the repository into a fresh in-memory filesystem,
// fetching only its default branch's latest commit - the full history isn't
// needed just to add or remove one file.
func (g *gitTarget) clone(ctx context.Context) (*git.Repository, *git.Worktree, error) {
	repo, err := git.CloneContext(ctx, memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:          g.repoURL,
		Auth:         g.auth,
		Depth:        1,
		SingleBranch: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("external connection git: clone failed: %w", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("external connection git: could not open worktree: %w", err)
	}
	return repo, wt, nil
}

// resolveDir returns the directory Upload should write filename's data
// into: the repo root in flat mode (ai-instruct-42), or whichever of
// year/months' candidate folders already exists, defaulting to
// year/months[0] when none do - the same rule s3Target.resolveKey and
// driveTarget.ensureTargetFolder apply for their own providers.
func (g *gitTarget) resolveDir(wt *git.Worktree, year string, months []string) string {
	if g.flat {
		return ""
	}
	for _, month := range months {
		dir := path.Join(year, month)
		if info, err := wt.Filesystem.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return path.Join(year, months[0])
}

// findExisting locates filename already committed under one of months'
// candidate folders (or at the repo root in flat mode), mirroring
// driveTarget.findTargetFolder/findFile.
func (g *gitTarget) findExisting(wt *git.Worktree, year string, months []string, filename string) (string, bool) {
	if g.flat {
		if _, err := wt.Filesystem.Stat(filename); err == nil {
			return filename, true
		}
		return utils.EMPTY, false
	}
	for _, month := range months {
		p := path.Join(year, month, filename)
		if _, err := wt.Filesystem.Stat(p); err == nil {
			return p, true
		}
	}
	return utils.EMPTY, false
}

// commitAndPush commits the worktree's staged changes with message and
// pushes them, treating "nothing changed" (an identical file re-uploaded, or
// a push with nothing new to send) as success rather than an error.
func (g *gitTarget) commitAndPush(ctx context.Context, repo *git.Repository, wt *git.Worktree, message string) error {
	_, err := wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{Name: "CWClock", Email: "cwclock@localhost", When: time.Now()},
	})
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return nil
		}
		return fmt.Errorf("external connection git: commit failed: %w", err)
	}

	if err := repo.PushContext(ctx, &git.PushOptions{Auth: g.auth}); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("external connection git: push failed: %w", err)
	}
	return nil
}

// invoiceID recovers the invoice number the commit message refers to from
// filename, which every caller in invoice_handler.go builds as
// "{invoiceNumber}.pdf".
func invoiceID(filename string) string {
	return strings.TrimSuffix(filename, path.Ext(filename))
}
