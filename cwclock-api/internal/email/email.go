// Package email sends CWClock's transactional emails (account
// confirmation, password reset, invoice delivery) through CWCloud's email
// API. It is deliberately best-effort throughout: a missing configuration
// or an unreachable CWCloud API is logged, never returned as an error, so
// the caller's own flow (registration, invoice generation, ...) never fails
// because of it.
package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"cwclock-api/internal/assets"
	"cwclock-api/internal/templates"
	"cwclock-api/internal/utils"
)

// Attachment is an optional file joined to an email, base64-encoded.
type Attachment struct {
	MimeType string `json:"mime_type"`
	FileName string `json:"file_name"`
	B64      string `json:"b64"`
}

type request struct {
	From       string      `json:"from"`
	To         string      `json:"to"`
	Bcc        string      `json:"bcc,omitempty"`
	Subject    string      `json:"subject"`
	Content    string      `json:"content"`
	Attachment *Attachment `json:"attachment,omitempty"`
}

// Sender posts emails to CWCloud's email API (POST {apiURL}/v1/email).
type Sender struct {
	apiURL string
	apiKey string
	from   string
	client *http.Client
}

// NewSender builds a Sender for the given CWCloud API base URL/key and
// From address. apiURL/apiKey are allowed to be blank - Send logs and skips
// rather than failing when they are.
func NewSender(apiURL, apiKey, from string) *Sender {
	return &Sender{apiURL: apiURL, apiKey: apiKey, from: from, client: &http.Client{Timeout: 15 * time.Second}}
}

var bodyTemplate = template.Must(template.New("email").Parse(templates.EmailHTML))

// renderBody wraps body in CWClock's shared email layout, with the
// CWClock logo or, when logoOverride is a data URI (an organization's own
// avatar), that image instead.
func renderBody(title string, body template.HTML, logoOverride string) (string, error) {
	logo := logoDataURI(logoOverride)
	var buf bytes.Buffer
	err := bodyTemplate.Execute(&buf, struct {
		Title string
		Logo  string
		Body  template.HTML
	}{Title: title, Logo: logo, Body: body})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// logoDataURI returns override as-is when it already looks like a data URI
// (an organization's uploaded avatar), otherwise the bundled CWClock logo
// encoded as one.
func logoDataURI(override string) string {
	if strings.HasPrefix(override, "data:") {
		return override
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(assets.CWClockLogoPNG)
}

// send posts one email best-effort: a blank apiURL/apiKey or a failed
// request is logged (with the payload, so it can be replayed by hand) and
// otherwise ignored.
func (s *Sender) send(ctx context.Context, to, subject, htmlContent string, attachment *Attachment) {
	payload := request{From: s.from, To: to, Subject: subject, Content: htmlContent, Attachment: attachment}

	if utils.IsBlank(s.apiURL) || utils.IsBlank(s.apiKey) {
		slog.Warn("cwcloud email api is not configured (CWCLOUD_API_URL/CWCLOUD_API_KEY), skipping email", "to", to, "subject", subject)
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal email payload", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+"/v1/email", bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to build cwcloud email request", "error", err)
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		slog.Error("cwcloud email api is not available", "error", err, "to", to, "subject", subject)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		slog.Error("cwcloud email api returned an error", "status", resp.StatusCode, "to", to, "subject", subject)
	}
}

// SendConfirmation emails the account-confirmation link to a newly
// registered user (activation mode "email").
func (s *Sender) SendConfirmation(ctx context.Context, to, confirmURL string) {
	body := template.HTML(fmt.Sprintf(
		`<p>Welcome to CWClock!</p><p>Please confirm your account by clicking the link below:</p>`+
			`<p><a href="%s">Confirm my account</a></p>`+
			`<p>If you didn't create this account, you can safely ignore this email.</p>`,
		template.HTMLEscapeString(confirmURL),
	))
	html, err := renderBody("Confirm your CWClock account", body, utils.EMPTY)
	if err != nil {
		slog.Error("failed to render confirmation email", "error", err)
		return
	}
	s.send(ctx, to, "Confirm your CWClock account", html, nil)
}

// SendPasswordReset emails the password-renewal link to a user who
// requested one.
func (s *Sender) SendPasswordReset(ctx context.Context, to, resetURL string) {
	body := template.HTML(fmt.Sprintf(
		`<p>We received a request to reset your CWClock password.</p>`+
			`<p><a href="%s">Choose a new password</a></p>`+
			`<p>If you didn't request this, you can safely ignore this email.</p>`,
		template.HTMLEscapeString(resetURL),
	))
	html, err := renderBody("Reset your CWClock password", body, utils.EMPTY)
	if err != nil {
		slog.Error("failed to render password reset email", "error", err)
		return
	}
	s.send(ctx, to, "Reset your CWClock password", html, nil)
}

// SendInvoice emails a generated invoice PDF to one or more recipients. The
// organization's avatar (orgPicture, a data URI) replaces the CWClock logo
// in the email header when it's set.
func (s *Sender) SendInvoice(ctx context.Context, recipients []string, orgName, orgPicture, invoiceNumber string, pdf []byte) {
	if len(recipients) == 0 {
		return
	}
	body := template.HTML(fmt.Sprintf(
		`<p>Please find attached invoice <strong>%s</strong> from %s.</p>`,
		template.HTMLEscapeString(invoiceNumber), template.HTMLEscapeString(orgName),
	))
	html, err := renderBody("Your invoice from "+orgName, body, orgPicture)
	if err != nil {
		slog.Error("failed to render invoice email", "error", err)
		return
	}
	attachment := &Attachment{
		MimeType: "application/pdf",
		FileName: invoiceNumber + ".pdf",
		B64:      base64.StdEncoding.EncodeToString(pdf),
	}
	s.send(ctx, strings.Join(recipients, ","), "Invoice "+invoiceNumber, html, attachment)
}
