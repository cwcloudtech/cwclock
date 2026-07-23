package handlers

import (
	"context"
	"fmt"

	"cwclock-api/internal/scheduler"
	"cwclock-api/internal/store"
)

// ExportReportGenerator adapts ReportHandler's report builders (and the
// existing InvoiceStore, for "invoices-pdf") to the
// scheduler.ExportReportGenerator interface, so a scheduled export job can
// produce the same summary/detailed PDF/CSV reports the Reports screen and
// invoice emails already generate, or attach already-generated invoices.
type ExportReportGenerator struct {
	reports  *ReportHandler
	invoices *store.InvoiceStore
}

func NewExportReportGenerator(reports *ReportHandler, invoices *store.InvoiceStore) *ExportReportGenerator {
	return &ExportReportGenerator{reports: reports, invoices: invoices}
}

// GenerateReport builds the export job attachment(s) for one reportType.
// includeFinancial mirrors canSeeAmount elsewhere in ReportHandler - export
// jobs are owner/admin-only (see router.go), the same roles allowed to see
// amounts, so it's the job's own choice to include them or not.
// startDate/endDate are already resolved by the scheduler (see
// ParseTimePeriod), so every report type in the same run shares the exact
// same range. "invoices-pdf" is handled separately (see
// generateInvoicesPDFs) - it doesn't touch time entries at all, so it skips
// the report-size check and can return any number of files, not just one.
func (g *ExportReportGenerator) GenerateReport(ctx context.Context, reportType, orgID string, clientIDs, projectIDs []string, startDate, endDate string, includeFinancial bool) ([]scheduler.ExportReportFile, error) {
	if reportType == "invoices-pdf" {
		return g.generateInvoicesPDFs(ctx, orgID, clientIDs, startDate, endDate)
	}

	filter := store.ReportFilter{Start: startDate, End: endDate, ClientIDs: clientIDs, ProjectIDs: projectIDs}
	if err := g.reports.checkReportSize(ctx, orgID, filter); err != nil {
		return nil, err
	}

	var data []byte
	var filename, mimeType string
	var err error
	switch reportType {
	case "summary-pdf":
		data, filename, err = g.reports.GenerateSummaryPDF(ctx, orgID, filter, includeFinancial)
		mimeType = "application/pdf"
	case "summary-csv":
		data, filename, err = g.reports.GenerateSummaryCSV(ctx, orgID, filter, includeFinancial)
		mimeType = "text/csv"
	case "detailed-pdf":
		data, filename, err = g.reports.GenerateDetailedPDF(ctx, orgID, filter, includeFinancial)
		mimeType = "application/pdf"
	case "detailed-csv":
		data, filename, err = g.reports.GenerateDetailedCSV(ctx, orgID, filter, includeFinancial)
		mimeType = "text/csv"
	default:
		return nil, fmt.Errorf("unknown export report type %q", reportType)
	}
	if err != nil {
		return nil, err
	}
	return []scheduler.ExportReportFile{{Filename: filename, MimeType: mimeType, Data: data}}, nil
}

// generateInvoicesPDFs attaches every invoice already generated for orgID
// whose selected period falls within [startDate, endDate], narrowed to
// clientIDs (empty means every client) - never a new one, per
// ai-instruct-79: the job just forwards what invoicing already produced,
// the same way a human would attach existing invoices to an email. Unlike
// the other report types, projectIDs plays no part here: an invoice isn't
// scoped to individual projects.
func (g *ExportReportGenerator) generateInvoicesPDFs(ctx context.Context, orgID string, clientIDs []string, startDate, endDate string) ([]scheduler.ExportReportFile, error) {
	invoices, err := g.invoices.List(ctx, orgID, clientIDs, startDate, endDate)
	if err != nil {
		return nil, err
	}

	files := make([]scheduler.ExportReportFile, 0, len(invoices))
	for _, inv := range invoices {
		pdf, number, err := g.invoices.GetPDF(ctx, inv.ID)
		if err != nil {
			return nil, err
		}
		files = append(files, scheduler.ExportReportFile{
			Filename: number + ".pdf",
			MimeType: "application/pdf",
			Data:     pdf,
		})
	}
	return files, nil
}
