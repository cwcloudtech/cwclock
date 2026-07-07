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
	chartWidth  = 1200
	chartHeight = 360
	chartMargin = 16
	chartBarGap = 4
)

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
	baselineY := chartHeight - chartMargin

	plotWidth := chartWidth - 2*chartMargin
	plotHeight := chartHeight - 2*chartMargin
	barWidth := plotWidth/len(daily) - chartBarGap
	if barWidth < 1 {
		barWidth = 1
	}

	for x := chartMargin; x < chartWidth-chartMargin; x++ {
		img.Set(x, baselineY, axisColor)
	}

	for i, d := range daily {
		barHeight := int(float64(d.DurationSecs) / float64(maxSecs) * float64(plotHeight))
		if barHeight < 1 {
			continue
		}
		x0 := chartMargin + i*(barWidth+chartBarGap)
		x1 := x0 + barWidth
		y0 := baselineY - barHeight
		rect := image.Rect(x0, y0, x1, baselineY)
		draw.Draw(img, rect, &image.Uniform{barColor}, image.Point{}, draw.Src)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}
