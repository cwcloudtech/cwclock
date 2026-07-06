package report

import (
	"bytes"

	"github.com/go-pdf/fpdf"
	"github.com/mandolyte/mdtopdf"
)

const (
	logoMaxHeight = 36.0 // pt
	logoMargin    = 24.0 // pt, from the page's right/top edge
)

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
	// Landscape, with a smaller table font than the library's default: report
	// tables have too many columns to fit readably in portrait at 12pt, and
	// mdtopdf makes no attempt to auto-fit column widths (a documented
	// limitation), so this is the practical way to keep cells legible.
	//
	// cp1252 covers accented Latin characters (French included, the other
	// language this app supports); without it, any non-ASCII rune is written
	// as raw UTF-8 bytes into a cp1252-encoded font and renders as mojibake.
	opts := []mdtopdf.RenderOption{mdtopdf.WithUnicodeTranslator("cp1252")}
	renderer := mdtopdf.NewPdfRenderer("L", "A4", "", "", opts, mdtopdf.LIGHT)
	renderer.THeader.Size = 9
	renderer.TBody.Size = 9

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
	var buf bytes.Buffer
	if err := renderer.Pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
