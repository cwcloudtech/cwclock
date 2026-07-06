package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"cwclock-api/internal/models"
)

func formatHMS(durationSecs int) string {
	h := durationSecs / 3600
	m := (durationSecs % 3600) / 60
	s := durationSecs % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func formatDecimalHours(durationSecs int) string {
	return fmt.Sprintf("%.2f", float64(durationSecs)/3600)
}

func formatAmount(amt float64) string {
	return fmt.Sprintf("%.2f", amt)
}

func formatAMPM(hm *string) string {
	if hm == nil {
		return ""
	}
	t, err := time.Parse("15:04", *hm)
	if err != nil {
		return *hm
	}
	return t.Format("03:04PM")
}

func formatUSDate(day string) string {
	d, err := time.Parse("2006-01-02", day)
	if err != nil {
		return day
	}
	return d.Format("01/02/2006")
}

// SummaryCSV renders the "Summary" export: one row per (project,
// description) group. Amount columns are omitted entirely when the caller
// isn't allowed to see billable amounts.
func SummaryCSV(rows []models.ReportSummaryRow, canSeeAmount bool, currency string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{"Project", "Client", "Description", "Time (h)", "Time (decimal)"}
	if canSeeAmount {
		header = append(header, fmt.Sprintf("Amount (%s)", currency))
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, r := range rows {
		record := []string{r.ProjectName, r.ClientName, r.Description, formatHMS(r.DurationSecs), formatDecimalHours(r.DurationSecs)}
		if canSeeAmount {
			amt := 0.0
			if r.Amount != nil {
				amt = *r.Amount
			}
			record = append(record, formatAmount(amt))
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

// DetailedCSV renders the "Detailed" export: one row per time entry. Task,
// Group and Tags are always blank (this app has no such concepts) and
// Billable is always "Yes", kept only for compatibility with Clockify-style
// importers. Billable Rate/Amount columns are omitted for callers not
// allowed to see billable amounts.
func DetailedCSV(entries []models.ReportEntry, canSeeAmount bool, currency string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{
		"Project", "Client", "Description", "Task", "User", "Group", "Email", "Tags", "Billable",
		"Start Date", "Start Time", "End Date", "End Time", "Duration (h)", "Duration (decimal)",
	}
	if canSeeAmount {
		header = append(header, fmt.Sprintf("Billable Rate (%s)", currency), fmt.Sprintf("Billable Amount (%s)", currency))
	}
	header = append(header, "Date of creation")
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, e := range entries {
		dayStr := formatUSDate(e.Day)
		startStr, endStr := "12:00AM", "11:59PM"
		if !e.AllDay {
			startStr, endStr = formatAMPM(e.Start), formatAMPM(e.End)
		}

		record := []string{
			e.ProjectName, e.ClientName, e.Text, "", e.UserName, "", e.UserEmail, "", "Yes",
			dayStr, startStr, dayStr, endStr, formatHMS(e.DurationSecs), formatDecimalHours(e.DurationSecs),
		}
		if canSeeAmount {
			rate, amt := 0.0, 0.0
			if e.Amount != nil {
				amt = *e.Amount
				if hours := float64(e.DurationSecs) / 3600; hours > 0 {
					rate = amt / hours
				}
			}
			record = append(record, formatAmount(rate), formatAmount(amt))
		}
		record = append(record, e.CreatedAt.Format("01/02/2006"))

		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}
