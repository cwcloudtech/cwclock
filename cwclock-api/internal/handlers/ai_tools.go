package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"cwclock-api/internal/ai"
)

// The assistant's tool (ai-instruct-132). The model is handed cwclock's own API and
// expected to call it — read the user's data, and generate reports and invoices —
// before it answers, rather than telling the user how to click through the UI. One
// tool: cwclock_api, a GET/POST against this very API as the signed-in user (its own
// Auth / OrgMembership / RequireRole gates decide what comes back).
const (
	toolCwclockAPIName = "cwclock_api"

	// apiToolTimeout bounds one cwclock_api call. Generous: a report over a long
	// range, or an invoice render, takes a moment.
	apiToolTimeout = 30 * time.Second
	// apiToolMaxBytes caps how much of a response is read into memory.
	apiToolMaxBytes = 2 << 20
	// apiToolTextCap caps the text body handed back to the model, so a large report
	// grounds an answer without blowing the prompt budget.
	apiToolTextCap = 64 * 1024
)

// apiToolClient is the HTTP client the cwclock_api tool calls this very API with.
// Its proxy is pinned off: the call is to the loopback self-base, and routing a
// self-call through a corporate proxy would hang, then fail.
var apiToolClient = &http.Client{
	Timeout:   apiToolTimeout,
	Transport: &http.Transport{Proxy: nil},
}

// aiTools is the tool catalogue offered to the model each chat turn.
func (h *AIHandler) aiTools() []ai.Tool {
	return []ai.Tool{
		{
			Name: toolCwclockAPIName,
			Description: "Call CWClock's own HTTP API as the signed-in user, to ground your answer in " +
				"their real data, to generate reports and invoices, and to fill a client/project's " +
				"working days with time entries. It runs with the user's OWN permissions: a user who is " +
				"not allowed gets a forbidden result you must relay as 'you don't have permission' " +
				"without retrying. Use method GET to read (clients, projects, members, time entries, " +
				"invoices, and the working-days preview) and method POST to generate (reports, invoices) " +
				"or to fill working days. The skill lists the useful paths, how organization-scoped paths " +
				"work, and how to resolve a client/project/person name to its id. The ONLY records this " +
				"tool can create are time entries, via the working-days fill endpoint, and ONLY after you " +
				"have previewed the exact days and the user has confirmed; it can otherwise never change, " +
				"delete or send anything, so never claim it did.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"method": map[string]any{
						"type":        "string",
						"enum":        []string{"GET", "POST"},
						"description": "GET to read, POST to generate a report/invoice or to fill working days (only after the user confirms the previewed days). Defaults to GET.",
					},
					"path": map[string]any{
						"type": "string",
						"description": "The API path, starting with '/v1/'. May carry its own query, e.g. " +
							"'/v1/organizations/{orgId}/clients', '/v1/organizations/{orgId}/reports/summary', " +
							"'/v1/organizations/{orgId}/invoices/'.",
					},
					"body": map[string]any{
						"type":        "object",
						"description": "JSON body for a POST (the report or invoice request). Omit for a GET.",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

// runTool executes one tool call and returns its result as a JSON string. It never
// returns an error: a failed tool reports the failure to the model as data, so the
// agent loop and the chat carry on.
func (h *AIHandler) runTool(r *http.Request, call ai.ToolCall) string {
	switch call.Name {
	case toolCwclockAPIName:
		var args struct {
			Method string          `json:"method"`
			Path   string          `json:"path"`
			Body   json.RawMessage `json:"body"`
		}
		_ = json.Unmarshal(call.Args, &args)
		return h.toolCwclockAPI(r, args.Method, args.Path, args.Body)
	default:
		return aiToolErrorJSON("unknown tool")
	}
}

// toolCwclockAPI performs one GET or POST against cwclock's own API as the signed-in
// user and returns the response as a JSON string the model reads. The whole point is
// that it carries the CALLER's own credential — the X-Api-Key or Bearer token that
// authenticated the incoming chat — so the API's own authorization decides what
// comes back: an admin reaches invoice generation, a member is refused (403),
// surfaced here as data so the model says "you don't have permission". GET and POST
// only (no delete/update), so the model can read and generate but never destroy. The
// path is pinned to this API (relative, under /v1/), so the user's token can never
// be pointed somewhere else. Every failure is returned as JSON, not an error, so a
// bad tool call never breaks the chat.
func (h *AIHandler) toolCwclockAPI(r *http.Request, method, path string, body json.RawMessage) string {
	if h.apiBaseURL == "" {
		return aiToolJSON(map[string]any{"error": "cwclock api access unavailable"})
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}
	if method != http.MethodGet && method != http.MethodPost {
		return aiToolJSON(map[string]any{"error": "method must be GET or POST"})
	}

	// The model supplies a server-relative path under /v1/ only. Reject anything that
	// could reach off this API — an absolute URL, a scheme, a protocol-relative
	// //host, or an opaque/host component.
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return aiToolJSON(map[string]any{"error": "path must be an absolute path on this API, e.g. /v1/organizations"})
	}
	parsed, err := url.Parse(path)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.Opaque != "" {
		return aiToolJSON(map[string]any{"error": "path must be a relative API path, not a full URL"})
	}
	if !strings.HasPrefix(parsed.EscapedPath(), "/v1/") {
		return aiToolJSON(map[string]any{"error": "path must start with /v1/"})
	}

	requestURI := parsed.EscapedPath()
	if parsed.RawQuery != "" {
		requestURI += "?" + parsed.RawQuery
	}
	endpoint := strings.TrimRight(h.apiBaseURL, "/") + requestURI

	var reqBody io.Reader
	if method == http.MethodPost {
		if len(bytes.TrimSpace(body)) == 0 || !json.Valid(body) {
			body = json.RawMessage("{}")
		}
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(r.Context(), method, endpoint, reqBody)
	if err != nil {
		return aiToolJSON(map[string]any{"error": "could not build the request"})
	}
	// Forward whatever credential authenticated the chat (X-Api-Key for mobile,
	// Bearer for web); the self-call then acts as exactly this user.
	if apiKey := r.Header.Get("X-Api-Key"); apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if cid := r.Header.Get("X-Cid"); cid != "" {
		req.Header.Set("X-Cid", cid)
	}
	req.Header.Set("Accept", "application/json, application/pdf, text/csv, text/plain")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := apiToolClient.Do(req)
	if err != nil {
		return aiToolJSON(map[string]any{"error": "the cwclock API could not be reached"})
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return aiToolJSON(map[string]any{
			"status": resp.StatusCode, "forbidden": true,
			"note": "the signed-in user is not allowed to do this; tell them they do not have " +
				"permission for it and do not retry",
		})
	}

	// A JSON/text body is read small — just enough to ground an answer. A generated
	// file (PDF/CSV) is binary and must be read whole so it can be parked for the
	// user to download, hence the larger cap when a download store is wired.
	binary := !aiLooksTextual(contentType)
	limit := int64(apiToolMaxBytes)
	if binary && h.downloads != nil && h.downloads.MaxBytes() > 0 {
		limit = h.downloads.MaxBytes()
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, limit))

	if binary || !utf8.Valid(raw) {
		return h.toolFileResult(r, method, resp.StatusCode, contentType, resp.Header.Get("Content-Disposition"), raw)
	}

	text, truncated := aiCapText(raw)
	return aiToolJSON(map[string]any{
		"status": resp.StatusCode, "contentType": contentType, "body": text, "truncated": truncated,
	})
}

// toolFileResult handles a binary (file) response — an invoice or an exported
// report PDF/CSV. When a download store is wired and the caller is known, it parks
// the bytes and hands the model a `downloadUrl` to post as a clickable link;
// otherwise it just names the file. Either way the model never sees the bytes.
func (h *AIHandler) toolFileResult(r *http.Request, method string, status int, contentType, disposition string, raw []byte) string {
	filename := aiDispositionFilename(disposition)
	result := map[string]any{"status": status, "contentType": contentType}
	if filename != "" {
		result["filename"] = filename
	}

	if status < 300 && len(raw) > 0 && h.downloads != nil {
		if userID, ok := aiUserID(r); ok {
			name := filename
			if name == "" {
				name = aiFallbackFilename(contentType)
			}
			if url, _, err := h.downloads.Save(userID, name, raw); err == nil {
				result["filename"] = name
				result["downloadUrl"] = url
				result["note"] = "a file was generated. Give the user this exact downloadUrl as a clickable " +
					"Markdown link — [" + name + "](" + url + ") — and tell them the link is temporary. " +
					"Never invent, alter or guess a download URL; only ever use one a tool returned."
				return aiToolJSON(result)
			}
		}
	}

	if method == http.MethodPost && status < 300 {
		result["note"] = "a file was generated (report or invoice); its name is the filename above " +
			"and it is available in the app. The file's contents are not shown here."
	} else {
		result["note"] = "binary response, not shown"
	}
	return aiToolJSON(result)
}

// aiLooksTextual is true for the content types whose body is worth returning
// verbatim (JSON, plain text, CSV, XML).
func aiLooksTextual(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "json") || strings.Contains(ct, "text") ||
		strings.Contains(ct, "csv") || strings.Contains(ct, "xml")
}

// aiCapText bounds one text field handed to the model, flagging when it was cut.
func aiCapText(b []byte) (string, bool) {
	if len(b) > apiToolTextCap {
		return string(b[:apiToolTextCap]), true
	}
	return string(b), false
}

// aiDispositionFilename pulls the filename out of a Content-Disposition header (e.g.
// `attachment; filename="2026-0001.pdf"`), so an invoice generation can be reported
// by its number. "" when there is none.
func aiDispositionFilename(disposition string) string {
	if disposition == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(params["filename"])
}

// aiFallbackFilename names a generated file when the response carried no
// Content-Disposition, from its content type.
func aiFallbackFilename(contentType string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "pdf"):
		return "document.pdf"
	case strings.Contains(ct, "csv"):
		return "export.csv"
	default:
		return "download"
	}
}

// aiToolJSON encodes a tool result; a marshalling failure still comes back as a JSON
// error the model can read.
func aiToolJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return aiToolErrorJSON("could not encode tool result")
	}
	return string(encoded)
}

// aiToolErrorJSON is a tool's failure, reported to the model as data.
func aiToolErrorJSON(message string) string {
	encoded, _ := json.Marshal(map[string]string{"error": message})
	return string(encoded)
}
