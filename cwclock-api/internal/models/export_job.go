package models

import "time"

// ExportJob defines a scheduled job that exports reports and sends them to target recipients.
type ExportJob struct {
	ID               string         `json:"id"`
	OrganizationID   string         `json:"organizationId"`
	Name             string         `json:"name"`
	CronExpression   string         `json:"cronExpression"`
	Targets          []ExportTarget `json:"targets"`
	ReportTypes      []string       `json:"reportTypes"` // "summary-pdf", "summary-csv", "detailed-pdf", "detailed-csv", "invoices-pdf"
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
	Type string `json:"type"` // "email", "s3", "google_drive", "git"
	// ToEmails/CCEmails are raw comma/semicolon-separated address lists, the
	// same format as models.Client.InvoiceEmails - split with
	// utils.SplitList at send time (see handlers.ExportDeliveryService).
	ToEmails string `json:"toEmails,omitempty"`
	CCEmails string `json:"ccEmails,omitempty"`
	// Connection holds the target's own S3/Google Drive/git credentials,
	// captured through the same form/fields as an organization's external
	// connections (see ExternalConnection), but stored independently in the
	// export job's own data payload rather than referencing one of the
	// organization's connections by id - so a job can push to a completely
	// different destination (e.g. its own S3 bucket) than the one
	// configured for invoices, unaffected by that connection later being
	// edited or removed.
	Connection *ExternalConnection `json:"connection,omitempty"`
}
