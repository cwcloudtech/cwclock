package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var relativeTimePattern = regexp.MustCompile(`^now\(\)((?:[+-]\d+[smhdw])*)$`)
var relativeTimeSegmentPattern = regexp.MustCompile(`([+-])(\d+)([smhdw])`)

// dateOnlyLayout is the bare "YYYY-MM-DD" layout (no time-of-day). Parsing
// it naturally zeroes the time to 00:00:00, which is exactly the default a
// begin date wants; ParseEndTimeExpr detects this layout specifically to
// apply the opposite default (23:59:59) for an end date instead (see
// ai-instruct-101).
const dateOnlyLayout = "2006-01-02"

var absoluteTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04",
	dateOnlyLayout,
}

// ParseTimeExpr resolves a date/time expression to an absolute time. It
// supports plain ISO-8601 strings (a bare "YYYY-MM-DD" defaults to
// 00:00:00) as well as Grafana-style relative expressions built from now()
// and a chain of +/-<amount><unit> segments (units: s, m, h, d, w), e.g.
// now(), now()-1h, now()-1d. Shared by every command accepting a date range
// (record, invoice, ...) for its begin/from value - see ParseEndTimeExpr
// for the end/to counterpart.
func ParseTimeExpr(expr string) (time.Time, error) {
	return parseTimeExpr(expr, false)
}

// ParseEndTimeExpr is ParseTimeExpr's counterpart for an end/to value: a
// bare "YYYY-MM-DD" defaults to the end of that day (23:59:59) instead of
// its start, so "--begin 2024-01-15 --end 2024-01-15" covers the entire
// day rather than a zero-length range (see ai-instruct-101). Every other
// expression (a date/time already carrying a time-of-day, or a now()-style
// expression) is parsed exactly like ParseTimeExpr.
func ParseEndTimeExpr(expr string) (time.Time, error) {
	return parseTimeExpr(expr, true)
}

func parseTimeExpr(expr string, endOfDay bool) (time.Time, error) {
	trimmed := strings.TrimSpace(expr)
	if IsBlank(trimmed) {
		return time.Time{}, fmt.Errorf("date is required")
	}

	if matches := relativeTimePattern.FindStringSubmatch(trimmed); matches != nil {
		t := time.Now()
		for _, segment := range relativeTimeSegmentPattern.FindAllStringSubmatch(matches[1], -1) {
			amount, err := strconv.Atoi(segment[2])
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid date expression %q", expr)
			}
			if segment[1] == "-" {
				amount = -amount
			}
			switch segment[3] {
			case "s":
				t = t.Add(time.Duration(amount) * time.Second)
			case "m":
				t = t.Add(time.Duration(amount) * time.Minute)
			case "h":
				t = t.Add(time.Duration(amount) * time.Hour)
			case "d":
				t = t.AddDate(0, 0, amount)
			case "w":
				t = t.AddDate(0, 0, amount*7)
			}
		}
		return t, nil
	}

	for _, layout := range absoluteTimeLayouts {
		if t, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			if endOfDay && layout == dateOnlyLayout {
				t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q: expected an ISO-8601 date/time or a now()/now()-1h/now()-1d style expression", expr)
}
