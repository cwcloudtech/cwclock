package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

type Organization struct {
	ID                   string               `json:"id"`
	OwnerID              string               `json:"ownerId"`
	Name                 string               `json:"name"`
	Email                string               `json:"email"`
	AccountingEmail      string               `json:"accountingEmail,omitempty"`
	Address              string               `json:"address"`
	PostalCode           string               `json:"postalCode"`
	City                 string               `json:"city"`
	Country              string               `json:"country"`
	VATNumber            string               `json:"vatNumber"`
	SIREN                string               `json:"siren"`
	SIRET                string               `json:"siret"`
	NAF                  string               `json:"naf"`
	MF                   string               `json:"mf"`
	IdentificationNumber string               `json:"identificationNumber"`
	IBAN                 string               `json:"iban,omitempty"`
	BIC                  string               `json:"bic,omitempty"`
	Currency             string               `json:"currency"`
	ExternalConnections  []ExternalConnection `json:"externalConnections"`
	CreatedAt            string               `json:"createdAt"`
	UpdatedAt            string               `json:"updatedAt"`
}

// OrganizationPayload mirrors cwclock-api's organizationPayload: every
// field is resent on both create and update (the API replaces the whole
// resource, it doesn't merge a partial patch).
type OrganizationPayload struct {
	Name                 string `json:"name"`
	AccountingEmail      string `json:"accountingEmail"`
	Address              string `json:"address"`
	PostalCode           string `json:"postalCode"`
	City                 string `json:"city"`
	Country              string `json:"country"`
	VATNumber            string `json:"vatNumber"`
	SIREN                string `json:"siren"`
	SIRET                string `json:"siret"`
	NAF                  string `json:"naf"`
	MF                   string `json:"mf"`
	IdentificationNumber string `json:"identificationNumber"`
	IBAN                 string `json:"iban"`
	BIC                  string `json:"bic"`
	Currency             string `json:"currency"`
}

func organizationsPath() string {
	return "/organizations"
}

func organizationPath(id string) string {
	return "/organizations/" + url.PathEscape(id)
}

func (c *Client) GetOrganization(id string) (Organization, error) {
	responseBody, err := c.httpRequest(organizationPath(id), "GET", bytes.Buffer{})
	if err != nil {
		return Organization{}, err
	}
	defer responseBody.Close()

	var org Organization
	if err := json.NewDecoder(responseBody).Decode(&org); err != nil {
		return Organization{}, fmt.Errorf("failed to decode organization response: %w", err)
	}
	return org, nil
}

func (c *Client) ListOrganizations() ([]Organization, error) {
	responseBody, err := c.httpRequest(organizationsPath(), "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var orgs []Organization
	if err := json.NewDecoder(responseBody).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("failed to decode organizations response: %w", err)
	}
	return orgs, nil
}

func (c *Client) CreateOrganization(payload OrganizationPayload) (Organization, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Organization{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(organizationsPath(), "POST", body)
	if err != nil {
		return Organization{}, err
	}
	defer responseBody.Close()

	var org Organization
	if err := json.NewDecoder(responseBody).Decode(&org); err != nil {
		return Organization{}, fmt.Errorf("failed to decode created organization response: %w", err)
	}
	return org, nil
}

func (c *Client) UpdateOrganization(id string, payload OrganizationPayload) (Organization, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Organization{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(organizationPath(id), "PATCH", body)
	if err != nil {
		return Organization{}, err
	}
	defer responseBody.Close()

	var org Organization
	if err := json.NewDecoder(responseBody).Decode(&org); err != nil {
		return Organization{}, fmt.Errorf("failed to decode updated organization response: %w", err)
	}
	return org, nil
}

func (c *Client) DeleteOrganization(id string) error {
	responseBody, err := c.httpRequest(organizationPath(id), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}

func (c *Client) AddOrganizationExternalConnection(orgID string, conn ExternalConnection) (Organization, error) {
	encoded, err := json.Marshal(conn)
	if err != nil {
		return Organization{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(organizationPath(orgID)+"/external-connections", "PATCH", body)
	if err != nil {
		return Organization{}, err
	}
	defer responseBody.Close()

	var org Organization
	if err := json.NewDecoder(responseBody).Decode(&org); err != nil {
		return Organization{}, fmt.Errorf("failed to decode organization response: %w", err)
	}
	return org, nil
}

func (c *Client) RemoveOrganizationExternalConnection(orgID string, connectionID string) (Organization, error) {
	path := organizationPath(orgID) + "/external-connections/" + url.PathEscape(connectionID)

	responseBody, err := c.httpRequest(path, "PATCH", bytes.Buffer{})
	if err != nil {
		return Organization{}, err
	}
	defer responseBody.Close()

	var org Organization
	if err := json.NewDecoder(responseBody).Decode(&org); err != nil {
		return Organization{}, fmt.Errorf("failed to decode organization response: %w", err)
	}
	return org, nil
}
