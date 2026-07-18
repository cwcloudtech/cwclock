package report

import (
	"encoding/base64"
	"strings"

	"cwclock-api/internal/assets"
	"cwclock-api/internal/utils"
)

// ResolveLogo returns the image bytes and fpdf image type ("PNG"/"JPG"/"GIF")
// to place in a report's header: the organization's own avatar when it's in
// a format fpdf can render directly, otherwise the bundled CWClock logo.
//
// An SVG avatar (or anything else undecodable) falls back to the default
// logo rather than being converted: mdtopdf's SVG path shells out to a
// headless-Chrome-based rasterizer, which this app avoids entirely.
func ResolveLogo(orgPicture string) (data []byte, imgType string) {
	if decoded, dt, ok := decodeDataURI(orgPicture); ok {
		return decoded, dt
	}
	return assets.CWClockLogoPNG, "PNG"
}

// decodeDataURI parses a "data:image/png;base64,...." URI (as produced by
// this app's avatar upload) into raw bytes plus an fpdf-compatible image
// type. ok is false for anything not in that exact shape, or not a raster
// type fpdf decodes natively (PNG/JPEG/GIF).
func decodeDataURI(dataURI string) (data []byte, imgType string, ok bool) {
	const prefix = "data:"
	if !strings.HasPrefix(dataURI, prefix) {
		return nil, utils.EMPTY, false
	}
	comma := strings.IndexByte(dataURI, ',')
	if comma < 0 {
		return nil, utils.EMPTY, false
	}
	meta := dataURI[len(prefix):comma]
	if !strings.Contains(meta, "base64") {
		return nil, utils.EMPTY, false
	}
	mime, _, _ := strings.Cut(meta, ";")
	imgType = strings.ToUpper(strings.TrimPrefix(mime, "image/"))
	if imgType == "JPEG" {
		imgType = "JPG"
	}
	if imgType != "PNG" && imgType != "JPG" && imgType != "GIF" {
		return nil, utils.EMPTY, false
	}

	decoded, err := base64.StdEncoding.DecodeString(dataURI[comma+1:])
	if err != nil {
		return nil, utils.EMPTY, false
	}
	return decoded, imgType, true
}
