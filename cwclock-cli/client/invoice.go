package client

import (
	"bytes"
	"cwclock/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Invoice mirrors cwclock-api's models.Invoice. The rendered PDF is
// deliberately not part of this struct - it's fetched separately as raw
// bytes by Preview/Generate.
type Invoice struct {
	ID                string  `json:"id"`
	OrganizationID    string  `json:"organizationId"`
	ClientID          string  `json:"clientId"`
	Number            string  `json:"number"`
	Status            string  `json:"status"`
	TotalHT           float64 `json:"totalHT"`
	TotalVAT          float64 `json:"totalVAT"`
	TotalTTC          float64 `json:"totalTTC"`
	SelectedBeginDate string  `json:"selectedBeginDate"`
	SelectedEndDate   string  `json:"selectedEndDate"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

// InvoiceRequest mirrors cwclock-api's invoiceRequest: one client and a date
// range, accepted by both the preview and generate endpoints.
type InvoiceRequest struct {
	ClientID       string   `json:"clientId"`
	DateRangeStart string   `json:"dateRangeStart"`
	DateRangeEnd   string   `json:"dateRangeEnd"`
	ProjectIDs     []string `json:"projectIds,omitempty"`
}

func invoicesPath(orgID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/invoices"
}

func invoicePath(orgID string, id string) string {
	return invoicesPath(orgID) + "/" + url.PathEscape(id)
}

func encodeInvoiceRequest(req InvoiceRequest) (bytes.Buffer, error) {
	var body bytes.Buffer
	encoded, err := json.Marshal(req)
	if err != nil {
		return body, err
	}
	body.Write(encoded)
	return body, nil
}

// PreviewInvoice renders an invoice PDF without saving anything server-side.
func (c *Client) PreviewInvoice(orgID string, req InvoiceRequest) (data []byte, filename string, err error) {
	body, err := encodeInvoiceRequest(req)
	if err != nil {
		return nil, utils.EMPTY, err
	}
	return c.httpRequestFile(invoicesPath(orgID)+"/preview", "POST", body)
}

// GenerateInvoice renders and saves an invoice, returning the same PDF the
// server persisted. Use ListInvoices afterwards to retrieve the created
// invoice's ID (the generate response is the raw PDF, not JSON).
func (c *Client) GenerateInvoice(orgID string, req InvoiceRequest) (data []byte, filename string, err error) {
	body, err := encodeInvoiceRequest(req)
	if err != nil {
		return nil, utils.EMPTY, err
	}
	return c.httpRequestFile(invoicesPath(orgID), "POST", body)
}

// ListInvoices returns an organization's invoices for one client within a
// date range, most recent first.
func (c *Client) ListInvoices(orgID string, clientID string, start string, end string) ([]Invoice, error) {
	path := fmt.Sprintf("%s?clientId=%s&start=%s&end=%s",
		invoicesPath(orgID), url.QueryEscape(clientID), url.QueryEscape(start), url.QueryEscape(end))

	responseBody, err := c.httpRequest(path, "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var invoices []Invoice
	if err := json.NewDecoder(responseBody).Decode(&invoices); err != nil {
		return nil, fmt.Errorf("failed to decode invoices response: %w", err)
	}
	return invoices, nil
}

// SendInvoice emails an already-generated invoice's PDF to its client.
func (c *Client) SendInvoice(orgID string, id string) error {
	responseBody, err := c.httpRequest(invoicePath(orgID, id)+"/send", "POST", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}

// UploadInvoice (re)uploads an already-generated invoice's PDF to every one
// of its organization's external connections.
func (c *Client) UploadInvoice(orgID string, id string) error {
	responseBody, err := c.httpRequest(invoicePath(orgID, id)+"/reupload", "POST", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}

// DeleteInvoice removes an invoice (and its stored PDF) entirely.
func (c *Client) DeleteInvoice(orgID string, id string) error {
	responseBody, err := c.httpRequest(invoicePath(orgID, id), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}
