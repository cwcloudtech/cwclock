package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/ai"
	"cwclock-api/internal/aidownload"
	"cwclock-api/internal/assistant"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/notify"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

// i18n error codes the AI endpoints return (the frontends key their messages/
// empty-states on these). errors.aiNotConfigured is the one the chat UI watches
// for: a 428 with it means the caller has no usable connection, so the chat shows
// its "add a subscription first" state.
const (
	codeAINotConfigured = "errors.aiNotConfigured"
	codeAIRefused       = "errors.aiProviderRefused"
	codeAIUpstream      = "errors.aiUpstream"
	codeAIKeyRequired   = "errors.aiKeyRequired"
	codeAIProvider      = "errors.aiInvalidProvider"
	codeAIBaseURL       = "errors.aiInvalidBaseUrl"
	codeAIRole          = "errors.aiInvalidRole"
	codeAIEmptyMessage  = "errors.aiEmptyMessage"
	codeAIShareTarget   = "errors.aiShareTarget"
)

// AIHandler serves the AI assistant (ai-instruct-132): a user's connections
// (subscriptions) to a language model, who else may spend them, and the questions
// they put through them with the cwclock API offered as a tool.
//
// A connection is a shared resource: an account holds a credential and decides who
// else may spend it. Editing, deleting and sharing a connection are the owner's
// alone (a non-owner gets 404, never 403 — whether somebody else's connection
// exists is not a question this answers). The credential is the user's own stored
// token; there is no global provider key.
type AIHandler struct {
	auditor
	connections *store.AIStore
	users       *store.UserStore
	orgs        *store.OrgStore
	// apiBaseURL is the loopback base the cwclock_api tool calls this very API with
	// (e.g. http://127.0.0.1:8080), so a self-call never leaves the host.
	apiBaseURL string
	client     *ai.Client
	// downloads is where a generated report/invoice PDF is parked so the agent can
	// hand the user a download link. Nil disables that (the tool then just names the
	// generated file, as before).
	downloads *aidownload.Store
	// disc posts an [AI] alert per chat question with the token spend, so the running
	// cost is visible without opening the audit trail. Optional and best-effort (a
	// nil/disabled transport is a no-op and never affects the answer).
	disc *notify.Discord
}

// NewAIHandler builds the /v1/ai handler. users resolves a share target and its
// account; orgs feeds the account context the chat injects so the model knows which
// organization ids to target; apiBaseURL backs the cwclock_api tool; downloads (may
// be nil) parks generated report/invoice PDFs so the agent can share a link; audit
// records each chat question's token spend in the superadmin trail.
func NewAIHandler(connections *store.AIStore, users *store.UserStore, orgs *store.OrgStore, apiBaseURL string, downloads *aidownload.Store, audit *store.AuditStore) *AIHandler {
	return &AIHandler{auditor: newAuditor(audit), connections: connections, users: users, orgs: orgs, apiBaseURL: apiBaseURL, client: ai.New(), downloads: downloads}
}

// WithNotifier wires the Discord transport so each chat question posts an [AI]
// alert naming the asker, the provider/model and the tokens it spent. Optional
// and best-effort (a nil/disabled transport never affects the answer).
func (h *AIHandler) WithNotifier(disc *notify.Discord) *AIHandler {
	h.disc = disc
	return h
}

func aiUserID(r *http.Request) (string, bool) {
	id, ok := middleware.UserIDFromContext(r.Context())
	return id, ok && utils.IsNotBlank(id)
}

func decodeAIBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return false
	}
	return true
}

// --- connections CRUD ---

type aiConnectionPayload struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	BaseURL   string `json:"baseUrl"`
	Model     string `json:"model"`
	MaxTokens int    `json:"maxTokens"`
	Token     string `json:"token"`
}

// normalize fills in what the user left blank and refuses what cannot work. The URL
// and the model come from the provider catalogue when left alone; both stay
// editable because pointing a connection at a local gateway is why the fields exist.
func (h *AIHandler) normalize(w http.ResponseWriter, r *http.Request, p aiConnectionPayload) (models.AIConnection, bool) {
	provider, err := h.connections.FindProvider(r.Context(), strings.ToLower(strings.TrimSpace(p.Provider)))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusBadRequest, "Pick a provider", codeAIProvider)
			return models.AIConnection{}, false
		}
		writeStoreError(w, err)
		return models.AIConnection{}, false
	}

	baseURL := strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
	if baseURL == "" {
		baseURL = provider.BaseURL
	}
	if !strings.HasPrefix(baseURL, "https://") && !strings.HasPrefix(baseURL, "http://") {
		writeError(w, http.StatusBadRequest, "The address must be an http or https URL", codeAIBaseURL)
		return models.AIConnection{}, false
	}

	model := strings.TrimSpace(p.Model)
	if model == "" {
		model = provider.DefaultModel
	}

	maxTokens := p.MaxTokens
	if maxTokens <= 0 {
		maxTokens = models.DefaultMaxTokens
	}
	if maxTokens > models.MaxMaxTokens {
		maxTokens = models.MaxMaxTokens
	}

	return models.AIConnection{
		Name: strings.TrimSpace(p.Name), Provider: provider.Name,
		BaseURL: baseURL, Model: model, MaxTokens: maxTokens,
		Token: strings.TrimSpace(p.Token),
	}, true
}

// List is every connection the caller may spend, their own first, without tokens —
// what the settings screen renders, and what the chat nav item is gated on (>=1 row).
func (h *AIHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	connections, err := h.connections.ListFor(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, connections)
}

// Providers is the catalogue: what this deployment can talk to, the address and
// model each defaults to, and the models worth suggesting.
func (h *AIHandler) Providers(w http.ResponseWriter, r *http.Request) {
	providers, err := h.connections.ListProviders(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, providers)
}

// Create stores a new connection. A key is required here (unlike an update).
func (h *AIHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	var p aiConnectionPayload
	if !decodeAIBody(w, r, &p) {
		return
	}
	connection, ok := h.normalize(w, r, p)
	if !ok {
		return
	}
	if strings.TrimSpace(connection.Token) == "" {
		writeError(w, http.StatusBadRequest, "A key is required", codeAIKeyRequired)
		return
	}

	created, err := h.connections.Create(r.Context(), userID, connection)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	created.Token = ""
	writeJSON(w, http.StatusCreated, created)
}

// Update replaces an owned connection's settings. A blank token keeps the stored one.
func (h *AIHandler) Update(w http.ResponseWriter, r *http.Request) {
	owned, ok := h.ownConnection(w, r)
	if !ok {
		return
	}
	var p aiConnectionPayload
	if !decodeAIBody(w, r, &p) {
		return
	}
	connection, ok := h.normalize(w, r, p)
	if !ok {
		return
	}
	updated, err := h.connections.Update(r.Context(), owned.ID, connection)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	updated.Token = ""
	writeJSON(w, http.StatusOK, updated)
}

type aiOrderPayload struct {
	ConnectionIDs []string `json:"connectionIds"`
}

// Reorder sets which connection is first — the default, and the first row of every
// "ask with" menu. Returns the list in its new order.
func (h *AIHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	var p aiOrderPayload
	if !decodeAIBody(w, r, &p) {
		return
	}
	if err := h.connections.Reorder(r.Context(), userID, p.ConnectionIDs); err != nil {
		writeStoreError(w, err)
		return
	}
	connections, err := h.connections.ListFor(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, connections)
}

// Delete removes an owned connection.
func (h *AIHandler) Delete(w http.ResponseWriter, r *http.Request) {
	owned, ok := h.ownConnection(w, r)
	if !ok {
		return
	}
	if err := h.connections.Delete(r.Context(), owned.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": owned.ID})
}

// ownConnection resolves the addressed connection and refuses anybody but its owner
// with a 404. Shared access is the right to spend a connection, never to edit it.
func (h *AIHandler) ownConnection(w http.ResponseWriter, r *http.Request) (models.AIConnection, bool) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return models.AIConnection{}, false
	}
	connection, err := h.connections.FindByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return models.AIConnection{}, false
	}
	if connection.OwnerID != userID {
		writeError(w, http.StatusNotFound, "Connection not found", CodeNotFound)
		return models.AIConnection{}, false
	}
	return connection, true
}

// --- sharing ---

// ListShares is everybody an owner has shared this connection with.
func (h *AIHandler) ListShares(w http.ResponseWriter, r *http.Request) {
	owned, ok := h.ownConnection(w, r)
	if !ok {
		return
	}
	grants, err := h.connections.ListGrants(r.Context(), owned.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, grants)
}

type aiGrantPayload struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Share lets another user put questions through this connection. The target is given
// by email (resolved to their account), which is friendlier than an opaque id.
func (h *AIHandler) Share(w http.ResponseWriter, r *http.Request) {
	owned, ok := h.ownConnection(w, r)
	if !ok {
		return
	}
	var p aiGrantPayload
	if !decodeAIBody(w, r, &p) {
		return
	}
	role := strings.TrimSpace(p.Role)
	if role == "" {
		role = models.AIRoleUser
	}
	if !models.IsValidAIRole(role) {
		writeError(w, http.StatusBadRequest, "Invalid role", codeAIRole)
		return
	}

	email := strings.ToLower(strings.TrimSpace(p.Email))
	if email == "" {
		writeError(w, http.StatusBadRequest, "Enter the email of the user to share with", codeAIShareTarget)
		return
	}
	target, err := h.users.FindByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusBadRequest, "No user with this email", codeAIShareTarget)
			return
		}
		writeStoreError(w, err)
		return
	}
	if target.ID == owned.OwnerID {
		writeError(w, http.StatusBadRequest, "This connection is already yours", codeAIShareTarget)
		return
	}

	if err := h.connections.Grant(r.Context(), owned.ID, target.ID, role); err != nil {
		writeStoreError(w, err)
		return
	}
	grants, err := h.connections.ListGrants(r.Context(), owned.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, grants)
}

// Unshare revokes a user's access.
func (h *AIHandler) Unshare(w http.ResponseWriter, r *http.Request) {
	owned, ok := h.ownConnection(w, r)
	if !ok {
		return
	}
	targetID := chi.URLParam(r, "userId")
	if utils.IsBlank(targetID) {
		writeError(w, http.StatusBadRequest, "Invalid user id", CodeInvalidRequestBody)
		return
	}
	if err := h.connections.Revoke(r.Context(), owned.ID, targetID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"userId": targetID})
}

// aiShareableContact is one autocomplete suggestion for the share picker: an
// email the caller is allowed to share a subscription with, plus a display name.
type aiShareableContact struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

// ShareableEmails powers the share picker's autocomplete, restricted to the people
// the caller actually knows (ai-instruct-132): a superuser may share with ANY
// account, so it searches every user; anyone else may only share with the members
// of the organizations THEY own, so it searches those. Free text is still accepted
// by Share itself — this only bounds the suggestions.
func (h *AIHandler) ShareableEmails(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	contacts := []aiShareableContact{}
	if user.Role == models.GlobalRoleSuperuser {
		// An admin can reach everyone, but only once they've typed something — an
		// empty term must not dump the whole user table.
		if query == "" {
			writeJSON(w, http.StatusOK, contacts)
			return
		}
		users, err := h.users.SearchByEmail(r.Context(), query, 10)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, u := range users {
			if u.ID == userID {
				continue
			}
			contacts = append(contacts, aiShareableContact{Email: u.Email, Name: u.Name, Surname: u.Surname})
		}
	} else {
		members, err := h.orgs.ShareableMembers(r.Context(), userID, query, 20)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, m := range members {
			contacts = append(contacts, aiShareableContact{Email: m.Email, Name: m.Name, Surname: m.Surname})
		}
	}
	writeJSON(w, http.StatusOK, contacts)
}

// --- asking ---

type chatSettings struct {
	MaxTokens int `json:"max_tokens"`
}

type chatPayload struct {
	Settings chatSettings `json:"settings"`
	Message  string       `json:"message"`
	// Messages is a client-held thread, used only when ConversationID is blank.
	Messages []models.AIMessage `json:"messages"`
	// ConversationID continues a stored thread; blank starts one.
	ConversationID string `json:"conversationId"`
	// ConnectionID is which connection to spend; blank is the caller's default.
	ConnectionID string `json:"connectionId"`
	// OrgID is the organization the user is currently looking at, injected into the
	// assistant's context so it can target /v1/organizations/{orgId}/... paths.
	OrgID string `json:"orgId"`
}

type chatResponse struct {
	Status         string         `json:"status"`
	Message        string         `json:"message"`
	Usage          map[string]any `json:"usage"`
	Model          string         `json:"model"`
	ConversationID string         `json:"conversationId"`
	Title          string         `json:"title"`
	// DownloadTTLSeconds lets the chat show a download link as expired once the
	// message carrying it is older than this (0 / omitted when downloads are off).
	DownloadTTLSeconds int `json:"downloadTtlSeconds,omitempty"`
}

// downloadTTLSeconds is the lifetime of a generated-file download link, in whole
// seconds, or 0 when downloads aren't wired.
func (h *AIHandler) downloadTTLSeconds() int {
	if h.downloads == nil {
		return 0
	}
	return int(h.downloads.TTL().Seconds())
}

// maxChatHistory bounds how much of a conversation is sent back to the model.
const maxChatHistory = 20

// Chat answers a question as the cwclock assistant and keeps the exchange. The
// connection spent is the caller's own, or the one they named, or the first shared
// with them. Every question and answer is stored so a chat survives a reload.
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	var p chatPayload
	if !decodeAIBody(w, r, &p) {
		return
	}

	var conversation models.AIConversation
	history := p.Messages
	if strings.TrimSpace(p.ConversationID) != "" {
		stored, err := h.connections.FindConversation(r.Context(), p.ConversationID, userID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		conversation = stored
		history = historyOf(stored)
	}

	question := strings.TrimSpace(p.Message)
	if question != "" {
		history = append(history, models.AIMessage{Role: models.AIChatRoleUser, Message: question})
	}
	if len(history) == 0 {
		writeError(w, http.StatusBadRequest, "Ask something", codeAIEmptyMessage)
		return
	}
	if question == "" {
		question = lastQuestion(history)
	}
	if len(history) > maxChatHistory {
		history = history[len(history)-maxChatHistory:]
	}

	// The skill, then who is asking and which organizations they can act in, then
	// the language. The model is given the cwclock_api tool and may call it before
	// it answers (the agent loop below).
	locale := normalizeLocale(r.URL.Query().Get("lng"))
	context := h.accountContext(r, userID, strings.TrimSpace(p.OrgID))
	answer, connection, ok := h.chatAnswer(w, r, userID, p.ConnectionID, ai.Request{
		System:    assistant.Skill + "\n\n" + context + "\n\n" + chatLanguageRule(locale),
		Messages:  turnsOf(history),
		MaxTokens: p.Settings.MaxTokens,
	})
	if !ok {
		return
	}

	if conversation.ID == "" {
		id, err := h.connections.CreateConversation(r.Context(), userID, models.AIConversationTitle(question))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		conversation.ID, conversation.Title = id, models.AIConversationTitle(question)
	}

	if err := h.connections.AppendExchange(r.Context(), conversation.ID, models.AIPrompt{
		Question: question, Answer: answer.Message,
		Provider: connection.Provider, Model: answeredWith(answer.Model, connection.Model),
		Connection: connection.Name, Usage: answer.Usage,
	}); err != nil {
		writeStoreError(w, err)
		return
	}

	// One entry per question: the agent loop may have called the provider
	// several times, but answer carries the totals it settled on, and that is
	// also what the conversation stores. Recorded once the exchange is
	// persisted, so the trail never names a call the user's history lacks.
	h.record(r, models.AuditAICall, models.AuditSubjectAIConnection, connection.ID, map[string]any{
		"provider":   connection.Provider,
		"model":      answeredWith(answer.Model, connection.Model),
		"input":      answer.Tokens.Input,
		"output":     answer.Tokens.Output,
		"cached":     answer.Tokens.Cached,
		"connection": connection.Name,
	})

	// Mirror the token spend to Discord so the running cost is visible without
	// opening the audit trail: same totals the audit entry and the conversation
	// carry. Best-effort and off the request path.
	notifyDiscord(h.disc, fmt.Sprintf("[AI] :robot: %s prompted the assistant (%s / %s) — tokens consumed: %d input + %d output (%d cached)",
		actingEmail(r), connection.Provider, answeredWith(answer.Model, connection.Model),
		answer.Tokens.Input, answer.Tokens.Output, answer.Tokens.Cached))

	writeJSON(w, http.StatusOK, chatResponse{
		Status: "ok", Message: answer.Message, Usage: answer.Usage, Model: answer.Model,
		ConversationID: conversation.ID, Title: conversation.Title,
		DownloadTTLSeconds: h.downloadTTLSeconds(),
	})
}

// usable is the connection a question goes through: the one asked for, or the
// caller's default.
func (h *AIHandler) usable(r *http.Request, userID, connectionID string) (models.AIConnection, error) {
	if strings.TrimSpace(connectionID) == "" {
		return h.connections.Usable(r.Context(), userID)
	}
	return h.connections.UsableByID(r.Context(), userID, strings.TrimSpace(connectionID))
}

// maxToolRounds bounds the agent loop: how many times the model may call tools
// before it must answer. Each round is one more provider call, so it is small. After
// the cap one final call is made with tools withheld, so the loop ends in prose.
const maxToolRounds = 4

// noToolsRule is appended to the system prompt for a format that cannot call
// tools (CWCloud/CWAI). The skill tells the assistant to read the cwclock API
// before answering; with no tools to do that, this stops it inventing figures.
const noToolsRule = "IMPORTANT: for THIS answer you have NO tools. You cannot call the cwclock API, " +
	"read the user's clients, projects, time entries or invoices, or generate or download any file. " +
	"Answer only from this conversation and the account context above. Never state specific figures " +
	"(hours, totals, amounts, invoice numbers, dates) as if you had looked them up — if the answer needs " +
	"data you cannot reach, say so plainly and point the user to the cwclock screen that holds it (the time " +
	"tracker, the reports screen, or the invoices screen). Never claim a report or invoice was generated, " +
	"and never offer a download link."

// chatAnswer runs the question through the model as an agent: it offers the
// cwclock_api tool, and whenever the model answers with a tool call rather than
// prose it executes the tool, feeds the result back and asks again, up to
// maxToolRounds times.
func (h *AIHandler) chatAnswer(w http.ResponseWriter, r *http.Request, userID, connectionID string,
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

	// A format that cannot do tools (CWCloud/CWAI) is asked once, with the tool
	// catalogue withheld AND a rule in the system prompt saying so — otherwise an
	// assistant told at length to read the API before answering reports figures
	// it never read.
	if !models.AIFormatSupportsTools(connection.Format) {
		request.Tools = nil
		request.System = strings.TrimRight(request.System, "\n") + "\n\n" + noToolsRule
		answer, err := h.client.Ask(r.Context(), connection, request)
		if err != nil {
			h.writeProviderError(w, connection, err)
			return ai.Response{}, models.AIConnection{}, false
		}
		return answer, connection, true
	}

	request.Tools = h.aiTools()
	for round := 0; round < maxToolRounds; round++ {
		answer, err := h.client.Ask(r.Context(), connection, request)
		if err != nil {
			h.writeProviderError(w, connection, err)
			return ai.Response{}, models.AIConnection{}, false
		}
		if len(answer.ToolCalls) == 0 {
			return answer, connection, true
		}
		request.Messages = append(request.Messages,
			ai.Turn{Role: models.AIChatRoleAssistant, Text: answer.Message, ToolCalls: answer.ToolCalls})
		for _, call := range answer.ToolCalls {
			slog.Info("ai: tool call", "tool", call.Name, "provider", connection.Provider)
			request.Messages = append(request.Messages, ai.Turn{
				Role: ai.RoleTool, ToolCallID: call.ID, ToolName: call.Name,
				Text: h.runTool(r, call),
			})
		}
	}

	// Rounds exhausted: one last call with the tools withheld, so the model must turn
	// what it has gathered into an answer rather than keep searching.
	request.Tools = nil
	answer, err := h.client.Ask(r.Context(), connection, request)
	if err != nil {
		h.writeProviderError(w, connection, err)
		return ai.Response{}, models.AIConnection{}, false
	}
	return answer, connection, true
}

// writeProviderError maps a provider-call failure to the code the clients translate.
func (h *AIHandler) writeProviderError(w http.ResponseWriter, connection models.AIConnection, err error) {
	var refused ai.ErrRefused
	switch {
	case errors.Is(err, ai.ErrNotConfigured):
		writeError(w, http.StatusPreconditionRequired, "Add an AI subscription first", codeAINotConfigured)
	case errors.As(err, &refused):
		slog.Error("ai: the provider refused", "provider", connection.Provider, "model", connection.Model, "err", err)
		writeError(w, http.StatusBadGateway, "The AI provider refused the request: "+refused.Detail, codeAIRefused)
	default:
		slog.Error("ai: the provider could not be reached", "provider", connection.Provider, "err", err)
		writeError(w, http.StatusBadGateway, "The AI provider could not be reached", codeAIUpstream)
	}
}

// --- account context ---

// accountContext tells the model who is asking and which organizations they can act
// in, so it can target /v1/organizations/{orgId}/... paths without guessing. Built
// fresh each question (never stored) so it is always current. Failures are skipped,
// never fatal.
func (h *AIHandler) accountContext(r *http.Request, userID, currentOrgID string) string {
	var out strings.Builder
	out.WriteString("# Your CWClock account\n\n")

	// Today's date, so the model can resolve relative periods ("this month", "last
	// week") correctly — see the date-range rules in the API skill.
	fmt.Fprintf(&out, "- Today's date is %s.\n", time.Now().Format("2006-01-02"))

	if user, err := h.users.FindByID(r.Context(), userID); err == nil {
		if name := strings.TrimSpace(user.Name + " " + user.Surname); name != "" {
			fmt.Fprintf(&out, "- You are helping %s (%s).\n", name, user.Email)
		} else {
			fmt.Fprintf(&out, "- You are helping %s.\n", user.Email)
		}
	}

	orgs, err := h.orgs.ListForUser(r.Context(), userID)
	if err == nil && len(orgs) > 0 {
		out.WriteString("- Organizations you can act in (use these ids in /v1/organizations/{orgId}/... paths):\n")
		for _, o := range orgs {
			role, roleErr := h.orgs.GetRole(r.Context(), o.ID, userID)
			if roleErr != nil {
				role = ""
			}
			current := ""
			if o.ID == currentOrgID {
				current = " [current]"
			}
			fmt.Fprintf(&out, "  - %s (id: %s, your role: %s)%s\n", o.Name, o.ID, role, current)
		}
	} else {
		out.WriteString("- You are not a member of any organization yet.\n")
	}

	out.WriteString("\nUse the organization ids above when calling the cwclock_api tool.\n")
	return out.String()
}

// --- transcript helpers ---

func turnsOf(history []models.AIMessage) []ai.Turn {
	turns := make([]ai.Turn, 0, len(history))
	for _, message := range history {
		turns = append(turns, ai.Turn{Role: message.Role, Text: message.Message})
	}
	return turns
}

func answeredWith(reported, configured string) string {
	for _, candidate := range []string{reported, configured} {
		if strings.TrimSpace(candidate) != "" {
			return strings.TrimSpace(candidate)
		}
	}
	return ""
}

func historyOf(conversation models.AIConversation) []models.AIMessage {
	history := make([]models.AIMessage, 0, len(conversation.Exchanges)*2)
	for _, exchange := range conversation.Exchanges {
		history = append(history,
			models.AIMessage{Role: models.AIChatRoleUser, Message: exchange.Question},
			models.AIMessage{Role: models.AIChatRoleAssistant, Message: exchange.Answer})
	}
	return history
}

func lastQuestion(history []models.AIMessage) string {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == models.AIChatRoleUser {
			return history[i].Message
		}
	}
	return history[len(history)-1].Message
}

// --- language ---

// normalizeLocale keeps only the leading language subtag ("fr-FR" -> "fr").
func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if i := strings.IndexAny(locale, "-_"); i > 0 {
		locale = locale[:i]
	}
	return locale
}

// chatLanguageRule lets the user's own words lead: answer in the language of their
// latest message, falling back to their UI language only when that is unclear.
func chatLanguageRule(locale string) string {
	return "Answer in the language of the user's latest message. " +
		"If that is unclear, answer in " + languageName(locale) + "."
}

// languageName names the locale's language in English (cwclock is English-first,
// with French).
func languageName(locale string) string {
	switch normalizeLocale(locale) {
	case "fr":
		return "French"
	case "en":
		return "English"
	}
	return "English"
}
