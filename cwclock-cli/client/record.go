package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

type TimeEntry struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organizationId"`
	ClientID       string  `json:"clientId"`
	ProjectID      string  `json:"projectId"`
	UserID         string  `json:"userId"`
	Text           string  `json:"text"`
	Day            string  `json:"day"`
	Start          *string `json:"start"`
	End            *string `json:"end"`
	AllDay         bool    `json:"allDay"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type TimeEntryList struct {
	Items    []TimeEntry `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	HasMore  bool        `json:"hasMore"`
}

type CreateTimeEntryPayload struct {
	ClientID  string  `json:"clientId"`
	ProjectID string  `json:"projectId"`
	Text      string  `json:"text"`
	Day       string  `json:"day"`
	Start     *string `json:"start,omitempty"`
	End       *string `json:"end,omitempty"`
	AllDay    bool    `json:"allDay"`
}

func timeEntriesPath(orgID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/time-entries"
}

func (c *Client) ListTimeEntries(orgID string, max int) (TimeEntryList, error) {
	path := fmt.Sprintf("%s?page=1&pageSize=%d", timeEntriesPath(orgID), max)

	responseBody, err := c.httpRequest(path, "GET", bytes.Buffer{})
	if err != nil {
		return TimeEntryList{}, err
	}
	defer responseBody.Close()

	var list TimeEntryList
	if err := json.NewDecoder(responseBody).Decode(&list); err != nil {
		return TimeEntryList{}, fmt.Errorf("failed to decode time entries response: %w", err)
	}

	return list, nil
}

func (c *Client) CreateTimeEntry(orgID string, payload CreateTimeEntryPayload) (TimeEntry, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return TimeEntry{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(timeEntriesPath(orgID), "POST", body)
	if err != nil {
		return TimeEntry{}, err
	}
	defer responseBody.Close()

	var entry TimeEntry
	if err := json.NewDecoder(responseBody).Decode(&entry); err != nil {
		return TimeEntry{}, fmt.Errorf("failed to decode created time entry response: %w", err)
	}

	return entry, nil
}

func (c *Client) DeleteTimeEntry(orgID string, id string) error {
	path := timeEntriesPath(orgID) + "/" + url.PathEscape(id)

	responseBody, err := c.httpRequest(path, "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}
