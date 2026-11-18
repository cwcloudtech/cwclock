package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

// cidHeader is the correlation-id header, matching sf-api / uprodit (X-Cid): read
// from the inbound request or generated, put on the context, echoed on the
// response, and attached to every span + access-log line (see Instrument).
const cidHeader = "X-Cid"

const cidKey contextKey = "cid"

// Correlation settles the correlation id before anything logs it: every request
// gets an X-Cid (the caller's, or a fresh one), stored on the context and echoed
// back on the response header. It is registered outermost, ahead of Instrument.
func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := strings.TrimSpace(r.Header.Get(cidHeader))
		if cid == "" {
			cid = newCID()
		}
		w.Header().Set(cidHeader, cid)
		ctx := context.WithValue(r.Context(), cidKey, cid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CIDFromContext returns the request's correlation id, or "" when absent.
func CIDFromContext(ctx context.Context) string {
	cid, _ := ctx.Value(cidKey).(string)
	return cid
}

func newCID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
