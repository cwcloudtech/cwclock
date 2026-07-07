package report

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"cwclock-api/internal/models"
)

// Pixel dimensions of the rendered chart; kept wide/short to match the
// letterbox shape of the frontend's daily bar chart, and large enough that
// it stays sharp once scaled down to the PDF page width.
const (
	chartWidth     = 1200
	chartHeight    = 360
	chartMargin    = 16
	chartBarGap    = 4
	chartLabelArea = 24 // px reserved below the baseline for day-of-month labels
	chartFontScale = 2  // pixel scale applied to the 3×5 bitmap font
)

// font3x5 maps digit runes to their 3×5 row bitmasks.
// Bit layout per row byte: bit 2 = leftmost column, bit 0 = rightmost.
var font3x5 = map[rune][5]byte{
	'0': {7, 5, 5, 5, 7},
	'1': {6, 2, 2, 2, 7},
	'2': {7, 1, 3, 4, 7},
	'3': {7, 1, 3, 1, 7},
	'4': {5, 5, 7, 1, 1},
	'5': {7, 4, 7, 1, 7},
	'6': {3, 4, 7, 5, 7},
	'7': {7, 1, 2, 2, 2},
	'8': {7, 5, 7, 5, 7},
	'9': {7, 5, 7, 1, 3},
}

// drawLabel renders label centered at horizontal pixel cx, with the top of the
// glyphs at y, using font3x5 at the given pixel scale s.
func drawLabel(img *image.RGBA, label string, cx, y, s int, c color.RGBA) {
	charW := 3 * s
	gap := s
	total := len(label)*charW + (len(label)-1)*gap
	x := cx - total/2
	for _, r := range label {
		bm, ok := font3x5[r]
		if !ok {
			x += charW + gap
			continue
		}
		for row, mask := range bm {
			for col := 0; col < 3; col++ {
				if (mask>>uint(2-col))&1 != 0 {
					for dy := 0; dy < s; dy++ {
						for dx := 0; dx < s; dx++ {
							img.Set(x+col*s+dx, y+row*s+dy, c)
						}
					}
				}
			}
		}
		x += charW + gap
	}
}

// DailyChartPNG rasterizes the summary report's per-day duration buckets as
// a bar chart PNG, mirroring cwclock-ui's DailyBarChart, so the PDF export
// can show the same at-a-glance shape as an image (mdtopdf's markdown image
// support is avoided elsewhere in this package - see RenderMarkdownPDF -
// so the chart is embedded directly via fpdf like the org logo). Returns
// nil when there's nothing to plot.
func DailyChartPNG(daily []models.ReportDailyBucket) []byte {
	if len(daily) == 0 {
		return nil
	}

	maxSecs := 0
	for _, d := range daily {
		if d.DurationSecs > maxSecs {
			maxSecs = d.DurationSecs
		}
	}
	if maxSecs == 0 {
		return nil
	}

	img := image.NewRGBA(image.Rect(0, 0, chartWidth, chartHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)

	axisColor := color.RGBA{209, 213, 219, 255}  // matches --cw-border
	barColor := color.RGBA{28, 185, 247, 255}    // matches --cw-primary (#1cb9f7)
	labelColor := color.RGBA{107, 114, 128, 255} // gray-500
	baselineY := chartHeight - chartLabelArea
	plotWidth := chartWidth - 2*chartMargin
	plotHeight := chartHeight - chartMargin - chartLabelArea
	barWidth := plotWidth/len(daily) - chartBarGap
	if barWidth < 1 {
		barWidth = 1
	}

	for x := chartMargin; x < chartWidth-chartMargin; x++ {
		img.Set(x, baselineY, axisColor)
	}

	for i, d := range daily {
		x0 := chartMargin + i*(barWidth+chartBarGap)
		barCx := x0 + barWidth/2

		barHeight := int(float64(d.DurationSecs) / float64(maxSecs) * float64(plotHeight))
		if barHeight >= 1 {
			rect := image.Rect(x0, baselineY-barHeight, x0+barWidth, baselineY)
			draw.Draw(img, rect, &image.Uniform{barColor}, image.Point{}, draw.Src)
		}

		// Day-of-month label ("DD") centered below the bar.
		label := d.Day
		if len(label) >= 10 {
			label = label[8:10]
		}
		drawLabel(img, label, barCx, baselineY+4, chartFontScale, labelColor)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}
