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

var absoluteTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04",
	"2006-01-02",
}

// ParseTimeExpr resolves a date/time expression to an absolute time. It
// supports plain ISO-8601 strings as well as Grafana-style relative
// expressions built from now() and a chain of +/-<amount><unit> segments
// (units: s, m, h, d, w), e.g. now(), now()-1h, now()-1d. Shared by every
// command accepting a date range (record, invoice, ...).
func ParseTimeExpr(expr string) (time.Time, error) {
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
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q: expected an ISO-8601 date/time or a now()/now()-1h/now()-1d style expression", expr)
}
