package report

import (
	"strings"
	"text/template"

	"cwclock-api/internal/models"
)

// reportHeader is the data both report markdown templates share: title,
// org name, period and totals line.
type reportHeader struct {
	Title         string
	OrgName       string
	Period        string
	TotalDuration string
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
		Currency:      totals.Currency,
	}
	if totals.Amount != nil {
		h.ShowAmount = true
		h.TotalAmount = formatAmount(*totals.Amount)
	}
	return h
}

const headerMarkdownTpl = `# {{.OrgName}} - {{.Title}}

{{.Period}}

**Total:** {{.TotalDuration}}   **Billable:** {{.TotalDuration}}{{if .ShowAmount}}   **Amount:** {{.TotalAmount}} {{.Currency}}{{end}}
`

// nbsp is used (instead of a plain space) to pad table headers up to a
// minimum width: mdtopdf sizes each column from its header cell alone (not
// the widest body cell) and the markdown table parser trims plain trailing
// spaces from cell text, but leaves non-breaking spaces alone.
const nbsp = " "

// padHeader widens a header label to approximately `width` runes so its
// column has enough room for typical body content.
func padHeader(label string, width int) string {
	pad := width - len([]rune(label))
	if pad <= 0 {
		return label
	}
	return label + strings.Repeat(nbsp, pad)
}

type summaryRowVM struct {
	ProjectName, ClientName, Description string
	DurationHMS, DurationDecimal         string
	ShowAmount                           bool
	AmountStr                            string
}

const summaryTableMarkdownTpl = `
| {{.ProjectHeader}} | {{.ClientHeader}} | {{.DescriptionHeader}} | Time (h) | Time (decimal) |{{if .ShowAmount}} Amount ({{.Currency}}) |{{end}}
|---|---|---|---|---|{{if .ShowAmount}}---|{{end}}
{{range .Rows}}| {{.ProjectName}} | {{.ClientName}} | {{.Description}} | {{.DurationHMS}} | {{.DurationDecimal}} |{{if $.ShowAmount}} {{.AmountStr}} |{{end}}
{{end}}`

type detailedRowVM struct {
	Day, Description, ProjectClient, Time, DurationHMS, User string
	ShowAmount                                               bool
	AmountStr                                                string
}

const detailedTableMarkdownTpl = `
| {{.DateHeader}} | {{.DescriptionHeader}} | {{.ProjectClientHeader}} | {{.TimeHeader}} | Duration | {{.UserHeader}} |{{if .ShowAmount}} Amount |{{end}}
|---|---|---|---|---|---|{{if .ShowAmount}}---|{{end}}
{{range .Rows}}| {{.Day}} | {{.Description}} | {{.ProjectClient}} | {{.Time}} | {{.DurationHMS}} | {{.User}} |{{if $.ShowAmount}} {{.AmountStr}} |{{end}}
{{end}}`

var headerTemplate = template.Must(template.New("header").Parse(headerMarkdownTpl))

type summaryTableData struct {
	ShowAmount                                     bool
	Currency                                       string
	ProjectHeader, ClientHeader, DescriptionHeader string
	Rows                                           []summaryRowVM
}

var summaryTableTemplate = template.Must(template.New("summaryTable").Parse(summaryTableMarkdownTpl))

type detailedTableData struct {
	ShowAmount                                                                 bool
	DateHeader, DescriptionHeader, ProjectClientHeader, TimeHeader, UserHeader string
	Rows                                                                       []detailedRowVM
}

var detailedTableTemplate = template.Must(template.New("detailedTable").Parse(detailedTableMarkdownTpl))

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
	rows := make([]summaryRowVM, 0, len(report.Rows))
	for _, r := range report.Rows {
		vm := summaryRowVM{
			ProjectName:     r.ProjectName,
			ClientName:      r.ClientName,
			Description:     r.Description,
			DurationHMS:     formatHMS(r.DurationSecs),
			DurationDecimal: formatDecimalHours(r.DurationSecs),
			ShowAmount:      showAmount,
		}
		if showAmount && r.Amount != nil {
			vm.AmountStr = formatAmount(*r.Amount)
		}
		rows = append(rows, vm)
	}

	var tableBuf strings.Builder
	if err := summaryTableTemplate.Execute(&tableBuf, summaryTableData{
		ShowAmount:        showAmount,
		Currency:          report.Totals.Currency,
		ProjectHeader:     padHeader("Project", 18),
		ClientHeader:      padHeader("Client", 14),
		DescriptionHeader: padHeader("Description", 42),
		Rows:              rows,
	}); err != nil {
		return nil, err
	}

	chartPNG := DailyChartPNG(report.Daily)
	return RenderSummaryPDF(headerBuf.String(), tableBuf.String(), chartPNG, logoData, logoType)
}

// DetailedPDF renders the detailed report as a PDF: a header with totals
// followed by one table row per time entry. logoData/logoType (see
// ResolveLogo) are placed in the header's top-right corner.
func DetailedPDF(orgName, start, end string, report models.DetailedReport, logoData []byte, logoType string) ([]byte, error) {
	var buf strings.Builder
	if err := headerTemplate.Execute(&buf, newReportHeader("Detailed report", orgName, start, end, report.Totals)); err != nil {
		return nil, err
	}

	showAmount := report.Totals.Amount != nil
	rows := make([]detailedRowVM, 0, len(report.Entries))
	for _, e := range report.Entries {
		timeRange := formatAMPM(e.Start) + " - " + formatAMPM(e.End)
		vm := detailedRowVM{
			Day:           formatUSDate(e.Day),
			Description:   e.Text,
			ProjectClient: e.ProjectName + " - " + e.ClientName,
			Time:          timeRange,
			DurationHMS:   formatHMS(e.DurationSecs),
			User:          e.UserName,
			ShowAmount:    showAmount,
		}
		if showAmount && e.Amount != nil {
			vm.AmountStr = formatAmount(*e.Amount)
		}
		rows = append(rows, vm)
	}

	if err := detailedTableTemplate.Execute(&buf, detailedTableData{
		ShowAmount:          showAmount,
		DateHeader:          padHeader("Date", 12),
		DescriptionHeader:   padHeader("Description", 42),
		ProjectClientHeader: padHeader("Project / Client", 30),
		TimeHeader:          padHeader("Time", 27),
		UserHeader:          padHeader("User", 16),
		Rows:                rows,
	}); err != nil {
		return nil, err
	}

	return RenderMarkdownPDF(buf.String(), logoData, logoType)
}
