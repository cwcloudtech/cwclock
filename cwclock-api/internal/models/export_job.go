package models

import "time"

// ExportJob defines a scheduled job that exports reports and sends them to target recipients.
type ExportJob struct {
	ID               string         `json:"id"`
	OrganizationID   string         `json:"organizationId"`
	Name             string         `json:"name"`
	CronExpression   string         `json:"cronExpression"`
	Targets          []ExportTarget `json:"targets"`
	ReportTypes      []string       `json:"reportTypes"` // "summary-pdf", "summary-csv", "detailed-pdf", "detailed-csv"
	TimePeriod       string         `json:"timePeriod"`  // e.g., "now()", "now()-1d", "now()-1h"
	ClientIDs        []string       `json:"clientIds"`   // Empty = all clients
	ProjectIDs       []string       `json:"projectIds"`  // Empty = all projects
	IncludeFinancial bool           `json:"includeFinancial"`
	Enabled          bool           `json:"enabled"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

// ExportTarget defines where reports should be sent.
type ExportTarget struct {
	Type       string   `json:"type"` // "email", "s3", "google_drive"
	ToEmails   []string `json:"toEmails,omitempty"`
	CCEmails   []string `json:"ccEmails,omitempty"`
	Connection string   `json:"connection,omitempty"` // Connection ID for S3/Google Drive
}
