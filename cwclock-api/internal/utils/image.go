package utils

import (
	"encoding/base64"
	"net/http"
	"strings"
)

// dataURIBase64Marker is the ";base64," separator between a data URI's
// media type and its payload, e.g. "data:image/png;base64,iVBOR...".
const dataURIBase64Marker = ";base64,"

// DecodeImage decodes payload into raw image bytes plus its sniffed mime
// type, accepting either a well-formed "data:image/...;base64,..." URI or a
// bare base64 string with no prefix at all (organization avatars are
// stored in the database the second way - just the base64 payload, with no
// mime metadata). The mime type is always sniffed from the decoded bytes
// rather than trusting any declared one, since mislabeling e.g. a JPEG as
// PNG makes most renderers refuse to display it. ok is false when payload
// is blank, isn't valid base64, or doesn't sniff as one of the raster
// types this app treats as an image (PNG/JPEG/GIF/WEBP).
func DecodeImage(payload string) (data []byte, mimeType string, ok bool) {
	if IsBlank(payload) {
		return nil, EMPTY, false
	}
	if strings.HasPrefix(payload, "data:") {
		if idx := strings.Index(payload, dataURIBase64Marker); idx >= 0 {
			payload = payload[idx+len(dataURIBase64Marker):]
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, EMPTY, false
	}
	switch mt := http.DetectContentType(decoded); mt {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return decoded, mt, true
	default:
		return nil, EMPTY, false
	}
}
