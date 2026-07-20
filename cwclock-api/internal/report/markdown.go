package report

import (
	"strings"
	"text/template"

	"cwclock-api/internal/models"
	"cwclock-api/internal/templates"
	"cwclock-api/internal/utils"
)

// reportHeader is the data both report markdown templates share: title,
// org name, period and totals line.
type reportHeader struct {
	Title         string
	OrgName       string
	Period        string
	TotalDuration string
	TotalDays     string
	ShowAmount    bool
	TotalAmount   string
	Currency      string
}

func newReportHeader(title, orgName string, start, end string, totals models.ReportTotals) reportHeader {
	h := reportHeader{
		Title:         title,
		OrgName:       orgName,
		Period:        formatUSDate(start) + " - " + formatUSDate(end),
		TotalDuration: formatHMS(totals.DurationSecs),
		TotalDays:     formatDays(totals.Days),
		Currency:      totals.Currency,
	}
	if totals.Amount != nil {
		h.ShowAmount = true
		h.TotalAmount = formatAmount(*totals.Amount)
	}
	return h
}

var headerTemplate = template.Must(template.New("header").Parse(templates.HeaderMarkdown))

// SummaryPDF renders the summary report as a PDF: a header with totals, the
// daily duration chart as an image, then one table row per (project,
// description) group. logoData/logoType (see ResolveLogo) are placed in the
// header's top-right corner.
func SummaryPDF(orgName, start, end string, report models.SummaryReport, logoData []byte, logoType string) ([]byte, error) {
	var headerBuf strings.Builder
	if err := headerTemplate.Execute(&headerBuf, newReportHeader("Summary report", orgName, start, end, report.Totals)); err != nil {
		return nil, err
	}

	showAmount := report.Totals.Amount != nil
	columns := []tableColumn{
		{Header: "Project", Weight: 20},
		{Header: "Client", Weight: 16},
		{Header: "Description", Weight: 36},
		{Header: "Email", Weight: 26},
		{Header: "Time (h)", Weight: 12},
		{Header: "Time (decimal)", Weight: 12},
	}
	if showAmount {
		columns = append(columns, tableColumn{Header: "Amount (" + report.Totals.Currency + ")", Weight: 14})
	}

	rows := make([][]string, 0, len(report.Rows))
	for _, r := range report.Rows {
		row := []string{
			r.ProjectName, r.ClientName, r.Description, r.UserEmail,
			formatHMS(r.DurationSecs), formatDecimalHours(r.DurationSecs),
		}
		if showAmount {
			amt := 0.0
			if r.Amount != nil {
				amt = *r.Amount
			}
			row = append(row, formatAmount(amt))
		}
		rows = append(rows, row)
	}

	chartPNG := DailyChartPNG(report.Daily)
	donutPNG := ProjectDonutPNG(report.ProjectDurations)
	return RenderReportTablePDF(headerBuf.String(), chartPNG, donutPNG, report.ProjectDurations, columns, rows, logoData, logoType)
}

// DetailedPDF renders the detailed report as a PDF: a header with totals
// followed by one table row per time entry. logoData/logoType (see
// ResolveLogo) are placed in the header's top-right corner.
func DetailedPDF(orgName, start, end string, report models.DetailedReport, logoData []byte, logoType string) ([]byte, error) {
	var headerBuf strings.Builder
	if err := headerTemplate.Execute(&headerBuf, newReportHeader("Detailed report", orgName, start, end, report.Totals)); err != nil {
		return nil, err
	}

	showAmount := report.Totals.Amount != nil
	columns := []tableColumn{
		{Header: "Date", Weight: 10},
		{Header: "Description", Weight: 26},
		{Header: "Project / Client", Weight: 20},
		{Header: "Time", Weight: 16},
		{Header: "Duration", Weight: 10},
		{Header: "User", Weight: 14},
		{Header: "Email", Weight: 18},
	}
	if showAmount {
		columns = append(columns, tableColumn{Header: "Amount", Weight: 10})
	}

	rows := make([][]string, 0, len(report.Entries))
	for _, e := range report.Entries {
		// The PDF's Time column shows "All day" rather than the synthetic
		// 9:00-to-9:00-plus-HoursPerDay window allDayWindow assigns an
		// all-day entry's Start/End for duration/sorting purposes - that
		// window is a computation detail, not a real clocked time range, and
		// printing it as one reads as an actual, precise shift. The CSV
		// export keeps showing that window as literal Start/End columns
		// (ai-instruct-66), since its Start Time/End Time columns are
		// expected to always hold a time, matching Clockify-style importers.
		timeRange := utils.If(e.AllDay, "All day", formatAMPM(e.Start)+" - "+formatAMPM(e.End))

		row := []string{
			formatUSDate(e.Day),
			e.Text,
			e.ProjectName + " - " + e.ClientName,
			timeRange,
			formatHMS(e.DurationSecs),
			e.UserName,
			e.UserEmail,
		}
		if showAmount {
			amt := 0.0
			if e.Amount != nil {
				amt = *e.Amount
			}
			row = append(row, formatAmount(amt))
		}
		rows = append(rows, row)
	}

	return RenderReportTablePDF(headerBuf.String(), nil, nil, nil, columns, rows, logoData, logoType)
}
