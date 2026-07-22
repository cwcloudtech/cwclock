package report

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
	"github.com/mandolyte/mdtopdf"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

const (
	logoMaxHeight   = 36.0  // pt
	logoMargin      = 24.0  // pt, from the page's right/top edge
	chartMaxWidthPt = 500.0 // pt, capped so a wide chart doesn't dwarf the table below it
	stampMaxWidthPt = 120.0 // pt, invoice stamp width cap

	donutDisplaySize     = 130.0 // pt, ring's displayed width/height
	donutLegendGap       = 10.0  // pt, gap between the ring and its legend
	donutColumnWidth     = 170.0 // pt, ring/legend column width
	donutSwatchSize      = 8.0   // pt, legend color-swatch square
	donutLegendFont      = 8.0   // pt, project name line
	donutLegendMetaFont  = 7.0   // pt, duration/percentage line
	donutLegendLine      = 11.0  // pt, name line height
	donutLegendMetaLine  = 9.0   // pt, meta line height
	donutLegendRowHeight = donutLegendLine + donutLegendMetaLine

	footerText = "Generated with cwcloud.me"
	footerURL  = "https://www.cwcloud.me"
)

// newPdfRenderer builds the shared mdtopdf renderer report and invoice PDFs
// all start from, in the given A4 orientation ("L" landscape, "P" portrait).
//
// A smaller table font than the library's default: mdtopdf makes no attempt
// to auto-fit column widths (a documented limitation), so this is the
// practical way to keep cells legible - drawTable's own width-driven
// wrapping is what actually keeps a wide report table's cells readable in
// RenderReportTablePDF's fixed A4 portrait (see its doc comment).
//
// cp1252 covers accented Latin characters (French included, the other
// language this app supports); without it, any non-ASCII rune is written
// as raw UTF-8 bytes into a cp1252-encoded font and renders as mojibake.
func newPdfRenderer(orientation string) *mdtopdf.PdfRenderer {
	opts := []mdtopdf.RenderOption{mdtopdf.WithUnicodeTranslator("cp1252")}
	renderer := mdtopdf.NewPdfRenderer(orientation, "A4", utils.EMPTY, utils.EMPTY, opts, mdtopdf.LIGHT)
	renderer.THeader.Size = 9
	renderer.TBody.Size = 9
	return renderer
}

// addFooter registers a footer drawing function on the fpdf instance that
// prints a small, italic, light-grey attribution line at the bottom of every
// page. Must be called before the first page is closed (i.e. before
// outputPDF).
func addFooter(pdf *fpdf.Fpdf) {
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "I", 7)
		pdf.SetTextColor(170, 170, 170)
		pdf.CellFormat(0, 5, footerText, utils.EMPTY, 0, "C", false, 0, footerURL)
	})
}

func outputPDF(pdf *fpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderMarkdownPDF converts markdown content into PDF bytes, optionally
// placing a logo in the header's top-right corner (pass nil logoData to
// omit it). Past the logo, it has no report-specific knowledge on purpose:
// it's the generic building block simple (tableless) markdown documents are
// rendered through.
//
// Note: markdown image syntax is deliberately not fed into this renderer.
// mdtopdf shells out to a headless-Chrome-based SVG rasterizer for any SVG
// image it encounters, which this app avoids entirely, so markdown fed to
// it should stick to text — tables go through drawTable instead (see
// RenderReportTablePDF and RenderInvoicePDF), since mdtopdf's own table
// rendering can neither wrap long cell text nor tolerate a blank header
// cell (see drawTable's doc comment).
func RenderMarkdownPDF(markdown string, logoData []byte, logoType string) ([]byte, error) {
	renderer := newPdfRenderer("L")
	addFooter(renderer.Pdf)

	// Placed before Run() writes any markdown: fpdf can't add content to an
	// earlier page once a later one exists, so the logo has to land on page
	// 1 while it's still the current page. flow=false keeps it from moving
	// the cursor markdown's own header text starts writing from.
	if len(logoData) > 0 {
		placeLogo(renderer.Pdf, logoData, logoType)
	}

	if err := renderer.Run([]byte(markdown)); err != nil {
		return nil, err
	}

	return outputPDF(renderer.Pdf)
}

// RenderReportTablePDF renders a report PDF: the header markdown, an
// optional charts row (pass nil chartPNG/donutPNG to omit either - see
// placeChartsRow), then the table itself. Unlike RenderMarkdownPDF, the
// table is drawn directly with fpdf (see drawTable) rather than through a
// markdown table: mdtopdf sizes columns from the header cell alone and
// can't wrap body text, so a long value would overflow its column and get
// clipped by the next one's background fill instead of wrapping. Always A4
// portrait (like invoices) - drawTable sizes columns from the actual usable
// page width, so it just wraps cell text onto more lines than landscape
// would rather than clipping it.
func RenderReportTablePDF(headerMarkdown string, chartPNG []byte, donutPNG []byte, projectDurations []models.ReportProjectDuration, columns []tableColumn, rows [][]string, logoData []byte, logoType string) ([]byte, error) {
	renderer := newPdfRenderer("P")
	addFooter(renderer.Pdf)

	if len(logoData) > 0 {
		placeLogo(renderer.Pdf, logoData, logoType)
	}

	if err := renderer.Run([]byte(headerMarkdown)); err != nil {
		return nil, err
	}

	translate := renderer.Pdf.UnicodeTranslatorFromDescriptor("cp1252")
	placeChartsRow(renderer.Pdf, translate, chartPNG, donutPNG, projectDurations)
	drawTable(renderer.Pdf, translate, columns, rows)

	return outputPDF(renderer.Pdf)
}

func placeLogo(pdf *fpdf.Fpdf, data []byte, imgType string) {
	options := fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
	info := pdf.RegisterImageOptionsReader("report-header-logo", options, bytes.NewReader(data))
	if info == nil || info.Height() <= 0 {
		return
	}

	width := info.Width() * (logoMaxHeight / info.Height())
	pageWidth, _ := pdf.GetPageSize()
	pdf.ImageOptions("report-header-logo", pageWidth-logoMargin-width, logoMargin/2, width, logoMaxHeight, false, options, 0, utils.EMPTY)
}

// placeImage embeds a PNG at an explicit (x, y), scaled to width with its
// own aspect ratio preserved, and returns the drawn height. Doesn't move the
// page cursor - callers position the next thing explicitly (see
// placeChartsRow).
func placeImage(pdf *fpdf.Fpdf, key string, data []byte, x, y, width float64) float64 {
	options := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	info := pdf.RegisterImageOptionsReader(key, options, bytes.NewReader(data))
	if info == nil || info.Width() <= 0 || info.Height() <= 0 {
		return 0
	}
	height := width * (info.Height() / info.Width())
	pdf.ImageOptions(key, x, y, width, height, false, options, 0, utils.EMPTY)
	return height
}

// placeChartsRow embeds the summary report's two charts side by side, below
// the header and above the table: the project donut ring + legend (see
// ProjectDonutPNG/placeProjectLegend) on the left, the daily bar chart (see
// DailyChartPNG) filling the remaining width to its right - the same
// left/right order cwclock-ui's SummaryReportView uses for the same two
// charts. Either image may be empty (nothing to plot); if both are, nothing
// is drawn and the cursor is left untouched.
func placeChartsRow(pdf *fpdf.Fpdf, translate func(string) string, chartPNG []byte, donutPNG []byte, projectDurations []models.ReportProjectDuration) {
	if len(chartPNG) == 0 && len(donutPNG) == 0 {
		return
	}

	left, _, right, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()
	usableWidth := pageWidth - left - right

	pdf.Ln(6)
	y := pdf.GetY()
	rowBottom := y
	chartX := left

	if len(donutPNG) > 0 {
		ringX := left + (donutColumnWidth-donutDisplaySize)/2
		ringHeight := placeImage(pdf, "report-project-donut", donutPNG, ringX, y, donutDisplaySize)
		legendY := y + ringHeight + donutLegendGap
		placeProjectLegend(pdf, translate, left, legendY, donutColumnWidth, projectDurations)
		rowBottom = max(rowBottom, legendY+float64(len(projectDurations))*donutLegendRowHeight)
		chartX = left + donutColumnWidth + donutLegendGap
	}

	if len(chartPNG) > 0 {
		width := usableWidth - (chartX - left)
		if width > chartMaxWidthPt {
			width = chartMaxWidthPt
		}
		height := placeImage(pdf, "report-daily-chart", chartPNG, chartX, y, width)
		rowBottom = max(rowBottom, y+height)
	}

	pdf.SetXY(left, rowBottom)
	pdf.Ln(6)
}

// placeProjectLegend draws one row per project below the donut ring: a
// color swatch matching its ring slice, its name (truncated with an
// ellipsis if it doesn't fit width), and a muted second line with its
// duration and share of the total. Doesn't move the page cursor.
func placeProjectLegend(pdf *fpdf.Fpdf, translate func(string) string, x, y, width float64, rows []models.ReportProjectDuration) {
	total := 0
	for _, r := range rows {
		total += r.DurationSecs
	}
	if total == 0 {
		return
	}

	textX := x + donutSwatchSize + 6
	textWidth := width - donutSwatchSize - 6
	rowY := y

	for _, r := range rows {
		c := parseHexColor(r.Color)
		pdf.SetFillColor(int(c.R), int(c.G), int(c.B))
		pdf.Rect(x, rowY+2, donutSwatchSize, donutSwatchSize, "F")

		pdf.SetFont("", "B", donutLegendFont)
		name := truncateToWidth(pdf, translate(r.ProjectName), textWidth)
		pdf.SetXY(textX, rowY)
		pdf.CellFormat(textWidth, donutLegendLine, name, utils.EMPTY, 0, "L", false, 0, utils.EMPTY)

		pct := int(float64(r.DurationSecs)/float64(total)*100 + 0.5)
		meta := fmt.Sprintf("%s - %d%%", formatHMS(r.DurationSecs), pct)
		pdf.SetFont(utils.EMPTY, utils.EMPTY, donutLegendMetaFont)
		pdf.SetTextColor(120, 120, 120)
		pdf.SetXY(textX, rowY+donutLegendLine)
		pdf.CellFormat(textWidth, donutLegendMetaLine, meta, utils.EMPTY, 0, "L", false, 0, utils.EMPTY)
		pdf.SetTextColor(0, 0, 0)

		rowY += donutLegendRowHeight
	}
}

// truncateToWidth shortens s with a trailing ellipsis until it fits width
// at the pdf's currently selected font, leaving it untouched if it already
// fits.
func truncateToWidth(pdf *fpdf.Fpdf, s string, width float64) string {
	if pdf.GetStringWidth(s) <= width {
		return s
	}
	runes := []rune(s)
	for len(runes) > 1 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "..."
		if pdf.GetStringWidth(candidate) <= width {
			return candidate
		}
	}
	return string(runes)
}

// placeStamp embeds an organization's stamp image below the current cursor
// position (i.e. below the invoice content already written), scaled to
// stampMaxWidthPt and right-aligned within the page margins.
func placeStamp(pdf *fpdf.Fpdf, data []byte, imgType string) {
	options := fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
	info := pdf.RegisterImageOptionsReader("invoice-stamp", options, bytes.NewReader(data))
	if info == nil || info.Width() <= 0 || info.Height() <= 0 {
		return
	}

	width := stampMaxWidthPt
	height := width * (info.Height() / info.Width())
	_, _, right, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()

	pdf.Ln(10)
	pdf.ImageOptions("invoice-stamp", pageWidth-right-width, pdf.GetY(), width, height, true, options, 0, utils.EMPTY)
}
