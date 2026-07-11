package report

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
	"github.com/mandolyte/mdtopdf"

	"cwclock-api/internal/models"
)

const (
	logoMaxHeight   = 36.0  // pt
	logoMargin      = 24.0  // pt, from the page's right/top edge
	chartMaxWidthPt = 500.0 // pt, capped so a wide chart doesn't dwarf the table below it
	stampMaxWidthPt = 120.0 // pt, invoice stamp width cap

	donutDisplaySize     = 130.0 // pt, ring's displayed width/height
	donutLegendGap       = 10.0  // pt, gap between the ring and its legend
	donutPanelWidth      = 170.0 // pt, ring/legend column width
	donutGutterWidth     = 190.0 // pt, table's reserved right margin (panel + breathing room)
	donutSwatchSize      = 8.0   // pt, legend color-swatch square
	donutLegendFont      = 8.0   // pt, project name line
	donutLegendMetaFont  = 7.0   // pt, duration/percentage line
	donutLegendLine      = 11.0  // pt, name line height
	donutLegendMetaLine  = 9.0   // pt, meta line height
	donutLegendRowHeight = donutLegendLine + donutLegendMetaLine

	footerText = "Generated with cwcloud.me"
	footerURL  = "https://www.cwcloud.me"
)

// newRenderer builds the shared mdtopdf renderer both RenderMarkdownPDF and
// RenderReportTablePDF start from.
//
// Landscape, with a smaller table font than the library's default: report
// tables have too many columns to fit readably in portrait at 12pt, and
// mdtopdf makes no attempt to auto-fit column widths (a documented
// limitation), so this is the practical way to keep cells legible.
//
// cp1252 covers accented Latin characters (French included, the other
// language this app supports); without it, any non-ASCII rune is written
// as raw UTF-8 bytes into a cp1252-encoded font and renders as mojibake.
func newRenderer() *mdtopdf.PdfRenderer {
	opts := []mdtopdf.RenderOption{mdtopdf.WithUnicodeTranslator("cp1252")}
	renderer := mdtopdf.NewPdfRenderer("L", "A4", "", "", opts, mdtopdf.LIGHT)
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
		pdf.CellFormat(0, 5, footerText, "", 0, "C", false, 0, footerURL)
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
	renderer := newRenderer()
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
// optional chart image (pass nil chartPNG to omit it, placed between the
// header and the table), then the table itself alongside an optional
// project donut chart + legend (pass nil donutPNG to omit it, placed to the
// table's right - see placeDonut/placeProjectLegend). Unlike
// RenderMarkdownPDF, the table is drawn directly with fpdf (see drawTable)
// rather than through a markdown table: mdtopdf sizes columns from the
// header cell alone and can't wrap body text, so a long value would
// overflow its column and get clipped by the next one's background fill
// instead of wrapping.
func RenderReportTablePDF(headerMarkdown string, chartPNG []byte, donutPNG []byte, projectDurations []models.ReportProjectDuration, columns []tableColumn, rows [][]string, logoData []byte, logoType string) ([]byte, error) {
	renderer := newRenderer()
	addFooter(renderer.Pdf)

	if len(logoData) > 0 {
		placeLogo(renderer.Pdf, logoData, logoType)
	}

	if err := renderer.Run([]byte(headerMarkdown)); err != nil {
		return nil, err
	}

	if len(chartPNG) > 0 {
		placeChart(renderer.Pdf, chartPNG)
	}

	translate := renderer.Pdf.UnicodeTranslatorFromDescriptor("cp1252")

	rightGutter := 0.0
	if len(donutPNG) > 0 {
		rightGutter = donutGutterWidth
		placeDonutPanel(renderer.Pdf, translate, donutPNG, projectDurations)
	}

	drawTable(renderer.Pdf, translate, columns, rows, rightGutter)

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
	pdf.ImageOptions("report-header-logo", pageWidth-logoMargin-width, logoMargin/2, width, logoMaxHeight, false, options, 0, "")
}

// placeChart embeds the daily chart PNG below the current cursor position,
// scaled to the content width (capped at chartMaxWidthPt) and left-aligned
// on the page's left margin. flow=true advances the cursor past the image's
// height so the table markdown run afterward starts writing below it.
func placeChart(pdf *fpdf.Fpdf, data []byte) {
	options := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	info := pdf.RegisterImageOptionsReader("report-daily-chart", options, bytes.NewReader(data))
	if info == nil || info.Width() <= 0 || info.Height() <= 0 {
		return
	}

	left, _, right, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()
	width := pageWidth - left - right
	if width > chartMaxWidthPt {
		width = chartMaxWidthPt
	}
	height := width * (info.Height() / info.Width())

	pdf.Ln(6)
	pdf.ImageOptions("report-daily-chart", left, pdf.GetY(), width, height, true, options, 0, "")
	pdf.Ln(6)
}

// placeDonutPanel draws the summary report's project donut ring plus a
// color-swatch legend (project name, duration, share of total) in the
// page's right-hand gutter (see donutGutterWidth), anchored at the table's
// starting Y. Drawn once - like placeChart's daily bar chart above it, it
// isn't repeated if the table paginates.
func placeDonutPanel(pdf *fpdf.Fpdf, translate func(string) string, data []byte, rows []models.ReportProjectDuration) {
	_, _, right, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()
	x := pageWidth - right - donutPanelWidth
	y := pdf.GetY()

	ringHeight := placeDonut(pdf, data, x+(donutPanelWidth-donutDisplaySize)/2, y, donutDisplaySize)
	placeProjectLegend(pdf, translate, x, y+ringHeight+donutLegendGap, donutPanelWidth, rows)
}

// placeDonut embeds a square chart image (see ProjectDonutPNG) at an
// explicit (x, y) position, scaled to size×size, and returns the drawn
// height so callers can stack content below it (see placeDonutPanel).
// Unlike placeChart, this doesn't move the page cursor - the donut sits
// beside the table, not inline with it.
func placeDonut(pdf *fpdf.Fpdf, data []byte, x, y, size float64) float64 {
	options := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	info := pdf.RegisterImageOptionsReader("report-project-donut", options, bytes.NewReader(data))
	if info == nil || info.Width() <= 0 || info.Height() <= 0 {
		return 0
	}
	pdf.ImageOptions("report-project-donut", x, y, size, size, false, options, 0, "")
	return size
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
		pdf.CellFormat(textWidth, donutLegendLine, name, "", 0, "L", false, 0, "")

		pct := int(float64(r.DurationSecs)/float64(total)*100 + 0.5)
		meta := fmt.Sprintf("%s - %d%%", formatHMS(r.DurationSecs), pct)
		pdf.SetFont("", "", donutLegendMetaFont)
		pdf.SetTextColor(120, 120, 120)
		pdf.SetXY(textX, rowY+donutLegendLine)
		pdf.CellFormat(textWidth, donutLegendMetaLine, meta, "", 0, "L", false, 0, "")
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
	pdf.ImageOptions("invoice-stamp", pageWidth-right-width, pdf.GetY(), width, height, true, options, 0, "")
}
