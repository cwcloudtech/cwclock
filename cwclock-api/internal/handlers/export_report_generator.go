package handlers

import (
	"context"
	"fmt"

	"cwclock-api/internal/scheduler"
	"cwclock-api/internal/store"
)

// ExportReportGenerator adapts ReportHandler's report builders to the
// scheduler.ExportReportGenerator interface, so a scheduled export job can
// produce the same summary/detailed PDF/CSV reports the Reports screen and
// invoice emails already generate.
type ExportReportGenerator struct {
	reports *ReportHandler
}

func NewExportReportGenerator(reports *ReportHandler) *ExportReportGenerator {
	return &ExportReportGenerator{reports: reports}
}

// GenerateReport builds one export job report attachment. includeFinancial
// mirrors canSeeAmount elsewhere in ReportHandler - export jobs are
// owner/admin-only (see router.go), the same roles allowed to see amounts,
// so it's the job's own choice to include them or not. startDate/endDate
// are already resolved by the scheduler (see ParseTimePeriod), so every
// report type in the same run shares the exact same range.
func (g *ExportReportGenerator) GenerateReport(ctx context.Context, reportType, orgID string, clientIDs, projectIDs []string, startDate, endDate string, includeFinancial bool) (scheduler.ExportReportFile, error) {
	filter := store.ReportFilter{Start: startDate, End: endDate, ClientIDs: clientIDs, ProjectIDs: projectIDs}
	if err := g.reports.checkReportSize(ctx, orgID, filter); err != nil {
		return scheduler.ExportReportFile{}, err
	}

	var data []byte
	var filename, mimeType string
	var err error
	switch reportType {
	case "summary-pdf":
		data, filename, err = g.reports.GenerateSummaryPDF(ctx, orgID, filter, includeFinancial, false)
		mimeType = "application/pdf"
	case "summary-pdf-portrait":
		data, filename, err = g.reports.GenerateSummaryPDF(ctx, orgID, filter, includeFinancial, true)
		mimeType = "application/pdf"
	case "summary-csv":
		data, filename, err = g.reports.GenerateSummaryCSV(ctx, orgID, filter, includeFinancial)
		mimeType = "text/csv"
	case "detailed-pdf":
		data, filename, err = g.reports.GenerateDetailedPDF(ctx, orgID, filter, includeFinancial, false)
		mimeType = "application/pdf"
	case "detailed-pdf-portrait":
		data, filename, err = g.reports.GenerateDetailedPDF(ctx, orgID, filter, includeFinancial, true)
		mimeType = "application/pdf"
	case "detailed-csv":
		data, filename, err = g.reports.GenerateDetailedCSV(ctx, orgID, filter, includeFinancial)
		mimeType = "text/csv"
	default:
		return scheduler.ExportReportFile{}, fmt.Errorf("unknown export report type %q", reportType)
	}
	if err != nil {
		return scheduler.ExportReportFile{}, err
	}
	return scheduler.ExportReportFile{Filename: filename, MimeType: mimeType, Data: data}, nil
}
