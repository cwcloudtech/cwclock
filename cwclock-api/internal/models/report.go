package models

import "time"

// ReportEntry is a time entry enriched with the display data a report needs
// (client/project/member names) plus its computed duration and, when the
// caller is allowed to see it, its billable amount.
type ReportEntry struct {
	ID           string    `json:"id"`
	Day          string    `json:"day"`
	Start        *string   `json:"start"`
	End          *string   `json:"end"`
	AllDay       bool      `json:"allDay"`
	DurationSecs int       `json:"durationSecs"`
	Text         string    `json:"text"`
	ClientID     string    `json:"clientId"`
	ClientName   string    `json:"clientName"`
	ProjectID    string    `json:"projectId"`
	ProjectName  string    `json:"projectName"`
	UserID       string    `json:"userId"`
	UserName     string    `json:"userName"`
	UserEmail    string    `json:"userEmail"`
	CreatedAt    time.Time `json:"createdAt"`
	Amount       *float64  `json:"amount,omitempty"`
}

// ReportTotals summarizes a report's overall duration/amount. Amount is nil
// for callers not allowed to see billable amounts (readers are blocked
// entirely; members never see amounts, only admins/owner do).
type ReportTotals struct {
	DurationSecs int      `json:"durationSecs"`
	Amount       *float64 `json:"amount,omitempty"`
	Currency     string   `json:"currency"`
}

// ReportSummaryRow aggregates entries sharing the same project+description
// ("aggregated by redundant labels") into a single summary line.
type ReportSummaryRow struct {
	ProjectID    string   `json:"projectId"`
	ProjectName  string   `json:"projectName"`
	ClientName   string   `json:"clientName"`
	Description  string   `json:"description"`
	DurationSecs int      `json:"durationSecs"`
	Amount       *float64 `json:"amount,omitempty"`
}

// ReportDailyBucket is one point of the summary report's per-day duration
// chart, one entry per calendar day in the requested range (even empty days).
type ReportDailyBucket struct {
	Day          string `json:"day"`
	DurationSecs int    `json:"durationSecs"`
}

type DetailedReport struct {
	Totals  ReportTotals  `json:"totals"`
	Entries []ReportEntry `json:"entries"`
}

type SummaryReport struct {
	Totals ReportTotals        `json:"totals"`
	Daily  []ReportDailyBucket `json:"daily"`
	Rows   []ReportSummaryRow  `json:"rows"`
}
