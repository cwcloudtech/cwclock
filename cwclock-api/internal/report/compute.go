// Package report computes and renders time-tracking reports (summary and
// detailed), independent of HTTP concerns so it can be unit-friendly and,
// per the PDF renderer, reused by a future invoicing feature.
package report

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"cwclock-api/internal/models"
)

// allDayStartMinutes is the wall-clock start assumed for an "all day" entry
// (9:00 AM), so it renders as a plausible work window instead of a literal
// 00:00-23:59 span.
const allDayStartMinutes = 9 * 60

// Lookups holds the related records a raw TimeEntry needs to be enriched
// into a ReportEntry, keyed by ID.
type Lookups struct {
	Clients  map[string]models.Client
	Projects map[string]models.Project
	Members  map[string]models.Member // keyed by userID
}

const fallbackHoursPerDay = 7

func hoursPerDay(client models.Client) float64 {
	if client.HoursPerDay <= 0 {
		return fallbackHoursPerDay
	}
	return client.HoursPerDay
}

// parseSecondsOfDay parses a wall-clock string into seconds since midnight.
// The time inputs that produce Start/End use step="1" (second granularity),
// so the value is "HH:MM:SS", not "HH:MM" — but this still accepts a bare
// "HH:MM" (as used by allDayWindow's synthetic display window below) by
// treating a missing part as zero.
func parseSecondsOfDay(hms string) int {
	parts := strings.Split(hms, ":")
	get := func(i int) int {
		if i >= len(parts) {
			return 0
		}
		v, _ := strconv.Atoi(parts[i])
		return v
	}
	return get(0)*3600 + get(1)*60 + get(2)
}

func formatMinutesOfDay(min int) string {
	return fmt.Sprintf("%02d:%02d", min/60, min%60)
}

// allDayWindow returns the display start/end for an "all day" entry: 9:00 AM
// through 9:00 AM plus the client's HoursPerDay (falling back to 7h if
// unset) — not a literal 00:00-23:59 span.
func allDayWindow(client models.Client) (start, end string) {
	endMinutes := allDayStartMinutes + int(hoursPerDay(client)*60)
	return formatMinutesOfDay(allDayStartMinutes), formatMinutesOfDay(endMinutes)
}

// durationSecs computes an entry's duration: an all-day entry takes its
// client's HoursPerDay (falling back to 7h if unset), otherwise it's the
// wall-clock gap between start and end, treating an end earlier than start
// as crossing midnight.
func durationSecs(entry models.TimeEntry, client models.Client) int {
	if entry.AllDay {
		return int(hoursPerDay(client) * 3600)
	}
	if entry.Start == nil || entry.End == nil {
		return 0
	}
	startSec := parseSecondsOfDay(*entry.Start)
	endSec := parseSecondsOfDay(*entry.End)
	if endSec < startSec {
		endSec += 24 * 3600
	}
	return endSec - startSec
}

// amount converts a duration into a billable amount: hours worked, divided
// by the client's HoursPerDay (a fraction of a full day), times the
// member's daily rate.
func amount(durationSecs int, client models.Client, member models.Member) float64 {
	if member.DailyRate == nil || *member.DailyRate == 0 {
		return 0
	}
	hours := float64(durationSecs) / 3600
	return (hours / hoursPerDay(client)) * *member.DailyRate
}

func memberName(m models.Member) string {
	name := strings.TrimSpace(strings.TrimSpace(m.Name) + " " + strings.TrimSpace(m.Surname))
	if name == "" {
		return m.Email
	}
	return name
}

// Enrich turns raw time entries into ReportEntry values. Amount is left nil
// for every entry unless canSeeAmount is true (readers can't reach this code
// path at all; members can see their hours but never a priced amount).
func Enrich(entries []models.TimeEntry, lk Lookups, canSeeAmount bool) []models.ReportEntry {
	out := make([]models.ReportEntry, 0, len(entries))
	for _, e := range entries {
		client := lk.Clients[e.ClientID]
		project := lk.Projects[e.ProjectID]
		member := lk.Members[e.UserID]
		dur := durationSecs(e, client)

		start, end := e.Start, e.End
		if e.AllDay {
			s, en := allDayWindow(client)
			start, end = &s, &en
		}

		re := models.ReportEntry{
			ID:           e.ID,
			Day:          e.Day,
			Start:        start,
			End:          end,
			AllDay:       e.AllDay,
			DurationSecs: dur,
			Text:         e.Text,
			ClientID:     e.ClientID,
			ClientName:   client.Name,
			ProjectID:    e.ProjectID,
			ProjectName:  project.Name,
			UserID:       e.UserID,
			UserName:     memberName(member),
			UserEmail:    member.Email,
			CreatedAt:    e.CreatedAt,
		}
		if canSeeAmount {
			amt := amount(dur, client, member)
			re.Amount = &amt
		}
		out = append(out, re)
	}

	// Most recent first, matching the reference report's ordering.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Day != out[j].Day {
			return out[i].Day > out[j].Day
		}
		return startOf(out[i]) > startOf(out[j])
	})
	return out
}

func startOf(e models.ReportEntry) string {
	if e.Start == nil {
		return ""
	}
	return *e.Start
}

// Totals sums a report's entries. Amount stays nil unless canSeeAmount.
func Totals(entries []models.ReportEntry, canSeeAmount bool, currency string) models.ReportTotals {
	t := models.ReportTotals{Currency: currency}
	var amt float64
	for _, e := range entries {
		t.DurationSecs += e.DurationSecs
		if e.Amount != nil {
			amt += *e.Amount
		}
	}
	if canSeeAmount {
		t.Amount = &amt
	}
	return t
}

// DailyBuckets sums duration per calendar day across [start, end], including
// days with no entries, for the summary report's chart.
func DailyBuckets(entries []models.ReportEntry, start, end time.Time) []models.ReportDailyBucket {
	byDay := map[string]int{}
	for _, e := range entries {
		byDay[e.Day] += e.DurationSecs
	}
	buckets := []models.ReportDailyBucket{}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		day := d.Format("2006-01-02")
		buckets = append(buckets, models.ReportDailyBucket{Day: day, DurationSecs: byDay[day]})
	}
	return buckets
}

// SummaryRows aggregates entries by task name/label: a "task" is one summary
// line, and every time record sharing the same name — regardless of which
// day or session it was logged in — is combined into that one line, summing
// duration/amount. Ranked by duration, most time-consuming task first.
func SummaryRows(entries []models.ReportEntry, canSeeAmount bool) []models.ReportSummaryRow {
	byKey := map[string]*models.ReportSummaryRow{}
	order := []string{}

	for _, e := range entries {
		k := e.Text
		row, ok := byKey[k]
		if !ok {
			row = &models.ReportSummaryRow{
				ProjectID:   e.ProjectID,
				ProjectName: e.ProjectName,
				ClientName:  e.ClientName,
				Description: e.Text,
			}
			if canSeeAmount {
				zero := 0.0
				row.Amount = &zero
			}
			byKey[k] = row
			order = append(order, k)
		}
		row.DurationSecs += e.DurationSecs
		if canSeeAmount && e.Amount != nil {
			*row.Amount += *e.Amount
		}
	}

	rows := make([]models.ReportSummaryRow, 0, len(order))
	for _, k := range order {
		rows = append(rows, *byKey[k])
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].DurationSecs > rows[j].DurationSecs })
	return rows
}
