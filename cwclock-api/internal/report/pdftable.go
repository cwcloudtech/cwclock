package report

import (
	"cwclock-api/internal/utils"

	"github.com/go-pdf/fpdf"
)

// Font sizing for drawTable's header/body text, matching the 9pt used
// elsewhere in report PDFs (see newRenderer).
const (
	tableFontSizePt   = 9.0
	tableLineHeightPt = 13.0
)

// tableColumn is one column of a report PDF table: its header label and its
// share (weight) of the table's total width, relative to the other columns.
type tableColumn struct {
	Header string
	Weight float64
}

// drawTable renders a bordered table directly with fpdf instead of through
// mdtopdf's markdown tables. mdtopdf sizes each column from its header cell
// alone and has no concept of wrapping (see newRenderer's comment on that
// limitation), so a body cell wider than its column just overflows past its
// border and gets visually clipped by the next column's opaque background
// fill drawn right after it.
//
// This measures each cell's wrapped line count up front (fpdf's own
// SplitLines, so it matches exactly how the text will actually break) so
// every cell in a row shares one row height, then draws every line of every
// cell at an explicit (x, y) position - never relying on the cursor a
// previous cell's drawing left behind - so a wrapped cell can never bleed
// into its neighbors.
func drawTable(pdf *fpdf.Fpdf, translate func(string) string, columns []tableColumn, rows [][]string) {
	left, _, right, bottom := pdf.GetMargins()
	pageWidth, pageHeight := pdf.GetPageSize()
	usableWidth := pageWidth - left - right

	totalWeight := 0.0
	for _, c := range columns {
		totalWeight += c.Weight
	}
	widths := make([]float64, len(columns))
	for i, c := range columns {
		widths[i] = usableWidth * c.Weight / totalWeight
	}

	drawHeader := func() {
		pdf.SetFont("", "B", tableFontSizePt)
		pdf.SetFillColor(28, 185, 247)
		y := pdf.GetY()
		x := left
		for i, c := range columns {
			pdf.SetXY(x, y)
			pdf.CellFormat(widths[i], tableLineHeightPt, translate(c.Header), "1", 0, "C", true, 0, utils.EMPTY)
			x += widths[i]
		}
		pdf.SetXY(left, y+tableLineHeightPt)
		pdf.SetFillColor(255, 255, 255)
		pdf.SetFont(utils.EMPTY, utils.EMPTY, tableFontSizePt)
	}

	drawHeader()

	fill := false
	for _, row := range rows {
		lines := make([][]string, len(row))
		maxLines := 1
		for i, cell := range row {
			raw := pdf.SplitLines([]byte(translate(cell)), widths[i])
			strs := make([]string, len(raw))
			for j, l := range raw {
				strs[j] = string(l)
			}
			if len(strs) == 0 {
				strs = []string{""}
			}
			lines[i] = strs
			if len(strs) > maxLines {
				maxLines = len(strs)
			}
		}

		rowHeight := float64(maxLines) * tableLineHeightPt
		if pdf.GetY()+rowHeight > pageHeight-bottom {
			pdf.AddPage()
			drawHeader()
		}

		y := pdf.GetY()
		x := left
		for i := range columns {
			for j := 0; j < maxLines; j++ {
				line := ""
				if j < len(lines[i]) {
					line = lines[i][j]
				}
				border := "LR"
				switch {
				case maxLines == 1:
					border = "1"
				case j == 0:
					border = "LRT"
				case j == maxLines-1:
					border = "LRB"
				}
				pdf.SetXY(x, y+float64(j)*tableLineHeightPt)
				pdf.CellFormat(widths[i], tableLineHeightPt, line, border, 0, utils.EMPTY, fill, 0, utils.EMPTY)
			}
			x += widths[i]
		}
		pdf.SetXY(left, y+rowHeight)
		fill = !fill
	}
}
