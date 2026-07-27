package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// ExportTarget defines where an export job's reports get sent. For a
// "email" target, ToEmails/CCEmails are used; for "s3"/"google_drive"/"git"
// targets, Connection carries that destination's own credentials.
type ExportTarget struct {
	Type       string              `json:"type"`
	ToEmails   string              `json:"toEmails,omitempty"`
	CCEmails   string              `json:"ccEmails,omitempty"`
	Connection *ExternalConnection `json:"connection,omitempty"`
}

type ExportJob struct {
	ID               string         `json:"id"`
	OrganizationID   string         `json:"organizationId"`
	Name             string         `json:"name"`
	CronExpression   string         `json:"cronExpression"`
	Targets          []ExportTarget `json:"targets"`
	ReportTypes      []string       `json:"reportTypes"`
	TimePeriod       string         `json:"timePeriod"`
	ClientIDs        []string       `json:"clientIds"`
	ProjectIDs       []string       `json:"projectIds"`
	IncludeFinancial bool           `json:"includeFinancial"`
	Enabled          bool           `json:"enabled"`
	NextRunAt        string         `json:"nextRunAt,omitempty"`
	NextRunInSeconds *int64         `json:"nextRunInSeconds,omitempty"`
	CreatedAt        string         `json:"createdAt"`
	UpdatedAt        string         `json:"updatedAt"`
}

// ExportJobPayload mirrors cwclock-api's exportJobPayload: every field is
// resent on both create and update (the API replaces the whole resource).
type ExportJobPayload struct {
	Name             string         `json:"name"`
	CronExpression   string         `json:"cronExpression"`
	Targets          []ExportTarget `json:"targets"`
	ReportTypes      []string       `json:"reportTypes"`
	TimePeriod       string         `json:"timePeriod"`
	ClientIDs        []string       `json:"clientIds"`
	ProjectIDs       []string       `json:"projectIds"`
	IncludeFinancial bool           `json:"includeFinancial"`
	Enabled          bool           `json:"enabled"`
}

func exportJobsPath(orgID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/export-jobs"
}

func exportJobPath(orgID string, id string) string {
	return exportJobsPath(orgID) + "/" + url.PathEscape(id)
}

func (c *Client) ListExportJobs(orgID string) ([]ExportJob, error) {
	responseBody, err := c.httpRequest(exportJobsPath(orgID), "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var jobs []ExportJob
	if err := json.NewDecoder(responseBody).Decode(&jobs); err != nil {
		return nil, fmt.Errorf("failed to decode export jobs response: %w", err)
	}
	return jobs, nil
}

func (c *Client) CreateExportJob(orgID string, payload ExportJobPayload) (ExportJob, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ExportJob{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(exportJobsPath(orgID), "POST", body)
	if err != nil {
		return ExportJob{}, err
	}
	defer responseBody.Close()

	var job ExportJob
	if err := json.NewDecoder(responseBody).Decode(&job); err != nil {
		return ExportJob{}, fmt.Errorf("failed to decode created export job response: %w", err)
	}
	return job, nil
}

func (c *Client) UpdateExportJob(orgID string, id string, payload ExportJobPayload) (ExportJob, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ExportJob{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(exportJobPath(orgID, id), "PUT", body)
	if err != nil {
		return ExportJob{}, err
	}
	defer responseBody.Close()

	var job ExportJob
	if err := json.NewDecoder(responseBody).Decode(&job); err != nil {
		return ExportJob{}, fmt.Errorf("failed to decode updated export job response: %w", err)
	}
	return job, nil
}

func (c *Client) DeleteExportJob(orgID string, id string) error {
	responseBody, err := c.httpRequest(exportJobPath(orgID, id), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}

// RunExportJob triggers an immediate, out-of-schedule run of an export job -
// see cwclock-api's ExportJobHandler.RunNow. It runs synchronously
// server-side, so a nil error confirms the run completed.
func (c *Client) RunExportJob(orgID string, id string) error {
	responseBody, err := c.httpRequest(exportJobPath(orgID, id)+"/run", "POST", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}
