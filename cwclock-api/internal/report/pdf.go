package report

import (
	"bytes"

	"github.com/go-pdf/fpdf"
	"github.com/mandolyte/mdtopdf"
)

const (
	logoMaxHeight   = 36.0  // pt
	logoMargin      = 24.0  // pt, from the page's right/top edge
	chartMaxWidthPt = 500.0 // pt, capped so a wide chart doesn't dwarf the table below it
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
// it's the generic building block the invoicing feature is meant to reuse.
//
// Note: markdown image syntax is deliberately not fed into this renderer.
// mdtopdf shells out to a headless-Chrome-based SVG rasterizer for any SVG
// image it encounters, which this app avoids entirely, so report/invoice
// markdown should stick to text and tables — the logo is placed directly
// via fpdf instead of through markdown.
func RenderMarkdownPDF(markdown string, logoData []byte, logoType string) ([]byte, error) {
	renderer := newRenderer()

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
// header and the table), then the table itself. Unlike RenderMarkdownPDF,
// the table is drawn directly with fpdf (see drawTable) rather than through
// a markdown table: mdtopdf sizes columns from the header cell alone and
// can't wrap body text, so a long value would overflow its column and get
// clipped by the next one's background fill instead of wrapping.
func RenderReportTablePDF(headerMarkdown string, chartPNG []byte, columns []tableColumn, rows [][]string, logoData []byte, logoType string) ([]byte, error) {
	renderer := newRenderer()

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
