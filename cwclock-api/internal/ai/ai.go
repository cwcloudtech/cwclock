// Package ai asks a language model a question on a user's behalf (ai-instruct-132),
// and lets the model call back for help — the cwclock API as a tool — before it
// answers.
//
// Three wire formats, not three vendors. OpenAI's chat-completions is also what
// Mistral, DeepSeek and OpenRouter speak, so a user pointing their connection at a
// local gateway is a change of URL and nothing else. What the format decides is the
// shape of the request body and where the key goes — a header named differently, a
// system prompt that is a field rather than a message — and that is all this package
// knows.
//
// Tool calling is the second thing the format decides. All three protocols let a
// request carry a list of functions the model may call, answer with a call rather
// than prose, and take the call's result back in the next turn — but each spells it
// differently: OpenAI's tool_calls / role=tool messages, Anthropic's tool_use /
// tool_result content blocks, Gemini's functionCall / functionResponse parts. This
// package normalizes the three into Tool / ToolCall / Turn so the caller runs one
// agent loop whatever the provider. A request with no tools behaves exactly as a
// plain chat.
//
// Nothing here is streamed. The credential is the user's own stored connection
// token; there is no global provider key in config.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cwclock-api/internal/models"
)

// ErrNotConfigured is a user asking a question with no connection behind them — no
// key of their own, and nobody has shared one.
var ErrNotConfigured = errors.New("ai: no connection available")

// ErrRefused is the provider rejecting the request: a bad key, a model the account
// cannot reach, a quota. It travels to the screen as itself because the user is the
// only one who can fix it.
type ErrRefused struct {
	Status int
	Detail string
}

func (e ErrRefused) Error() string {
	return fmt.Sprintf("ai: the provider refused the request (http %d): %s", e.Status, e.Detail)
}

// RoleTool is the transcript role of a tool result, alongside the user/assistant
// roles. Each format carries it differently; Turn hides that.
const RoleTool = "tool"

// requestTimeout bounds one call to a provider. A large model reasoning over an
// organization's data takes its time, so it is deliberately generous.
const requestTimeout = 300 * time.Second

// Client talks to one user's connection.
type Client struct {
	http *http.Client
}

// New builds the client with its own timeout-bounded HTTP client.
func New() *Client {
	return &Client{http: &http.Client{Timeout: requestTimeout}}
}

// Tool is a function the model may call to help itself answer: a name, a sentence
// telling the model when to reach for it, and a JSON-Schema object describing its
// arguments. The three formats all take this shape, under different field names.
type Tool struct {
	Name        string
	Description string
	// Parameters is a JSON-Schema object, passed through to whichever format is in use.
	Parameters map[string]any
}

// ToolCall is the model asking to run a tool: which one, with what arguments. ID
// matches the call to its result in the formats that key on an id (OpenAI,
// Anthropic); Gemini matches on Name instead.
type ToolCall struct {
	ID   string
	Name string
	// Args is the arguments object the model produced, as raw JSON.
	Args json.RawMessage
}

// Turn is one entry of the transcript, richer than models.AIMessage because a turn
// may be an assistant's tool call or a tool's result, not only prose.
type Turn struct {
	Role       string
	Text       string
	ToolCalls  []ToolCall
	ToolCallID string
	ToolName   string
}

// TurnOf is a plain text turn — the shape a stored conversation flattens to.
func TurnOf(role, text string) Turn { return Turn{Role: role, Text: text} }

// Request is one call to the model.
type Request struct {
	// System is the instruction the model answers under — the assistant skill plus
	// the account context plus the language rule.
	System string
	// Messages is the transcript so far, oldest first, the last being what is being
	// asked now.
	Messages []Turn
	// Tools are the functions the model may call this turn. Empty means a plain,
	// single-shot answer.
	Tools     []Tool
	MaxTokens int
}

// Response is one answer. When ToolCalls is non-empty the model has not answered
// yet — it wants those calls run and their results handed back before it will
// produce prose.
type Response struct {
	Message   string         `json:"message"`
	ToolCalls []ToolCall     `json:"-"`
	Usage     map[string]any `json:"usage,omitempty"`
	Model     string         `json:"model,omitempty"`
	Tokens    Usage          `json:"tokens"`
}

// Usage is a provider's token count, normalized. Cached input is counted apart.
type Usage struct {
	Input  int64 `json:"input"`
	Output int64 `json:"output"`
	Cached int64 `json:"cached"`
}

// UsageOf reads a provider's own usage object into those three numbers. Three
// vocabularies for the same counts: OpenAI says prompt/completion, Anthropic
// input/output, Gemini promptTokenCount/candidatesTokenCount.
func UsageOf(usage map[string]any) Usage {
	read := func(keys ...string) int64 {
		var total int64
		for _, key := range keys {
			total += number(usage[key])
		}
		return total
	}

	tokens := Usage{
		Input: read("prompt_tokens", "input_tokens", "promptTokenCount"),
		Output: read("completion_tokens", "output_tokens", "candidatesTokenCount",
			"candidates_token_count"),
		Cached: read("cache_read_input_tokens", "cache_creation_input_tokens",
			"cachedContentTokenCount", "cached_content_token_count"),
	}
	// OpenAI nests its cached count a level down.
	if details, ok := usage["prompt_tokens_details"].(map[string]any); ok {
		tokens.Cached += number(details["cached_tokens"])
	}
	return tokens
}

// number reads a JSON value that should have been a count.
func number(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0
		}
		return parsed
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}

// Ask puts the question to the model behind connection. When the request carries
// tools the answer may be a set of ToolCalls rather than prose; the caller runs them
// and asks again with the results appended (see the handler's agent loop).
func (c *Client) Ask(ctx context.Context, connection models.AIConnection, request Request) (Response, error) {
	if !connection.Configured() {
		return Response{}, ErrNotConfigured
	}
	if request.MaxTokens <= 0 {
		request.MaxTokens = connection.MaxTokens
	}
	if request.MaxTokens <= 0 {
		request.MaxTokens = models.DefaultMaxTokens
	}
	model := strings.TrimSpace(connection.Model)

	switch connection.Format {
	case models.AIFormatAnthropic:
		return c.askAnthropic(ctx, connection, model, request)
	case models.AIFormatGemini:
		return c.askGemini(ctx, connection, model, request)
	default:
		return c.askOpenAI(ctx, connection, model, request)
	}
}

// --- OpenAI-compatible (also Mistral, OpenRouter, DeepSeek, a local gateway...) ---

type openAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // a JSON string, not an object (the format's own quirk)
}

type openAIToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    any              `json:"content"` // string, or null on a pure tool-call turn
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

func (c *Client) askOpenAI(ctx context.Context, connection models.AIConnection,
	model string, request Request) (Response, error) {
	body := struct {
		Model     string          `json:"model"`
		Messages  []openAIMessage `json:"messages"`
		MaxTokens int             `json:"max_tokens"`
		Tools     []openAITool    `json:"tools,omitempty"`
	}{Model: model, MaxTokens: request.MaxTokens}

	if request.System != "" {
		body.Messages = append(body.Messages,
			openAIMessage{Role: models.AIChatRoleSystem, Content: request.System})
	}
	for _, turn := range request.Messages {
		body.Messages = append(body.Messages, openAITurn(turn))
	}
	for _, tool := range request.Tools {
		var t openAITool
		t.Type = "function"
		t.Function.Name = tool.Name
		t.Function.Description = tool.Description
		t.Function.Parameters = tool.Parameters
		body.Tools = append(body.Tools, t)
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content   string           `json:"content"`
				ToolCalls []openAIToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage map[string]any `json:"usage"`
		Model string         `json:"model"`
	}
	if err := c.call(ctx, connection, "/chat/completions", map[string]string{
		"Authorization": "Bearer " + connection.Token,
	}, body, &parsed); err != nil {
		return Response{}, err
	}
	if len(parsed.Choices) == 0 {
		return Response{}, errors.New("ai: the provider returned no answer")
	}
	message := parsed.Choices[0].Message
	response := Response{
		Message: strings.TrimSpace(message.Content),
		Usage:   parsed.Usage, Model: parsed.Model, Tokens: UsageOf(parsed.Usage),
	}
	for _, call := range message.ToolCalls {
		response.ToolCalls = append(response.ToolCalls, ToolCall{
			ID: call.ID, Name: call.Function.Name, Args: rawArgs(call.Function.Arguments),
		})
	}
	if len(response.ToolCalls) == 0 && response.Message == "" {
		return Response{}, errors.New("ai: the provider returned no answer")
	}
	return response, nil
}

// openAITurn renders one transcript turn in OpenAI's message shape.
func openAITurn(turn Turn) openAIMessage {
	switch turn.Role {
	case RoleTool:
		return openAIMessage{Role: RoleTool, ToolCallID: turn.ToolCallID, Content: turn.Text}
	case models.AIChatRoleAssistant:
		if len(turn.ToolCalls) > 0 {
			message := openAIMessage{Role: models.AIChatRoleAssistant}
			if turn.Text != "" {
				message.Content = turn.Text
			}
			for _, call := range turn.ToolCalls {
				message.ToolCalls = append(message.ToolCalls, openAIToolCall{
					ID: call.ID, Type: "function",
					Function: openAIFunction{Name: call.Name, Arguments: argsString(call.Args)},
				})
			}
			return message
		}
		return openAIMessage{Role: models.AIChatRoleAssistant, Content: turn.Text}
	default:
		return openAIMessage{Role: models.AIChatRoleUser, Content: turn.Text}
	}
}

// --- Anthropic ---

type anthropicBlock struct {
	Type string `json:"type"`
	// text block
	Text string `json:"text,omitempty"`
	// tool_use block
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
	// tool_result block
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // a plain string, or a []anthropicBlock for tool turns
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

func (c *Client) askAnthropic(ctx context.Context, connection models.AIConnection,
	model string, request Request) (Response, error) {
	body := struct {
		Model     string             `json:"model"`
		System    string             `json:"system,omitempty"`
		Messages  []anthropicMessage `json:"messages"`
		MaxTokens int                `json:"max_tokens"`
		Tools     []anthropicTool    `json:"tools,omitempty"`
	}{Model: model, System: request.System, MaxTokens: request.MaxTokens}

	for i := 0; i < len(request.Messages); {
		turn := request.Messages[i]
		if turn.Role == models.AIChatRoleSystem {
			i++
			continue
		}
		// A round's tool results must arrive as ONE user message of tool_result
		// blocks — two user messages in a row break Claude's role alternation.
		if turn.Role == RoleTool {
			blocks := []anthropicBlock{}
			for i < len(request.Messages) && request.Messages[i].Role == RoleTool {
				result := request.Messages[i]
				blocks = append(blocks, anthropicBlock{
					Type: "tool_result", ToolUseID: result.ToolCallID, Content: result.Text,
				})
				i++
			}
			body.Messages = append(body.Messages,
				anthropicMessage{Role: models.AIChatRoleUser, Content: blocks})
			continue
		}
		body.Messages = append(body.Messages, anthropicTurn(turn))
		i++
	}
	for _, tool := range request.Tools {
		body.Tools = append(body.Tools, anthropicTool{
			Name: tool.Name, Description: tool.Description, InputSchema: tool.Parameters,
		})
	}

	var parsed struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"content"`
		Usage map[string]any `json:"usage"`
		Model string         `json:"model"`
	}
	if err := c.call(ctx, connection, "/messages", map[string]string{
		"x-api-key":         connection.Token,
		"anthropic-version": "2023-06-01",
	}, body, &parsed); err != nil {
		return Response{}, err
	}

	response := Response{Usage: parsed.Usage, Model: parsed.Model, Tokens: UsageOf(parsed.Usage)}
	var answer strings.Builder
	for _, block := range parsed.Content {
		switch block.Type {
		case "text":
			answer.WriteString(block.Text)
		case "tool_use":
			response.ToolCalls = append(response.ToolCalls, ToolCall{
				ID: block.ID, Name: block.Name, Args: rawObject(block.Input),
			})
		}
	}
	response.Message = strings.TrimSpace(answer.String())
	if len(response.ToolCalls) == 0 && response.Message == "" {
		return Response{}, errors.New("ai: the provider returned no answer")
	}
	return response, nil
}

// anthropicTurn renders one non-tool-result turn in Claude's shape.
func anthropicTurn(turn Turn) anthropicMessage {
	switch turn.Role {
	case models.AIChatRoleAssistant:
		if len(turn.ToolCalls) > 0 {
			blocks := []anthropicBlock{}
			if turn.Text != "" {
				blocks = append(blocks, anthropicBlock{Type: "text", Text: turn.Text})
			}
			for _, call := range turn.ToolCalls {
				blocks = append(blocks, anthropicBlock{
					Type: "tool_use", ID: call.ID, Name: call.Name, Input: rawObject(call.Args),
				})
			}
			return anthropicMessage{Role: models.AIChatRoleAssistant, Content: blocks}
		}
		return anthropicMessage{Role: models.AIChatRoleAssistant, Content: turn.Text}
	default:
		return anthropicMessage{Role: models.AIChatRoleUser, Content: turn.Text}
	}
}

// --- Gemini ---

type geminiFunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args,omitempty"`
}

type geminiFunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type geminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse `json:"functionResponse,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

func (c *Client) askGemini(ctx context.Context, connection models.AIConnection,
	model string, request Request) (Response, error) {
	var systemInstruction *geminiContent
	if request.System != "" {
		systemInstruction = &geminiContent{Parts: []geminiPart{{Text: request.System}}}
	}
	var contents []geminiContent
	for i := 0; i < len(request.Messages); {
		turn := request.Messages[i]
		if turn.Role == models.AIChatRoleSystem {
			i++
			continue
		}
		if turn.Role == RoleTool {
			parts := []geminiPart{}
			for i < len(request.Messages) && request.Messages[i].Role == RoleTool {
				result := request.Messages[i]
				parts = append(parts, geminiPart{FunctionResponse: &geminiFunctionResponse{
					Name: result.ToolName, Response: toolResponseObject(result.Text),
				}})
				i++
			}
			contents = append(contents, geminiContent{Role: models.AIChatRoleUser, Parts: parts})
			continue
		}
		contents = append(contents, geminiTurn(turn))
		i++
	}

	var tools []geminiTool
	if len(request.Tools) > 0 {
		declarations := make([]geminiFunctionDeclaration, 0, len(request.Tools))
		for _, tool := range request.Tools {
			declarations = append(declarations, geminiFunctionDeclaration{
				Name: tool.Name, Description: tool.Description, Parameters: tool.Parameters,
			})
		}
		tools = []geminiTool{{FunctionDeclarations: declarations}}
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []geminiPart `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata map[string]any `json:"usageMetadata"`
	}
	path := fmt.Sprintf("/models/%s:generateContent", model)
	if err := c.call(ctx, connection, path, map[string]string{
		"x-goog-api-key": connection.Token,
	}, struct {
		SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
		Contents          []geminiContent `json:"contents"`
		Tools             []geminiTool    `json:"tools,omitempty"`
		GenerationConfig  any             `json:"generationConfig"`
	}{systemInstruction, contents, tools, struct {
		MaxOutputTokens int `json:"maxOutputTokens"`
	}{request.MaxTokens}}, &parsed); err != nil {
		return Response{}, err
	}

	response := Response{Usage: parsed.UsageMetadata, Model: model, Tokens: UsageOf(parsed.UsageMetadata)}
	var answer strings.Builder
	for _, candidate := range parsed.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.FunctionCall != nil {
				response.ToolCalls = append(response.ToolCalls, ToolCall{
					Name: part.FunctionCall.Name, Args: rawObject(part.FunctionCall.Args),
				})
				continue
			}
			answer.WriteString(part.Text)
		}
		break
	}
	response.Message = strings.TrimSpace(answer.String())
	if len(response.ToolCalls) == 0 && response.Message == "" {
		return Response{}, errors.New("ai: the provider returned no answer")
	}
	return response, nil
}

// geminiTurn renders one non-tool-result turn in Gemini's shape (Gemini calls the
// assistant "model").
func geminiTurn(turn Turn) geminiContent {
	switch turn.Role {
	case models.AIChatRoleAssistant:
		if len(turn.ToolCalls) > 0 {
			parts := []geminiPart{}
			if turn.Text != "" {
				parts = append(parts, geminiPart{Text: turn.Text})
			}
			for _, call := range turn.ToolCalls {
				parts = append(parts, geminiPart{FunctionCall: &geminiFunctionCall{
					Name: call.Name, Args: rawObject(call.Args),
				}})
			}
			return geminiContent{Role: "model", Parts: parts}
		}
		return geminiContent{Role: "model", Parts: []geminiPart{{Text: turn.Text}}}
	default:
		return geminiContent{Role: models.AIChatRoleUser, Parts: []geminiPart{{Text: turn.Text}}}
	}
}

// --- tool-argument helpers ---

// rawArgs turns OpenAI's arguments string into the raw JSON an argument object is,
// defaulting an empty or malformed value to {}.
func rawArgs(arguments string) json.RawMessage {
	arguments = strings.TrimSpace(arguments)
	if arguments == "" || !json.Valid([]byte(arguments)) {
		return json.RawMessage("{}")
	}
	return json.RawMessage(arguments)
}

// rawObject keeps a raw-JSON argument object usable on the wire, defaulting a
// missing or invalid one to {}.
func rawObject(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage("{}")
	}
	return raw
}

// argsString renders an argument object back as the JSON string OpenAI echoes tool
// calls in.
func argsString(raw json.RawMessage) string {
	if len(raw) == 0 || !json.Valid(raw) {
		return "{}"
	}
	return string(raw)
}

// toolResponseObject wraps a tool's result for Gemini's functionResponse.response,
// which must be an object.
func toolResponseObject(result string) map[string]any {
	var object map[string]any
	if json.Unmarshal([]byte(result), &object) == nil {
		return object
	}
	return map[string]any{"result": result}
}

// --- the wire ---

func (c *Client) call(ctx context.Context, connection models.AIConnection, path string,
	headers map[string]string, body, out any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(connection.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return ErrRefused{Status: resp.StatusCode, Detail: snippet(raw)}
	}
	return json.Unmarshal(raw, out)
}

// snippet is as much of a refusal as belongs in a log line or an error.
func snippet(body []byte) string {
	const limit = 300
	text := strings.TrimSpace(string(body))
	if len(text) > limit {
		return text[:limit] + "..."
	}
	return text
}
