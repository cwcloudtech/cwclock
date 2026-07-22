package scheduler
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

type ExportJobScheduler struct {
	cron      *cron.Cron
	jobs      *store.ExportJobStore
	orgs      *store.OrgStore
	clients   *store.ClientStore
	projects  *store.ProjectStore
	entries   *store.TimeEntryStore
	reports   ExportReportGenerator
	emailer   ExportEmailSender
	entryIDs  map[string]cron.EntryID // Maps export job ID to cron entry ID
}

type ExportReportGenerator interface {
	GenerateReport(ctx context.Context, reportType string, orgID string, clientIDs []string, projectIDs []string, timePeriod string, includeFinancial bool) ([]byte, error)
}

type ExportEmailSender interface {
	SendExportEmail(ctx context.Context, targets []models.ExportTarget, reports map[string][]byte) error
}

func NewExportJobScheduler(
	jobs *store.ExportJobStore,
	orgs *store.OrgStore,
	clients *store.ClientStore,
	projects *store.ProjectStore,
	entries *store.TimeEntryStore,
	reportGenerator ExportReportGenerator,
	emailer ExportEmailSender,
) *ExportJobScheduler {
	return &ExportJobScheduler{
		cron:      cron.New(),
		jobs:      jobs,
		orgs:      orgs,
		clients:   clients,
		projects:  projects,
		entries:   entries,
		reports:   reportGenerator,
		emailer:   emailer,
		entryIDs:  make(map[string]cron.EntryID),
	}
}

func (s *ExportJobScheduler) Start(ctx context.Context) error {
	s.cron.Start()
	
	// Load all enabled export jobs for all organizations and schedule them
	// This is a simplified version - in production, you'd want to query all orgs
	// and their enabled jobs
	return nil
}

func (s *ExportJobScheduler) Stop() {
	<-s.cron.Stop().Done()
}

func (s *ExportJobScheduler) ScheduleJob(ctx context.Context, job models.ExportJob) error {
	if !job.Enabled {
		return nil
	}

	// Remove existing entry if present
	if entryID, exists := s.entryIDs[job.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.entryIDs, job.ID)
	}

	// Schedule the new job
	entryID, err := s.cron.AddFunc(job.CronExpression, func() {
		s.executeJob(ctx, job)
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
	log.Printf("Executing export job %s", job.ID)

	// Generate all requested reports
	reports := make(map[string][]byte)
	for _, reportType := range job.ReportTypes {
		reportData, err := s.reports.GenerateReport(
			ctx,
			reportType,
			job.OrganizationID,
			job.ClientIDs,
			job.ProjectIDs,
			job.TimePeriod,
			job.IncludeFinancial,
		)
		if err != nil {
			log.Printf("Failed to generate report %s for job %s: %v", reportType, job.ID, err)
			continue
		}
		reports[reportType] = reportData
	}

	// Send reports to all targets
	for _, target := range job.Targets {
		if err := s.emailer.SendExportEmail(ctx, []models.ExportTarget{target}, reports); err != nil {
			log.Printf("Failed to send export to target in job %s: %v", job.ID, err)
		}
	}

	log.Printf("Export job %s completed", job.ID)
}

// ParseTimePeriod converts a time period expression like "now()-1d" to actual dates
func ParseTimePeriod(period string) (startDate, endDate string, err error) {
	period = strings.TrimSpace(period)
	
	now := time.Now()
	endTime := now
	var startTime time.Time

	if strings.HasPrefix(period, "now()") {
		startTime = now
	} else if strings.HasPrefix(period, "now()-") {
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
	} else {
		return "", "", fmt.Errorf("unsupported time period format: %s", period)
	}

	// Return dates in YYYY-MM-DD format
	return startTime.Format("2006-01-02"), endTime.Format("2006-01-02"), nil
}
