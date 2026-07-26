package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// OrgClient is named to avoid colliding with the Client type (the HTTP API
// client itself); it mirrors cwclock-api's models.Client, an organization's
// billing customer.
type OrgClient struct {
	ID                     string   `json:"id"`
	OrganizationID         string   `json:"organizationId"`
	Name                   string   `json:"name"`
	Email                  string   `json:"email,omitempty"`
	InvoiceEmails          string   `json:"invoiceEmails,omitempty"`
	ContactName            string   `json:"contactName"`
	Address                string   `json:"address"`
	PostalCode             string   `json:"postalCode"`
	City                   string   `json:"city"`
	Country                string   `json:"country"`
	VATNumber              string   `json:"vatNumber"`
	VATRate                float64  `json:"vatRate"`
	VATDischargeMotive     string   `json:"vatDischargeMotive"`
	SIREN                  string   `json:"siren"`
	SIRET                  string   `json:"siret"`
	NAF                    string   `json:"naf"`
	MF                     string   `json:"mf"`
	IdentificationNumber   string   `json:"identificationNumber"`
	PurchaseOrder          string   `json:"purchaseOrder"`
	HoursPerDay            float64  `json:"hoursPerDay"`
	DailyRate              *float64 `json:"dailyRate,omitempty"`
	SendReportsWithInvoice bool     `json:"sendReportsWithInvoice"`
	CreatedAt              string   `json:"createdAt"`
	UpdatedAt              string   `json:"updatedAt"`
}

// OrgClientPayload mirrors cwclock-api's clientPayload: every field is
// resent on both create and update (the API replaces the whole resource).
type OrgClientPayload struct {
	Name                   string   `json:"name"`
	Email                  string   `json:"email"`
	InvoiceEmails          string   `json:"invoiceEmails"`
	ContactName            string   `json:"contactName"`
	Address                string   `json:"address"`
	PostalCode             string   `json:"postalCode"`
	City                   string   `json:"city"`
	Country                string   `json:"country"`
	VATNumber              string   `json:"vatNumber"`
	VATRate                *float64 `json:"vatRate"`
	VATDischargeMotive     string   `json:"vatDischargeMotive"`
	SIREN                  string   `json:"siren"`
	SIRET                  string   `json:"siret"`
	NAF                    string   `json:"naf"`
	MF                     string   `json:"mf"`
	IdentificationNumber   string   `json:"identificationNumber"`
	PurchaseOrder          string   `json:"purchaseOrder"`
	HoursPerDay            float64  `json:"hoursPerDay"`
	DailyRate              *float64 `json:"dailyRate"`
	SendReportsWithInvoice bool     `json:"sendReportsWithInvoice"`
}

func orgClientsPath(orgID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/clients"
}

func orgClientPath(orgID string, id string) string {
	return orgClientsPath(orgID) + "/" + url.PathEscape(id)
}

func (c *Client) ListOrgClients(orgID string) ([]OrgClient, error) {
	responseBody, err := c.httpRequest(orgClientsPath(orgID), "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var clients []OrgClient
	if err := json.NewDecoder(responseBody).Decode(&clients); err != nil {
		return nil, fmt.Errorf("failed to decode clients response: %w", err)
	}
	return clients, nil
}

func (c *Client) CreateOrgClient(orgID string, payload OrgClientPayload) (OrgClient, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return OrgClient{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(orgClientsPath(orgID), "POST", body)
	if err != nil {
		return OrgClient{}, err
	}
	defer responseBody.Close()

	var orgClient OrgClient
	if err := json.NewDecoder(responseBody).Decode(&orgClient); err != nil {
		return OrgClient{}, fmt.Errorf("failed to decode created client response: %w", err)
	}
	return orgClient, nil
}

func (c *Client) UpdateOrgClient(orgID string, id string, payload OrgClientPayload) (OrgClient, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return OrgClient{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(orgClientPath(orgID, id), "PUT", body)
	if err != nil {
		return OrgClient{}, err
	}
	defer responseBody.Close()

	var orgClient OrgClient
	if err := json.NewDecoder(responseBody).Decode(&orgClient); err != nil {
		return OrgClient{}, fmt.Errorf("failed to decode updated client response: %w", err)
	}
	return orgClient, nil
}

func (c *Client) DeleteOrgClient(orgID string, id string) error {
	responseBody, err := c.httpRequest(orgClientPath(orgID, id), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}
