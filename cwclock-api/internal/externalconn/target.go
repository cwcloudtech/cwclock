// Package externalconn pushes generated invoice PDFs to an organization's
// optional external storage connections (S3-compatible object storage and
// Google Drive - see ai-instruct-39). Every connection is addressed through
// the same Target interface, keyed by a "YYYY/MM.MonthName" folder path and
// a filename, so the invoice handlers don't need to know which provider a
// given connection is.
package externalconn

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cwclock-api/internal/models"
)

// YearFolder returns the language-independent year folder ("2026") an
// invoice generated/uploaded at t belongs under.
func YearFolder(t time.Time) string {
	return t.Format("2006")
}

var frenchMonthNames = [12]string{
	"Janvier", "Février", "Mars", "Avril", "Mai", "Juin",
	"Juillet", "Août", "Septembre", "Octobre", "Novembre", "Décembre",
}

// MonthCandidates returns the month folder name candidates for t (e.g.
// "07.July" and "07.Juillet"), ordered with the default (English) name
// first. A Target reuses whichever candidate folder already exists rather
// than creating a duplicate in the other language, falling back to the
// first (English) candidate when creating a new folder.
func MonthCandidates(t time.Time) []string {
	return []string{
		t.Format("01.January"),
		fmt.Sprintf("%02d.%s", int(t.Month()), frenchMonthNames[t.Month()-1]),
	}
}

// Target is one external storage destination an invoice PDF can be pushed
// to or removed from.
type Target interface {
	// Upload stores data as filename under the year folder and whichever of
	// months' candidate month folders already exists (creating the first
	// candidate if none do), replacing an existing file of the same name.
	Upload(ctx context.Context, year string, months []string, filename string, data []byte) error
	// Delete removes filename from the year/month folder, if present. A
	// missing folder or file is not an error.
	Delete(ctx context.Context, year string, months []string, filename string) error
}

// BuildTarget constructs the Target for one external connection's
// configured type.
func BuildTarget(conn models.ExternalConnection) (Target, error) {
	switch conn.Type {
	case models.ExternalConnectionS3:
		return newS3Target(conn), nil
	case models.ExternalConnectionGoogleDrive:
		return newDriveTarget(conn)
	case models.ExternalConnectionGit:
		return newGitTarget(conn)
	default:
		return nil, fmt.Errorf("unknown external connection type %q", conn.Type)
	}
}

// syncOne runs one connection's op (upload/delete) in isolation so one
// failing connection never aborts the others - or the caller's request. A
// build error, an op error, and even a panic from a misbehaving provider
// client are all logged and swallowed, never propagated: the invoice's own
// DB row (not these external copies) is the source of truth, and there's no
// panic-recovery middleware upstream, so a panic here would otherwise drop
// the whole HTTP request.
func syncOne(action string, conn models.ExternalConnection, op func(Target) error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("external connection: "+action+" panicked", "type", conn.Type, "connectionId", conn.ID, "panic", r)
		}
	}()
	target, err := BuildTarget(conn)
	if err != nil {
		slog.Error("external connection: unsupported type", "type", conn.Type, "connectionId", conn.ID, "error", err)
		return
	}
	if err := op(target); err != nil {
		slog.Error("external connection: "+action+" failed", "type", conn.Type, "connectionId", conn.ID, "error", err)
	}
}

// SyncUpload pushes data to every one of an organization's external
// connections, best-effort: a failing connection (error or panic) is logged
// and skipped rather than returned, since the invoice's own DB row (not
// these external copies) is the source of truth.
func SyncUpload(ctx context.Context, conns []models.ExternalConnection, year string, months []string, filename string, data []byte) {
	for _, conn := range conns {
		syncOne("upload", conn, func(t Target) error {
			return t.Upload(ctx, year, months, filename, data)
		})
	}
}

// SyncDelete removes filename from every one of an organization's external
// connections, best-effort (see SyncUpload).
func SyncDelete(ctx context.Context, conns []models.ExternalConnection, year string, months []string, filename string) {
	for _, conn := range conns {
		syncOne("delete", conn, func(t Target) error {
			return t.Delete(ctx, year, months, filename)
		})
	}
}
