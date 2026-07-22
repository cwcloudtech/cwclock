package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"cwclock-api/internal/email"
	"cwclock-api/internal/externalconn"
	"cwclock-api/internal/models"
	"cwclock-api/internal/scheduler"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

// ExportDeliveryService adapts email and external-connection delivery to
// the scheduler.ExportDeliveryService interface, sending a scheduled export
// job's generated reports to one target - either by email, or by pushing
// them to one of the organization's S3/Google Drive/git external
// connections (see externalconn), the same connections invoices already
// push to (see ai-instruct-39/ai-instruct-68).
type ExportDeliveryService struct {
	mailer *email.Sender
	orgs   *store.OrgStore
}

func NewExportDeliveryService(mailer *email.Sender, orgs *store.OrgStore) *ExportDeliveryService {
	return &ExportDeliveryService{mailer: mailer, orgs: orgs}
}

func (d *ExportDeliveryService) Deliver(ctx context.Context, orgID, jobName string, target models.ExportTarget, reports []scheduler.ExportReportFile) error {
	switch target.Type {
	case "email":
		return d.deliverEmail(ctx, orgID, jobName, target, reports)
	case "s3", "google_drive", "git":
		return d.deliverExternalConnection(ctx, orgID, target, reports)
	default:
		return fmt.Errorf("unsupported export target type %q", target.Type)
	}
}

func (d *ExportDeliveryService) deliverEmail(ctx context.Context, orgID, jobName string, target models.ExportTarget, reports []scheduler.ExportReportFile) error {
	to := utils.SplitList(target.ToEmails)
	if len(to) == 0 {
		return fmt.Errorf("email target has no recipients")
	}

	org, err := d.orgs.FindByID(ctx, orgID)
	if err != nil {
		return err
	}

	attachments := make([]email.Attachment, len(reports))
	for i, r := range reports {
		attachments[i] = email.Attachment{MimeType: r.MimeType, FileName: r.Filename, B64: base64.StdEncoding.EncodeToString(r.Data)}
	}
	d.mailer.SendExportJob(ctx, to, utils.SplitList(target.CCEmails), org.ID, org.Name, jobName, attachments)
	return nil
}

func (d *ExportDeliveryService) deliverExternalConnection(ctx context.Context, orgID string, target models.ExportTarget, reports []scheduler.ExportReportFile) error {
	org, err := d.orgs.FindByID(ctx, orgID)
	if err != nil {
		return err
	}

	conn, ok := findExternalConnection(org.ExternalConnections, target.Connection)
	if !ok {
		return fmt.Errorf("external connection %q not found", target.Connection)
	}

	dest, err := externalconn.BuildTarget(conn)
	if err != nil {
		return err
	}

	now := time.Now()
	year := externalconn.YearFolder(now)
	months := externalconn.MonthCandidates(now)
	for _, r := range reports {
		if err := dest.Upload(ctx, year, months, r.Filename, r.Data); err != nil {
			return fmt.Errorf("failed to upload %s: %w", r.Filename, err)
		}
	}
	return nil
}

func findExternalConnection(conns []models.ExternalConnection, id string) (models.ExternalConnection, bool) {
	for _, c := range conns {
		if c.ID == id {
			return c, true
		}
	}
	return models.ExternalConnection{}, false
}
