package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"cwclock-api/internal/ai"
	"cwclock-api/internal/assistant"
	"cwclock-api/internal/models"
)

// The CWAI external-adapter interface (ai-instruct-133). CWCloud's CWAI can use
// another platform's LLM access through an "external adapter": an HTTP endpoint
// that takes a prompt and answers it. This exposes the caller's cwclock AI
// subscriptions as such an adapter, so CWCloud can spend them.
//
// Contract: https://www.cwcloud.tech/docs/tutorials/cwai#external-adapter-interface
//
// It answers as the cwclock assistant, exactly as cwclock's own chat does: the
// same embedded skill, the same account context, and the same cwclock_api tool
// run through the same agent loop — so a question asked from CWCloud can read and
// generate the asker's timesheets, reports and invoices. What it does NOT keep is
// the thread: CWCloud owns that, so nothing is stored here, it sends whatever
// history it wants in `messages`, and `regenerate`/`enable_history` are its
// concern — they don't change the single answer produced here.
//
// A generated file can only be handed over as a link: the answer is plain text in
// somebody else's chat, so the system prompt tells the model to post the exact
// `downloadUrl` the tool returned (see adapterIntegrationRule).

// cwcloudAdapterMessage is one {role, message} turn in the multi-prompt shape.
type cwcloudAdapterMessage struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

// cwcloudAdapterPayload is either shape CWCloud sends: a single `message`, or a
// `messages` array. `adapter` names which subscription to spend.
type cwcloudAdapterPayload struct {
	Adapter       string                  `json:"adapter"`
	Settings      chatSettings            `json:"settings"`
	Message       string                  `json:"message"`
	Messages      []cwcloudAdapterMessage `json:"messages"`
	Regenerate    bool                    `json:"regenerate"`
	EnableHistory bool                    `json:"enable_history"`
}

// cwcloudAdapterUsage is the token count in the adapter's reply, named the way
// the CWAI contract expects.
type cwcloudAdapterUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

// cwcloudAdapterResponse is the adapter's reply (CWAI "Format 1": a single
// message plus usage).
type cwcloudAdapterResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Usage   cwcloudAdapterUsage `json:"usage"`
}

// CWCloudAdapter answers a CWCloud CWAI prompt as the cwclock assistant (skill,
// context and tools), spending one of the caller's AI subscriptions. POST
// /v1/ai/cwcloud/adapter, authenticated as the cwclock user whose subscription is
// spent (Bearer or X-Api-Key, like every /v1/ai route).
// The subscription is picked by the `adapter` field (matched against a
// subscription name as a contained, case-insensitive substring), falling back to
// the caller's default.
func (h *AIHandler) CWCloudAdapter(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	var p cwcloudAdapterPayload
	if !decodeAIBody(w, r, &p) {
		return
	}

	// Build the transcript: `system` turns are collected as the calling
	// integration's own instructions (appended to cwclock's system prompt below,
	// never replacing it), the rest are user/assistant turns, and a lone `message`
	// is one user turn.
	var system strings.Builder
	turns := make([]ai.Turn, 0, len(p.Messages)+1)
	for _, m := range p.Messages {
		text := strings.TrimSpace(m.Message)
		if text == "" {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(m.Role)) {
		case models.AIChatRoleSystem:
			if system.Len() > 0 {
				system.WriteString("\n\n")
			}
			system.WriteString(text)
		case models.AIChatRoleAssistant:
			turns = append(turns, ai.Turn{Role: models.AIChatRoleAssistant, Text: text})
		default:
			turns = append(turns, ai.Turn{Role: models.AIChatRoleUser, Text: text})
		}
	}
	if msg := strings.TrimSpace(p.Message); msg != "" {
		turns = append(turns, ai.Turn{Role: models.AIChatRoleUser, Text: msg})
	}
	// Regenerate: CWCloud wants the last answer redone, so drop trailing
	// assistant turns — the model re-answers the last question instead of being
	// handed its own previous reply as the final turn.
	if p.Regenerate {
		for len(turns) > 0 && turns[len(turns)-1].Role == models.AIChatRoleAssistant {
			turns = turns[:len(turns)-1]
		}
	}
	if len(turns) == 0 {
		writeError(w, http.StatusBadRequest, "Ask something", codeAIEmptyMessage)
		return
	}

	// The same system prompt cwclock's own chat answers under — the embedded
	// skill, who is asking and which organizations they can act in, the language
	// rule — plus what is particular to being called through the adapter. CWCloud
	// has no `lng`, so the language rule falls back to the user's latest message.
	connectionID := h.adapterConnectionID(r, userID, p.Adapter)
	answer, connection, ok := h.chatAnswer(w, r, userID, connectionID, ai.Request{
		System: adapterSystem(h.accountContext(r, userID, ""),
			normalizeLocale(r.URL.Query().Get("lng")), system.String()),
		Messages:  turns,
		MaxTokens: p.Settings.MaxTokens,
	})
	if !ok {
		return
	}

	// Record and mirror the spend exactly like Chat does, so an adapter call is
	// as visible in the trail and on Discord as a chat question.
	model := answeredWith(answer.Model, connection.Model)
	h.record(r, models.AuditAICall, models.AuditSubjectAIConnection, connection.ID, map[string]any{
		"provider":   connection.Provider,
		"model":      model,
		"input":      answer.Tokens.Input,
		"output":     answer.Tokens.Output,
		"cached":     answer.Tokens.Cached,
		"connection": connection.Name,
		"adapter":    true,
	})
	notifyDiscord(h.disc, fmt.Sprintf("[AI] :robot: %s prompted via the CWCloud adapter (%s / %s) — tokens consumed: %d input + %d output (%d cached)",
		actingEmail(r), connection.Provider, model, answer.Tokens.Input, answer.Tokens.Output, answer.Tokens.Cached))

	writeJSON(w, http.StatusOK, cwcloudAdapterResponse{
		Status:  "ok",
		Message: answer.Message,
		Usage: cwcloudAdapterUsage{
			PromptTokens:     answer.Tokens.Input,
			CompletionTokens: answer.Tokens.Output,
			TotalTokens:      answer.Tokens.Input + answer.Tokens.Output,
		},
	})
}

// adapterConnectionID resolves the `adapter` field to a subscription id: the
// caller's connection whose name contains it (case-insensitively) or whose id is
// it exactly, or "" to let the default (first) be used — matching the spec's "if
// it's not matching, take the first subscription".
func (h *AIHandler) adapterConnectionID(r *http.Request, userID, adapter string) string {
	if strings.TrimSpace(adapter) == "" {
		return ""
	}
	connections, err := h.connections.ListFor(r.Context(), userID)
	if err != nil {
		return ""
	}
	return pickAdapterConnection(connections, adapter)
}

// pickAdapterConnection is which subscription an `adapter` string names, out of
// the ones the caller may spend (ListFor order: their own before the ones shared
// with them, then position/created_at). The name is matched as a contained,
// case-insensitive substring — both sides trimmed — so an adapter registered in
// CWCloud as "claude" finds a subscription called "Claude (daily driver)".
//
// Several can match, so the tiers run from the most certain to the least: the id
// exactly, then a name that IS the adapter, then a name that merely contains it;
// within a tier the list's own order decides, which keeps "test" from picking
// "test (big)" and the result stable. Resolving in Go (not SQL) means a non-uuid
// adapter name never reaches a uuid cast. "" when nothing matches at all.
func pickAdapterConnection(connections []models.AIConnection, adapter string) string {
	adapter = strings.TrimSpace(adapter)
	if adapter == "" {
		return ""
	}
	needle := strings.ToLower(adapter)
	var exact, contained string
	for _, c := range connections {
		if c.ID == adapter {
			return c.ID
		}
		name := strings.ToLower(strings.TrimSpace(c.Name))
		switch {
		case name == needle:
			if exact == "" {
				exact = c.ID
			}
		case strings.Contains(name, needle):
			if contained == "" {
				contained = c.ID
			}
		}
	}
	if exact != "" {
		return exact
	}
	return contained
}

// adapterIntegrationRule is what the model has to know about being called
// through the adapter rather than from cwclock's own chat: the answer is plain
// text in somebody else's product, so a generated file can only be handed over as
// a link.
const adapterIntegrationRule = `# You are answering through an external integration

This question reached you through CWCloud's CWAI external adapter, so your answer
is rendered in another product's chat, not in cwclock's own screens. You are still
the cwclock assistant and you still have the cwclock_api tool: use it exactly as
you would in cwclock's own chat.

Nothing can be attached to this answer — it is plain text. So when a tool hands you
a ` + "`downloadUrl`" + ` for a generated file (an invoice, an exported report), give the user
that exact URL as a clickable Markdown link — [filename](downloadUrl) — and tell
them the link is temporary. Never say a file is attached, is being sent, or can be
found in the app, and never invent, alter or shorten a download URL: only ever use
one a tool returned. Do not point at a cwclock screen the reader cannot open from
here.`

// adapterCallerSystemHeading introduces whatever `system` turns CWCloud sent, so
// they read as the caller's instructions rather than as cwclock's own.
const adapterCallerSystemHeading = `# Instructions from the calling integration

The lines below were sent by the integration relaying this question. Follow them
where they do not contradict anything above; everything above takes precedence.`

// adapterSystem is the system prompt an adapter question is answered under: the
// same skill + account context + language rule cwclock's own chat builds, then
// the integration rule, and only then whatever `system` turns CWCloud sent —
// appended and explicitly subordinate, never replacing cwclock's own prompt (which
// is what the first version of this endpoint did).
func adapterSystem(context, locale, callerSystem string) string {
	parts := []string{assistant.Skill}
	if c := strings.TrimSpace(context); c != "" {
		parts = append(parts, c)
	}
	parts = append(parts, chatLanguageRule(locale), adapterIntegrationRule)
	if cs := strings.TrimSpace(callerSystem); cs != "" {
		parts = append(parts, adapterCallerSystemHeading+"\n\n"+cs)
	}
	return strings.Join(parts, "\n\n")
}
