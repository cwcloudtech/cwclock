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
	Cc         string      `json:"cc,omitempty"`
	Bcc        string      `json:"bcc,omitempty"`
	ReplyTo    string      `json:"reply_to,omitempty"`
	Subject    string      `json:"subject"`
	Content    string      `json:"content"`
	Attachment *Attachment `json:"attachment,omitempty"`
}

// Sender posts emails to CWCloud's email API (POST {apiURL}/v1/email).
type Sender struct {
	apiURL string
	apiKey string
	from   string
	// selfBaseURL is this CWClock API's own public base URL (cfg.APIBaseURL),
	// used to build the <img src> logo URLs emails reference (see logoURL) -
	// distinct from apiURL, which is CWCloud's email-sending API.
	selfBaseURL string
	client      *http.Client
}

// NewSender builds a Sender for the given CWCloud API base URL/key and
// From address. apiURL/apiKey are allowed to be blank - Send logs and skips
// rather than failing when they are. selfBaseURL is this API's own public
// base URL, used to build hosted logo URLs (see logoURL).
func NewSender(apiURL, apiKey, from, selfBaseURL string) *Sender {
	return &Sender{apiURL: apiURL, apiKey: apiKey, from: from, selfBaseURL: selfBaseURL, client: &http.Client{Timeout: 15 * time.Second}}
}

var bodyTemplate = template.Must(template.New("email").Parse(templates.EmailHTML))

// buttonStyle mirrors the frontend's primary .cw-button (index.css) so a
// CTA link in an email looks the same as one in the app: --cw-primary
// background, white text, --cw-radius-sm corners.
const buttonStyle = "display:inline-block;margin-top:8px;padding:9px 18px;" +
	"background-color:#1cb9f7;color:#ffffff;font-weight:600;" +
	"border-radius:6px;text-decoration:none;"

// mutedStyle mirrors --cw-text-muted, for secondary/help text under a CTA.
const mutedStyle = "color:#64748b;"

// centerStyle centers the paragraph wrapping a CTA button.
const centerStyle = "text-align:center;"

// renderBody wraps body in CWClock's shared email layout, with the header
// image pointed at logoURL (see Sender.logoURL).
func renderBody(title string, body template.HTML, logoURL string) (string, error) {
	var buf bytes.Buffer
	err := bodyTemplate.Execute(&buf, struct {
		Title string
		Logo  template.URL
		Body  template.HTML
	}{Title: title, Logo: template.URL(logoURL), Body: body})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// logoURL resolves the <img src> for an email's header to a stable HTTPS
// URL rather than an embedded data: URI. Data URIs were tried first, and
// even once correctly escaped (html/template defangs a data: URI spliced
// into a plain string src="..." into the literal text "#ZgotmplZ" unless
// it's given a trusted type) the logo still didn't render: many
// email-sending APIs and mail clients strip data: URIs from <img src>
// outright, a limitation no amount of correct HTML can work around. orgID
// selects that organization's public logo endpoint (see
// handlers.OrganizationHandler.PublicLogo, which falls back to the default
// CWClock logo itself when the org has no usable picture) - pass "" to
// always get the default logo (see handlers.AssetsLogo).
func (s *Sender) logoURL(orgID string) string {
	if utils.IsBlank(orgID) {
		return s.selfBaseURL + "/v1/assets/logo.png"
	}
	return s.selfBaseURL + "/v1/organizations/" + orgID + "/logo"
}

// send posts one email best-effort: a blank apiURL/apiKey or a failed
// request is logged (with the payload, so it can be replayed by hand) and
// otherwise ignored. cc/replyTo are optional - pass "" to leave either unset.
func (s *Sender) send(ctx context.Context, to, cc, replyTo, subject, htmlContent string, attachment *Attachment) {
	payload := request{From: s.from, To: to, Cc: cc, ReplyTo: replyTo, Subject: subject, Content: htmlContent, Attachment: attachment}

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
		`<p>Welcome to CWClock!</p><p>Please confirm your account by clicking the button below:</p>`+
			`<p style="%s"><a href="%s" style="%s">Confirm my account</a></p>`+
			`<p style="%s">If you didn't create this account, you can safely ignore this email.</p>`,
		centerStyle, template.HTMLEscapeString(confirmURL), buttonStyle, mutedStyle,
	))
	html, err := renderBody("Confirm your CWClock account", body, s.logoURL(utils.EMPTY))
	if err != nil {
		slog.Error("failed to render confirmation email", "error", err)
		return
	}
	s.send(ctx, to, utils.EMPTY, utils.EMPTY, "Confirm your CWClock account", html, nil)
}

// SendPasswordReset emails the password-renewal link to a user who
// requested one.
func (s *Sender) SendPasswordReset(ctx context.Context, to, resetURL string) {
	body := template.HTML(fmt.Sprintf(
		`<p>We received a request to reset your CWClock password.</p>`+
			`<p style="%s"><a href="%s" style="%s">Choose a new password</a></p>`+
			`<p style="%s">If you didn't request this, you can safely ignore this email.</p>`,
		centerStyle, template.HTMLEscapeString(resetURL), buttonStyle, mutedStyle,
	))
	html, err := renderBody("Reset your CWClock password", body, s.logoURL(utils.EMPTY))
	if err != nil {
		slog.Error("failed to render password reset email", "error", err)
		return
	}
	s.send(ctx, to, utils.EMPTY, utils.EMPTY, "Reset your CWClock password", html, nil)
}

// formatUSDate formats a "2006-01-02" day string as CWClock's US display
// date (01/02/2006), matching the invoice PDF (see report.formatUSDate).
// Returns day unchanged if it doesn't parse.
func formatUSDate(day string) string {
	d, err := time.Parse("2006-01-02", day)
	if err != nil {
		return day
	}
	return d.Format("01/02/2006")
}

// SendInvoice emails a generated invoice PDF to one or more recipients. The
// organization's own avatar (see handlers.OrganizationHandler.PublicLogo)
// replaces the CWClock logo in the email header when it has one. ownerEmail
// (the organization owner's email) is set as replyTo, so a reply from the
// client reaches them directly rather than the noreply From address, and -
// along with accountingEmail, the organization's optional accounting
// department address, when set - always cc'd so a copy of every invoice
// sent to a client reaches them too. startDay/endDay are the invoice's
// billed period ("2006-01-02"), shown in parentheses in the subject/title.
// language is the client_language decision table's result (models.
// ClientLanguage) - "fr" sends the email in French, anything else in
// English.
func (s *Sender) SendInvoice(ctx context.Context, recipients []string, orgID, orgName, ownerEmail, accountingEmail, invoiceNumber, startDay, endDay, language string, pdf []byte) {
	if len(recipients) == 0 {
		return
	}

	cc := make([]string, 0, 2)
	if utils.IsNotBlank(ownerEmail) {
		cc = append(cc, ownerEmail)
	}

	if utils.IsNotBlank(accountingEmail) {
		cc = append(cc, accountingEmail)
	}

	period := fmt.Sprintf("%s - %s", formatUSDate(startDay), formatUSDate(endDay))

	var subject, title string
	var body template.HTML
	if language == "fr" {
		subject = fmt.Sprintf("Facture %s (%s)", invoiceNumber, period)
		title = fmt.Sprintf("Votre facture %s de %s", invoiceNumber, orgName)
		body = template.HTML(fmt.Sprintf(
			`<p>Veuillez trouver ci-joint la facture <strong>%s</strong> de <strong>%s</strong> (%s).</p>`,
			template.HTMLEscapeString(invoiceNumber), template.HTMLEscapeString(orgName), template.HTMLEscapeString(period),
		))
	} else {
		subject = fmt.Sprintf("Invoice %s (%s)", invoiceNumber, period)
		title = fmt.Sprintf("Your invoice %s from %s", invoiceNumber, orgName)
		body = template.HTML(fmt.Sprintf(
			`<p>Please find attached invoice <strong>%s</strong> from <strong>%s</strong> (%s).</p>`,
			template.HTMLEscapeString(invoiceNumber), template.HTMLEscapeString(orgName), template.HTMLEscapeString(period),
		))
	}

	html, err := renderBody(title, body, s.logoURL(orgID))
	if err != nil {
		slog.Error("failed to render invoice email", "error", err)
		return
	}
	attachment := &Attachment{
		MimeType: "application/pdf",
		FileName: invoiceNumber + ".pdf",
		B64:      base64.StdEncoding.EncodeToString(pdf),
	}
	s.send(ctx, strings.Join(recipients, ","), strings.Join(cc, ","), ownerEmail, subject, html, attachment)
}
