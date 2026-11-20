// Package aidownload is the short-lived file store behind the AI assistant's
// downloadable reports and invoices (ai-instruct-132 follow-up).
//
// When the assistant generates a PDF (a report export or an invoice), the bytes
// can't go back to the model and the user shouldn't have to hunt for them in the
// UI. Instead the file is written to a random, per-user folder here and the agent
// is handed a download URL to post as a clickable link — much like ChatGPT's file
// links. The folder name is `{userID}-ttl-{random}`: the user id scopes the file
// (and drives cleanup), and the unguessable random component is the capability
// that stands in for a token, so the link works when clicked straight from the
// chat in any browser. Everything is swept on a TTL, so nothing lingers.
package aidownload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// idSeparator sits between the owner's user id and the random capability in a
// folder name: `{userID}-ttl-{random}`. A user id is a UUID (which contains
// hyphens), so the separator is distinctive enough to split on reliably.
const idSeparator = "-ttl-"

var (
	// ErrNotFound: no such download (never existed, or already swept).
	ErrNotFound = errors.New("aidownload: not found")
	// ErrExpired: the download is older than the TTL (and has been removed).
	ErrExpired = errors.New("aidownload: expired")
	// ErrInvalidID: the id isn't a well-formed `{userID}-ttl-{random}` folder name.
	ErrInvalidID = errors.New("aidownload: invalid id")
)

// idPattern is the only shape an id may take: a user id and a random, both drawn
// from a safe alphabet, joined by the separator. This is the first guard against
// path traversal — an id can carry neither '/' nor '.'.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9-]+` + regexp.QuoteMeta(idSeparator) + `[A-Za-z0-9]+$`)

// Store writes generated files under baseDir, one random folder per file, and
// resolves/sweeps them. Safe for concurrent use (the filesystem is the state).
type Store struct {
	baseDir    string
	ttl        time.Duration
	sweep      time.Duration
	maxBytes   int64
	publicBase string // e.g. https://api.example.com — the base of the download URL
	log        *slog.Logger
}

// New builds a Store. baseDir is created lazily on the first Save. sweep is how
// often the purge job runs. publicBase is the user-facing base URL of this API
// (CWCLOCK_API_URL), used to build the links the agent posts — NOT the loopback
// base the tool calls the API with.
func New(baseDir string, ttl, sweep time.Duration, maxBytes int64, publicBase string, log *slog.Logger) *Store {
	if log == nil {
		log = slog.Default()
	}
	return &Store{
		baseDir:    baseDir,
		ttl:        ttl,
		sweep:      sweep,
		maxBytes:   maxBytes,
		publicBase: strings.TrimRight(publicBase, "/"),
		log:        log,
	}
}

// MaxBytes is the per-file cap (so the caller can bound how much it reads before
// handing the bytes over). Zero/negative means "no explicit cap".
func (s *Store) MaxBytes() int64 { return s.maxBytes }

// SweepInterval is how often the background purge job runs (the configured value;
// falls back to one minute if misconfigured, so the ticker never gets a non-
// positive period).
func (s *Store) SweepInterval() time.Duration {
	if s.sweep <= 0 {
		return time.Minute
	}
	return s.sweep
}

// Save writes data to `{baseDir}/{userID}-ttl-{random}/{filename}` and returns
// the public download URL plus the folder id. filename is reduced to a safe base
// name; userID must be a plain id (a UUID), never a path.
func (s *Store) Save(userID, filename string, data []byte) (url, id string, err error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || strings.ContainsAny(userID, `/\`) || strings.Contains(userID, "..") {
		return "", "", fmt.Errorf("aidownload: bad user id")
	}

	name := safeFilename(filename)

	random, err := randomToken()
	if err != nil {
		return "", "", err
	}
	id = userID + idSeparator + random

	dir := filepath.Join(s.baseDir, id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", fmt.Errorf("aidownload: creating folder: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
		// Don't leave an empty folder behind on a write failure.
		_ = os.RemoveAll(dir)
		return "", "", fmt.Errorf("aidownload: writing file: %w", err)
	}

	return s.publicBase + "/v1/ai-downloads/" + id, id, nil
}

// Open resolves an id to the one file inside its folder, for serving. It returns
// the absolute path, the stored filename and its content type. A folder older
// than the TTL is treated as expired (and removed), so a stale link never serves
// even if the sweeper hasn't run yet.
func (s *Store) Open(id string) (path, filename, contentType string, err error) {
	id = strings.TrimSpace(id)
	if !idPattern.MatchString(id) {
		return "", "", "", ErrInvalidID
	}

	dir := filepath.Join(s.baseDir, id)
	// Defence in depth: the resolved path must still sit directly under baseDir.
	if filepath.Dir(dir) != filepath.Clean(s.baseDir) {
		return "", "", "", ErrInvalidID
	}

	info, statErr := os.Stat(dir)
	if statErr != nil || !info.IsDir() {
		return "", "", "", ErrNotFound
	}
	if s.ttl > 0 && time.Since(info.ModTime()) > s.ttl {
		_ = os.RemoveAll(dir)
		return "", "", "", ErrExpired
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		return "", "", "", ErrNotFound
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		return filepath.Join(dir, name), name, contentTypeFor(name), nil
	}
	return "", "", "", ErrNotFound
}

// Sweep removes every download folder older than the TTL. Returns how many it
// removed. A TTL of zero disables sweeping (nothing is ever old enough).
func (s *Store) Sweep() (int, error) {
	if s.ttl <= 0 {
		return 0, nil
	}
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // nothing generated yet
		}
		return 0, err
	}

	removed := 0
	cutoff := time.Now().Add(-s.ttl)
	for _, entry := range entries {
		if !entry.IsDir() || !strings.Contains(entry.Name(), idSeparator) {
			continue // leave anything that isn't one of our folders alone
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.RemoveAll(filepath.Join(s.baseDir, entry.Name())); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

// StartSweeper sweeps once now, then every SweepInterval until ctx is done. Meant
// to run in its own goroutine for the life of the process.
func (s *Store) StartSweeper(ctx context.Context) {
	sweep := func() {
		if n, err := s.Sweep(); err != nil {
			s.log.Warn("ai download sweep failed", "error", err)
		} else if n > 0 {
			s.log.Info("ai download sweep removed expired files", "count", n)
		}
	}

	sweep()
	ticker := time.NewTicker(s.SweepInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sweep()
		}
	}
}

// safeFilename reduces whatever the upstream Content-Disposition carried to a
// single, harmless base name: no path (so it can never escape the folder), and no
// quote/backslash/control character (so it's always safe to put in a
// `Content-Disposition: attachment; filename="…"` header later).
func safeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(filepath.FromSlash(name))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == '"' || r == '\\' {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(strings.Trim(name, `/\.`))
	if name == "" || name == "." || name == ".." {
		return "download"
	}
	return name
}

// randomToken is a 128-bit capability, hex-encoded: the unguessable part of a
// download id.
func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("aidownload: generating token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// contentTypeFor maps a filename to a content type, covering the two the
// assistant actually produces (PDF, CSV) and falling back for anything else.
func contentTypeFor(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf":
		return "application/pdf"
	case ".csv":
		return "text/csv; charset=utf-8"
	}
	if ct := mime.TypeByExtension(filepath.Ext(name)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}
