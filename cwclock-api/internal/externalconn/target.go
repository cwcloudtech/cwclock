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

// FolderPathLayout is the time.Format layout for an invoice's destination
// folder path: the year, then the two-digit month and its full English name
// (e.g. "2026/07.July"), per ai-instruct-39.
const FolderPathLayout = "2006/01.January"

// FolderPath returns the "YYYY/MM.MonthName" folder path an invoice
// generated/uploaded at t belongs under.
func FolderPath(t time.Time) string {
	return t.Format(FolderPathLayout)
}

// Target is one external storage destination an invoice PDF can be pushed
// to or removed from.
type Target interface {
	// Upload stores data as filename under the yearMonth ("2006/01.January")
	// folder,
	// creating any missing folder structure and replacing an existing file
	// of the same name.
	Upload(ctx context.Context, yearMonth, filename string, data []byte) error
	// Delete removes filename from the yearMonth folder, if present. A
	// missing folder or file is not an error.
	Delete(ctx context.Context, yearMonth, filename string) error
}

// BuildTarget constructs the Target for one external connection's
// configured type.
func BuildTarget(conn models.ExternalConnection) (Target, error) {
	switch conn.Type {
	case models.ExternalConnectionS3:
		return newS3Target(conn), nil
	case models.ExternalConnectionGoogleDrive:
		return newDriveTarget(conn)
	default:
		return nil, fmt.Errorf("unknown external connection type %q", conn.Type)
	}
}

// SyncUpload pushes data to every one of an organization's external
// connections, best-effort: a failing connection is logged and skipped
// rather than returned, since the invoice's own DB row (not these external
// copies) is the source of truth.
func SyncUpload(ctx context.Context, conns []models.ExternalConnection, yearMonth, filename string, data []byte) {
	for _, conn := range conns {
		target, err := BuildTarget(conn)
		if err != nil {
			slog.Error("external connection: unsupported type", "type", conn.Type, "connectionId", conn.ID, "error", err)
			continue
		}
		if err := target.Upload(ctx, yearMonth, filename, data); err != nil {
			slog.Error("external connection: upload failed", "type", conn.Type, "connectionId", conn.ID, "error", err)
		}
	}
}

// SyncDelete removes filename from every one of an organization's external
// connections, best-effort (see SyncUpload).
func SyncDelete(ctx context.Context, conns []models.ExternalConnection, yearMonth, filename string) {
	for _, conn := range conns {
		target, err := BuildTarget(conn)
		if err != nil {
			slog.Error("external connection: unsupported type", "type", conn.Type, "connectionId", conn.ID, "error", err)
			continue
		}
		if err := target.Delete(ctx, yearMonth, filename); err != nil {
			slog.Error("external connection: delete failed", "type", conn.Type, "connectionId", conn.ID, "error", err)
		}
	}
}
