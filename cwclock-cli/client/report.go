package client

import (
	"bytes"
	"cwclock/utils"
	"encoding/json"
	"net/url"
)

// ReportIDFilter mirrors cwclock-api's idFilter: only IDs is honored, the
// rest of the Clockify-style shape is accepted but ignored server-side.
type ReportIDFilter struct {
	IDs []string `json:"ids"`
}

// ReportRequest mirrors cwclock-api's exportRequest for the summary/detailed
// report endpoints. ExportType is "PDF" or "CSV" to stream back a file
// (matching cwclock's own export scripts) - this CLI command never asks for
// the JSON shape the reports page itself uses.
type ReportRequest struct {
	ExportType     string          `json:"exportType"`
	DateRangeStart string          `json:"dateRangeStart"`
	DateRangeEnd   string          `json:"dateRangeEnd"`
	Clients        *ReportIDFilter `json:"clients,omitempty"`
	Projects       *ReportIDFilter `json:"projects,omitempty"`
}

func reportPath(orgID string, reportType string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/reports/" + reportType
}

// GenerateReport downloads a summary or detailed time report (reportType is
// "summary" or "detailed") as a PDF or CSV file.
func (c *Client) GenerateReport(orgID string, reportType string, req ReportRequest) (data []byte, filename string, err error) {
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, utils.EMPTY, err
	}
	var body bytes.Buffer
	body.Write(encoded)
	return c.httpRequestFile(reportPath(orgID, reportType), "POST", body)
}
