package client

import (
	"bytes"
	"cwclock/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

type Project struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organizationId"`
	ClientID       string   `json:"clientId"`
	Name           string   `json:"name"`
	Color          string   `json:"color"`
	DailyRate      *float64 `json:"dailyRate,omitempty"`
	Subdivisions   []string `json:"subdivisions,omitempty"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

// ProjectPayload mirrors cwclock-api's projectPayload. ClientID is only
// honored by Update (as an optional reassignment); Create takes its client
// from the URL path instead (see CreateProject).
type ProjectPayload struct {
	ClientID     string   `json:"clientId,omitempty"`
	Name         string   `json:"name"`
	Color        string   `json:"color"`
	DailyRate    *float64 `json:"dailyRate"`
	Subdivisions []string `json:"subdivisions"`
}

func projectsPath(orgID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/projects"
}

func projectPath(orgID string, id string) string {
	return projectsPath(orgID) + "/" + url.PathEscape(id)
}

func (c *Client) ListProjects(orgID string, clientID string) ([]Project, error) {
	path := projectsPath(orgID)
	if utils.IsNotBlank(clientID) {
		path += "?clientId=" + url.QueryEscape(clientID)
	}

	responseBody, err := c.httpRequest(path, "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var projects []Project
	if err := json.NewDecoder(responseBody).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode projects response: %w", err)
	}
	return projects, nil
}

func (c *Client) CreateProject(orgID string, clientID string, payload ProjectPayload) (Project, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Project{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	path := "/organizations/" + url.PathEscape(orgID) + "/clients/" + url.PathEscape(clientID) + "/projects"
	responseBody, err := c.httpRequest(path, "POST", body)
	if err != nil {
		return Project{}, err
	}
	defer responseBody.Close()

	var project Project
	if err := json.NewDecoder(responseBody).Decode(&project); err != nil {
		return Project{}, fmt.Errorf("failed to decode created project response: %w", err)
	}
	return project, nil
}

func (c *Client) UpdateProject(orgID string, id string, payload ProjectPayload) (Project, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Project{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(projectPath(orgID, id), "PUT", body)
	if err != nil {
		return Project{}, err
	}
	defer responseBody.Close()

	var project Project
	if err := json.NewDecoder(responseBody).Decode(&project); err != nil {
		return Project{}, fmt.Errorf("failed to decode updated project response: %w", err)
	}
	return project, nil
}

func (c *Client) DeleteProject(orgID string, id string) error {
	responseBody, err := c.httpRequest(projectPath(orgID, id), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}
