package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

// ExportReportFile is one generated report attachment ready for delivery to
// an export job's targets.
type ExportReportFile struct {
	Filename string
	MimeType string
	Data     []byte
}

// ExportReportGenerator produces one report attachment for an export job's
// reportType ("summary-pdf", "summary-csv", "detailed-pdf", "detailed-csv").
// startDate/endDate are already-resolved "YYYY-MM-DD" bounds (see
// ParseTimePeriod) - resolved once per run so every report and the
// delivery email agree on the exact same range.
type ExportReportGenerator interface {
	GenerateReport(ctx context.Context, reportType string, orgID string, clientIDs, projectIDs []string, startDate, endDate string, includeFinancial bool) (ExportReportFile, error)
}

// ExportDeliveryService delivers a set of already-generated reports to one
// export target - by email, or by pushing them to the organization's
// S3/Google Drive/git external connection (see models.ExportTarget.Type).
// startDate/endDate are the run's resolved period, for the delivery email's
// subject/body.
type ExportDeliveryService interface {
	Deliver(ctx context.Context, orgID, jobName, startDate, endDate string, target models.ExportTarget, reports []ExportReportFile) error
}

// ExportJobScheduler runs every organization's enabled export jobs on their
// configured cron expression, generating the requested reports and
// delivering them to each of the job's targets.
type ExportJobScheduler struct {
	cron     *cron.Cron
	jobs     *store.ExportJobStore
	reports  ExportReportGenerator
	delivery ExportDeliveryService
	entryIDs map[string]cron.EntryID // Maps export job ID to cron entry ID
}

func NewExportJobScheduler(
	jobs *store.ExportJobStore,
	reportGenerator ExportReportGenerator,
	delivery ExportDeliveryService,
) *ExportJobScheduler {
	return &ExportJobScheduler{
		cron:     cron.New(),
		jobs:     jobs,
		reports:  reportGenerator,
		delivery: delivery,
		entryIDs: make(map[string]cron.EntryID),
	}
}

// Start starts the cron runner and schedules every currently-enabled export
// job across all organizations, so jobs created in a previous run of the
// API resume firing after a restart.
func (s *ExportJobScheduler) Start(ctx context.Context) error {
	s.cron.Start()

	jobs, err := s.jobs.ListAllEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to load enabled export jobs: %w", err)
	}
	for _, job := range jobs {
		if err := s.ScheduleJob(job); err != nil {
			log.Printf("export job %s: failed to schedule at startup: %v", job.ID, err)
		}
	}
	return nil
}

// ValidCronExpression reports whether expr parses as a standard 5-field
// cron expression, the same parser ScheduleJob's cron.AddFunc uses - so an
// invalid expression is rejected by the API at creation time rather than
// only failing silently once the scheduler tries to run it.
func ValidCronExpression(expr string) bool {
	_, err := cron.ParseStandard(expr)
	return err == nil
}

func (s *ExportJobScheduler) Stop() {
	<-s.cron.Stop().Done()
}

// ScheduleJob (re)schedules job on its cron expression, replacing any
// entry already scheduled for the same job ID - so it's safe to call again
// after an update. A disabled job is only unscheduled (a no-op if it wasn't
// scheduled to begin with).
func (s *ExportJobScheduler) ScheduleJob(job models.ExportJob) error {
	if entryID, exists := s.entryIDs[job.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.entryIDs, job.ID)
	}

	if !job.Enabled {
		return nil
	}

	// context.Background() rather than a caller-supplied context: this
	// closure fires later, on the cron runner's own goroutine, well after
	// whichever request (or startup call) scheduled it has returned - a
	// request-scoped context would already be canceled by then.
	entryID, err := s.cron.AddFunc(job.CronExpression, func() {
		s.executeJob(context.Background(), job)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule job %s: %w", job.ID, err)
	}

	s.entryIDs[job.ID] = entryID
	return nil
}

func (s *ExportJobScheduler) UnscheduleJob(jobID string) {
	if entryID, exists := s.entryIDs[jobID]; exists {
		s.cron.Remove(entryID)
		delete(s.entryIDs, jobID)
	}
}

func (s *ExportJobScheduler) executeJob(ctx context.Context, job models.ExportJob) {
	log.Printf("export job %s: starting", job.ID)

	startDate, endDate, err := ParseTimePeriod(job.TimePeriod)
	if err != nil {
		log.Printf("export job %s: invalid time period %q: %v", job.ID, job.TimePeriod, err)
		return
	}

	reports := make([]ExportReportFile, 0, len(job.ReportTypes))
	for _, reportType := range job.ReportTypes {
		file, err := s.reports.GenerateReport(ctx, reportType, job.OrganizationID, job.ClientIDs, job.ProjectIDs, startDate, endDate, job.IncludeFinancial)
		if err != nil {
			log.Printf("export job %s: failed to generate %s report: %v", job.ID, reportType, err)
			continue
		}
		reports = append(reports, file)
	}
	if len(reports) == 0 {
		log.Printf("export job %s: no reports generated, skipping delivery", job.ID)
		return
	}

	for _, target := range job.Targets {
		if err := s.delivery.Deliver(ctx, job.OrganizationID, job.Name, startDate, endDate, target, reports); err != nil {
			log.Printf("export job %s: failed to deliver to %s target: %v", job.ID, target.Type, err)
		}
	}

	log.Printf("export job %s: completed", job.ID)
}

// ParseTimePeriod converts a time period expression like "now()-1d" to actual dates
func ParseTimePeriod(period string) (startDate, endDate string, err error) {
	period = strings.TrimSpace(period)

	now := time.Now()
	endTime := now
	var startTime time.Time

	if strings.HasPrefix(period, "now()-") {
		// Parse expressions like "now()-1d", "now()-1h", "now()-30m", etc.
		suffix := strings.TrimPrefix(period, "now()-")

		if strings.HasSuffix(suffix, "d") {
			daysStr := strings.TrimSuffix(suffix, "d")
			var days int
			if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil {
				return "", "", fmt.Errorf("invalid day count: %s", daysStr)
			}
			startTime = now.AddDate(0, 0, -days)
		} else if strings.HasSuffix(suffix, "h") {
			hoursStr := strings.TrimSuffix(suffix, "h")
			var hours int
			if _, err := fmt.Sscanf(hoursStr, "%d", &hours); err != nil {
				return "", "", fmt.Errorf("invalid hour count: %s", hoursStr)
			}
			startTime = now.Add(-time.Duration(hours) * time.Hour)
		} else if strings.HasSuffix(suffix, "m") {
			minutesStr := strings.TrimSuffix(suffix, "m")
			var minutes int
			if _, err := fmt.Sscanf(minutesStr, "%d", &minutes); err != nil {
				return "", "", fmt.Errorf("invalid minute count: %s", minutesStr)
			}
			startTime = now.Add(-time.Duration(minutes) * time.Minute)
		} else {
			return "", "", fmt.Errorf("unsupported time period format: %s", period)
		}
	} else if strings.HasPrefix(period, "now()") {
		startTime = now
	} else {
		return "", "", fmt.Errorf("unsupported time period format: %s", period)
	}

	// Return dates in YYYY-MM-DD format
	return startTime.Format("2006-01-02"), endTime.Format("2006-01-02"), nil
}
