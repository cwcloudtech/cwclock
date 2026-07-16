package metrics

import (
	"cwclock-api/internal/utils"
	"regexp"
	"strings"
)

var (
	invalidNameChars = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	leadingDigit     = regexp.MustCompile(`^[0-9]`)
)

// SanitizeMetricName turns arbitrary text (e.g. a user-typed task name) into
// a valid Prometheus/OTEL metric name component: spaces become underscores
// and anything else outside [a-zA-Z0-9_] is stripped, since task names are
// used as part of a dynamic metric name rather than as a label value.
func SanitizeMetricName(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), " ", "_")
	s = invalidNameChars.ReplaceAllString(s, utils.EMPTY)
	s = strings.Trim(s, "_")
	s = strings.ToLower(s)
	if utils.IsBlank(s) {
		return "unnamed"
	}

	if leadingDigit.MatchString(s) {
		s = "_" + s
	}
	return s
}
