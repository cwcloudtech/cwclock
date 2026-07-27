package handlers

import (
	"cwclock/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// dateRangeTimeLayout is the "YYYY-MM-DDTHH:MM:SS" shape cwclock-api's
// invoice and report endpoints expect for dateRangeStart/dateRangeEnd (see
// utils.ParseTimeExpr for the accepted input expressions).
const dateRangeTimeLayout = "2006-01-02T15:04:05"

// resolveDateRangeParams parses --begin and --end (or its --to alias) into
// the strings the invoice and export endpoints expect. Shared by both since
// neither is restricted to a single day the way a time record is.
func resolveDateRangeParams(beginExpr string, endExpr string, toExpr string) (start string, end string, err error) {
	effectiveEnd := utils.If(utils.IsNotBlank(endExpr), endExpr, toExpr)

	if utils.IsBlank(beginExpr) {
		return utils.EMPTY, utils.EMPTY, fmt.Errorf("begin date is required: use --begin")
	}
	if utils.IsBlank(effectiveEnd) {
		return utils.EMPTY, utils.EMPTY, fmt.Errorf("end date is required: use --end or --to")
	}

	begin, err := utils.ParseTimeExpr(beginExpr)
	if err != nil {
		return utils.EMPTY, utils.EMPTY, fmt.Errorf("invalid begin date: %w", err)
	}
	endTime, err := utils.ParseTimeExpr(effectiveEnd)
	if err != nil {
		return utils.EMPTY, utils.EMPTY, fmt.Errorf("invalid end date: %w", err)
	}

	return begin.Format(dateRangeTimeLayout), endTime.Format(dateRangeTimeLayout), nil
}

// normalizeIDs trims a repeatable id flag's values (e.g. --client, --project
// passed multiple times) and drops any blanks, returning nil when nothing is
// left so it serializes as an omitted field rather than an empty array.
func normalizeIDs(values []string) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); utils.IsNotBlank(trimmed) {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// dayPart extracts the leading "YYYY-MM-DD" from a dateRangeTimeLayout
// string, matching cwclock-api's own dayPart used to filter by day.
func dayPart(v string) string {
	if len(v) >= len(dayLayout) {
		return v[:len(dayLayout)]
	}
	return v
}

// saveDownloadedFile writes downloaded binary data (an invoice or report
// file) to outputOverride when set, otherwise to the server-suggested
// filename (from Content-Disposition), falling back to fallbackFilename
// when neither is available. Returns the absolute path it wrote to.
func saveDownloadedFile(data []byte, serverFilename string, outputOverride string, fallbackFilename string) (string, error) {
	path := strings.TrimSpace(outputOverride)
	if utils.IsBlank(path) {
		path = strings.TrimSpace(serverFilename)
	}
	if utils.IsBlank(path) {
		path = fallbackFilename
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return utils.EMPTY, fmt.Errorf("failed to save file: %w", err)
	}

	if abs, err := filepath.Abs(path); err == nil {
		return abs, nil
	}
	return path, nil
}
