package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cwclock-api/internal/ai"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

// The CWAI external-adapter interface (ai-instruct-133). CWCloud's CWAI can use
// another platform's LLM access through an "external adapter": an HTTP endpoint
// that takes a prompt and answers it. This exposes the caller's cwclock AI
// subscriptions as such an adapter, so CWCloud can spend them.
//
// Contract: https://www.cwcloud.tech/docs/tutorials/cwai#external-adapter-interface
//
// It is a plain LLM gateway, NOT the cwclock assistant: no timesheet skill, no
// cwclock_api tool, no stored conversation. CWCloud owns the thread, so it sends
// whatever history it wants in `messages` and `regenerate`/`enable_history` are
// its concern — they don't change the single answer produced here.

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

// CWCloudAdapter answers a CWCloud CWAI prompt through one of the caller's AI
// subscriptions. POST /v1/ai/cwcloud/adapter, authenticated as the cwclock user
// whose subscription is spent (Bearer or X-Api-Key, like every /v1/ai route).
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

	// Build the transcript: `system` turns feed the Request.System instruction,
	// the rest are user/assistant turns, and a lone `message` is one user turn.
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

	connectionID := h.adapterConnectionID(r, userID, p.Adapter)
	answer, connection, ok := h.adapterAnswer(w, r, userID, connectionID, ai.Request{
		System:    system.String(),
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

// adapterAnswer resolves the subscription and makes one plain model call (no
// tools, no agent loop) — the adapter is a gateway, not the assistant.
func (h *AIHandler) adapterAnswer(w http.ResponseWriter, r *http.Request, userID, connectionID string,
	request ai.Request) (ai.Response, models.AIConnection, bool) {
	connection, err := h.usable(r, userID, connectionID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound) && strings.TrimSpace(connectionID) != "":
			writeError(w, http.StatusNotFound, "Connection not found", CodeNotFound)
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusPreconditionRequired, "Add an AI subscription first", codeAINotConfigured)
		default:
			writeStoreError(w, err)
		}
		return ai.Response{}, models.AIConnection{}, false
	}
	answer, err := h.client.Ask(r.Context(), connection, request)
	if err != nil {
		h.writeProviderError(w, connection, err)
		return ai.Response{}, models.AIConnection{}, false
	}
	return answer, connection, true
}
