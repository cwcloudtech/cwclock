package report

import (
	"bytes"

	"github.com/mandolyte/mdtopdf"
)

// RenderMarkdownPDF converts markdown content into PDF bytes. It has no
// report-specific knowledge on purpose: it's the generic building block the
// invoicing feature is meant to reuse.
//
// Note: markdown image syntax is deliberately not fed into this renderer.
// mdtopdf shells out to a headless-Chrome-based SVG rasterizer for any SVG
// image it encounters, which this app avoids entirely, so report/invoice
// markdown should stick to text and tables.
func RenderMarkdownPDF(markdown string) ([]byte, error) {
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
	if err := renderer.Run([]byte(markdown)); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := renderer.Pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
