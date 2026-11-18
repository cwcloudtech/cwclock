// Package notify is the Discord alert transport, ported from the uprodit/sf-api
// implementation: Slack-format webhook POSTs to
// https://discord.com/api/webhooks/{token}/slack. It is disabled (every call a
// no-op) when the token is blank or the "changeit" sentinel, so a deployment
// without a real webhook silently sends nothing.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	disableSentinel = "changeit"
	webhookTpl      = "https://discord.com/api/webhooks/%s/slack"
	// limitSize is the Discord message length cap; longer notices are truncated.
	limitSize = 2000
	// infoBlue is the attachment colour for every notice - the "info" blue the
	// comwork tooling uses, so the alerts read as one family.
	infoBlue = "#95C8F3"
)

// Discord posts Slack-format messages to a Discord webhook.
type Discord struct {
	token    string
	disabled bool
	http     *http.Client
}

// NewDiscord builds the transport from the webhook token (DISCORD_TOKEN).
func NewDiscord(token string) *Discord {
	return &Discord{
		token:    token,
		disabled: token == "" || strings.Contains(token, disableSentinel),
		http:     &http.Client{Timeout: 15 * time.Second},
	}
}

// Enabled reports whether the transport will actually call Discord.
func (d *Discord) Enabled() bool { return d != nil && !d.disabled }

// SendText posts a notice as a coloured attachment: the "[TAG] :emoji:" prefix
// becomes the attachment title and the rest becomes the body. Truncated to the
// Discord limit. A no-op when disabled, and safe on a nil receiver; errors are
// swallowed by callers that treat alerting as best-effort.
func (d *Discord) SendText(ctx context.Context, text string) error {
	if d == nil || d.disabled {
		return nil
	}
	if len(text) > limitSize {
		text = text[:limitSize]
	}
	title, bodyText := splitNotice(text)
	payload := map[string]any{
		"attachments": []map[string]any{{
			"color": infoBlue,
			"title": title,
			"text":  bodyText,
		}},
	}
	body, _ := json.Marshal(payload)
	return d.post(ctx, body)
}

// splitNotice turns a "[TAG] :emoji: rest" line into an attachment title
// ("TAG :emoji:") and body ("rest"). A line that doesn't open with a [tag] has
// no title and becomes the body as-is.
func splitNotice(text string) (title, body string) {
	s := strings.TrimSpace(text)
	if !strings.HasPrefix(s, "[") {
		return "", s
	}
	closeIdx := strings.IndexByte(s, ']')
	if closeIdx < 0 {
		return "", s
	}
	tag := strings.TrimSpace(s[1:closeIdx])
	rest := strings.TrimSpace(s[closeIdx+1:])
	// An optional :emoji: right after the tag joins it in the title.
	if strings.HasPrefix(rest, ":") {
		if end := strings.IndexByte(rest[1:], ':'); end >= 0 {
			tag += " " + rest[:end+2]
			rest = strings.TrimSpace(rest[end+2:])
		}
	}
	return tag, rest
}

func (d *Discord) post(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(webhookTpl, d.token), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.http.Do(req)
	if err != nil {
		slog.Default().Error("discord webhook transport error", "err", err)
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	// The webhook URL carries the secret token, so it is deliberately NOT logged.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Default().Error("discord webhook response", "status", resp.StatusCode, "body", string(b))
		return fmt.Errorf("discord webhook responded %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
