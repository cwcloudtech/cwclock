// Package contact forwards CWClock's contact form submissions to CWCloud's
// contact-request API (POST {apiURL}/v1/contactreq).
package contact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cwclock-api/internal/utils"
)

// request is the JSON body CWCloud's /v1/contactreq expects. Unlike every
// other CWCloud API call this app makes, it takes no X-Api-Key/X-Auth-Token
// header at all - id (the contact form's uuid, see Client.formID) is what
// scopes the submission on CWCloud's side instead.
type request struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Name      string `json:"name,omitempty"`
	Firstname string `json:"firstname,omitempty"`
}

// Submission is one contact form submission to send.
type Submission struct {
	Email     string
	Subject   string
	Message   string
	Name      string
	Firstname string
}

// Client posts contact form submissions to CWCloud's contact-request API.
type Client struct {
	apiURL string
	formID string
	client *http.Client
}

// New builds a Client for the given CWCloud API base URL and contact form
// id (CWCLOUD_CONTACT_FORM_ID). formID may be blank - see Configured.
func New(apiURL, formID string) *Client {
	return &Client{apiURL: apiURL, formID: formID, client: &http.Client{Timeout: 15 * time.Second}}
}

// Configured reports whether a contact form id is set. Callers should check
// this before calling Send and reject the request themselves (see
// ai-instruct-54: "If this variable is not set, a 405 error will be sent")
// rather than relying on Send to fail.
func (c *Client) Configured() bool {
	return utils.IsNotBlank(c.formID)
}

// Send posts one contact form submission. Returns an error if the request
// couldn't be built/sent or CWCloud's API responded with a non-2xx status -
// unlike internal/email, this is a live user-facing form, so failures are
// surfaced to the caller rather than best-effort swallowed.
func (c *Client) Send(ctx context.Context, s Submission) error {
	payload := request{
		ID: c.formID, Email: s.Email, Subject: s.Subject, Message: s.Message,
		Name: s.Name, Firstname: s.Firstname,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal contact request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/v1/contactreq", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build contact request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("cwcloud contact api is not available: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("cwcloud contact api returned status %d", resp.StatusCode)
	}
	return nil
}
