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
	"cwclock-api/internal/utils"
)

// allDayStartMinutes is the wall-clock start assumed for an "all day" entry
// (9:00 AM), so it renders as a plausible work window instead of a literal
// 00:00-23:59 span. allDayLunchStartMinutes/allDayLunchBreakMinutes bake in
// an unpaid 1-hour lunch pause at noon (ai-instruct-61): the morning runs
// 9:00-12:00 (or less, if HoursPerDay leaves fewer than 3 hours to work),
// and the afternoon, starting after the break, makes up whatever of the
// client's HoursPerDay the morning didn't cover - see allDayWindow.
const (
	allDayStartMinutes      = 9 * 60
	allDayLunchStartMinutes = 12 * 60
	allDayLunchBreakMinutes = 60
)

// DayLayout is the plain calendar-day format ("2006-01-02" in Go's reference
// time, i.e. YYYY-MM-DD) used everywhere a report deals with a day as a bare
// date with no time-of-day or timezone component: ReportEntry.Day, the
// start/end query params reports are filtered by, and DailyBuckets' keys.
// Sharing one constant keeps every parse/format of that string in sync.
const DayLayout = "2006-01-02"

// USDateLayout ("01/02/2006", i.e. MM/DD/YYYY) is how a day is displayed to
// users in report output (CSV rows, the PDF header's period line): US date
// order, matching this app's other user-facing date formatting (formatAMPM).
const USDateLayout = "01/02/2006"

// FilenameDateLayout ("01_02_2006", i.e. MM_DD_YYYY) is how a day is written
// into exported filenames. Same field order as USDateLayout, but underscored
// instead of slashed since "/" isn't a valid filename character.
const FilenameDateLayout = "01_02_2006"

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

// allDayWindow returns the display start/end for an "all day" entry: a work
// day starting at 9:00 AM with a 1-hour lunch pause at noon, ending once
// the client's HoursPerDay (falling back to 7h if unset) worth of work is
// done - not a literal 00:00-23:59 span, nor a lunch-free 9:00-to-9:00-
// plus-HoursPerDay span. The morning covers up to 3 hours (9:00-12:00); if
// HoursPerDay is 3 or less, the day ends there with no lunch pause at all.
// Otherwise the afternoon, starting at 13:00, makes up the rest.
func allDayWindow(client models.Client) (start, end string) {
	totalMinutes := int(hoursPerDay(client) * 60)
	morningMinutes := allDayLunchStartMinutes - allDayStartMinutes
	if totalMinutes <= morningMinutes {
		return formatMinutesOfDay(allDayStartMinutes), formatMinutesOfDay(allDayStartMinutes + totalMinutes)
	}
	afternoonMinutes := totalMinutes - morningMinutes
	endMinutes := allDayLunchStartMinutes + allDayLunchBreakMinutes + afternoonMinutes
	return formatMinutesOfDay(allDayStartMinutes), formatMinutesOfDay(endMinutes)
}

// DurationSecs computes an entry's duration: an all-day entry takes its
// client's HoursPerDay (falling back to 7h if unset), otherwise it's the
// wall-clock gap between start and end, treating an end earlier than start
// as crossing midnight. Exported for reuse by the task-duration metric.
func DurationSecs(entry models.TimeEntry, client models.Client) int {
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

// effectiveDailyRate resolves which daily rate bills an entry: the
// project's own rate takes priority over the client's, which in turn takes
// priority over the member's — project is the most specific level, member
// the fallback (see ai-instruct-19).
func effectiveDailyRate(client models.Client, project models.Project, member models.Member) *float64 {
	if project.DailyRate != nil && *project.DailyRate != 0 {
		return project.DailyRate
	}
	if client.DailyRate != nil && *client.DailyRate != 0 {
		return client.DailyRate
	}
	return member.DailyRate
}

// amount converts a duration into a billable amount: hours worked, divided
// by the client's HoursPerDay (a fraction of a full day), times the
// effective daily rate (see effectiveDailyRate).
func amount(durationSecs int, client models.Client, project models.Project, member models.Member) float64 {
	rate := effectiveDailyRate(client, project, member)
	if rate == nil || *rate == 0 {
		return 0
	}
	hours := float64(durationSecs) / 3600
	return (hours / hoursPerDay(client)) * *rate
}

func memberName(m models.Member) string {
	name := strings.TrimSpace(strings.TrimSpace(m.Name) + " " + strings.TrimSpace(m.Surname))
	if utils.IsBlank(name) {
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
		dur := DurationSecs(e, client)

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
			Days:         float64(dur) / 3600 / hoursPerDay(client),
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
			amt := amount(dur, client, project, member)
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
		return utils.EMPTY
	}
	return *e.Start
}

// Totals sums a report's entries. Amount stays nil unless canSeeAmount.
func Totals(entries []models.ReportEntry, canSeeAmount bool, currency string) models.ReportTotals {
	t := models.ReportTotals{Currency: currency}
	var amt float64
	for _, e := range entries {
		t.DurationSecs += e.DurationSecs
		t.Days += e.Days
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
		day := d.Format(DayLayout)
		buckets = append(buckets, models.ReportDailyBucket{Day: day, DurationSecs: byDay[day]})
	}
	return buckets
}

// summaryKey groups entries into one summary line: same project, same task
// label and same user (durations from different users are always kept on
// separate rows, since each row carries a single UserEmail).
type summaryKey struct {
	ProjectID string
	Text      string
	UserID    string
}

// SummaryRows aggregates entries by (project, task label, user): every time
// record sharing all three — regardless of which day or session it was
// logged in — is combined into that one line, summing duration/amount.
// Ranked by duration, most time-consuming task first.
func SummaryRows(entries []models.ReportEntry, canSeeAmount bool) []models.ReportSummaryRow {
	byKey := map[summaryKey]*models.ReportSummaryRow{}
	order := []summaryKey{}

	for _, e := range entries {
		k := summaryKey{ProjectID: e.ProjectID, Text: e.Text, UserID: e.UserID}
		row, ok := byKey[k]
		if !ok {
			row = &models.ReportSummaryRow{
				ProjectID:   e.ProjectID,
				ProjectName: e.ProjectName,
				ClientName:  e.ClientName,
				Description: e.Text,
				UserEmail:   e.UserEmail,
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

// defaultProjectColor matches the frontend's ProjectBadge fallback, used
// whenever a project has no color set.
const defaultProjectColor = "#1cb9f7"

// maxDonutSlices caps how many distinct projects the summary donut chart
// (web and PDF) shows individually; anything past the largest maxDonutSlices
// projects is folded into one "Other" slice, keeping the chart and its PDF
// legend readable and bounded regardless of how many projects a report
// spans. This only affects the donut aggregate - the full-fidelity Rows
// table is untouched.
const maxDonutSlices = 8

// ProjectDurations aggregates entries by project for the summary report's
// donut chart, carrying each project's color (see defaultProjectColor for
// unset colors). Ranked by duration, most time-consuming project first, with
// anything past maxDonutSlices folded into a single gray "Other" slice.
func ProjectDurations(entries []models.ReportEntry, lk Lookups) []models.ReportProjectDuration {
	byProject := map[string]*models.ReportProjectDuration{}
	order := []string{}

	for _, e := range entries {
		pd, ok := byProject[e.ProjectID]
		if !ok {
			color := lk.Projects[e.ProjectID].Color
			if utils.IsBlank(color) {
				color = defaultProjectColor
			}

			pd = &models.ReportProjectDuration{ProjectID: e.ProjectID, ProjectName: e.ProjectName, Color: color}
			byProject[e.ProjectID] = pd
			order = append(order, e.ProjectID)
		}
		pd.DurationSecs += e.DurationSecs
	}

	all := make([]models.ReportProjectDuration, 0, len(order))
	for _, id := range order {
		all = append(all, *byProject[id])
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].DurationSecs > all[j].DurationSecs })

	if len(all) <= maxDonutSlices {
		return all
	}
	other := models.ReportProjectDuration{ProjectName: "Other", Color: "#9ca3af"}
	for _, pd := range all[maxDonutSlices:] {
		other.DurationSecs += pd.DurationSecs
	}
	return append(all[:maxDonutSlices], other)
}
