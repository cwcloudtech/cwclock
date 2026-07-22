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
// them to the target's own S3/Google Drive/git connection (see
// models.ExportTarget.Connection - captured through the same fields as an
// organization's external connections, but stored independently in the
// job, not looked up from the organization).
type ExportDeliveryService struct {
	mailer *email.Sender
	orgs   *store.OrgStore
}

func NewExportDeliveryService(mailer *email.Sender, orgs *store.OrgStore) *ExportDeliveryService {
	return &ExportDeliveryService{mailer: mailer, orgs: orgs}
}

func (d *ExportDeliveryService) Deliver(ctx context.Context, orgID, jobName string, startTime, endTime time.Time, target models.ExportTarget, reports []scheduler.ExportReportFile) error {
	switch target.Type {
	case "email":
		return d.deliverEmail(ctx, orgID, jobName, startTime, endTime, target, reports)
	case "s3", "google_drive", "git":
		return d.deliverExternalConnection(ctx, target, reports)
	default:
		return fmt.Errorf("unsupported export target type %q", target.Type)
	}
}

func (d *ExportDeliveryService) deliverEmail(ctx context.Context, orgID, jobName string, startTime, endTime time.Time, target models.ExportTarget, reports []scheduler.ExportReportFile) error {
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
	d.mailer.SendExportJob(ctx, to, utils.SplitList(target.CCEmails), org.ID, org.Name, jobName, startTime, endTime, attachments)
	return nil
}

func (d *ExportDeliveryService) deliverExternalConnection(ctx context.Context, target models.ExportTarget, reports []scheduler.ExportReportFile) error {
	if target.Connection == nil {
		return fmt.Errorf("%s target has no connection configured", target.Type)
	}

	dest, err := externalconn.BuildTarget(*target.Connection)
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
