package models

import (
	"strings"
	"time"
)

// The AI assistant (ai-instruct-132): a user's connections to a language model, the
// catalogue of providers behind them, and the conversations they have through them.
// The design mirrors the strong-fish / uprodit AI feature; the storage is JSONB
// (see V12). user_id is cwclock's UUID (string), dates are RFC3339 time.Time like
// the rest of the models.

// The three wire formats a provider can speak. Three cover every provider worth
// naming: what differs between OpenAI, Mistral, DeepSeek and OpenRouter is the
// address, not the protocol - only Claude and Gemini need a format of their own.
const (
	AIFormatOpenAI    = "openai"
	AIFormatAnthropic = "anthropic"
	AIFormatGemini    = "gemini"
)

// Roles on a connection. Two, not three: there is no reading a connection without
// spending it.
const (
	AIRoleOwner = "owner"
	AIRoleUser  = "user"
)

// IsValidAIRole reports whether role is one an owner may grant (owner is implicit).
func IsValidAIRole(role string) bool { return role == AIRoleUser }

// CanUseAI reports whether role may ask the model.
func CanUseAI(role string) bool { return role == AIRoleOwner || role == AIRoleUser }

// DefaultMaxTokens is how long an answer may run when a connection carries no
// ceiling of its own.
const DefaultMaxTokens = 4096

// MaxMaxTokens is the largest ceiling the form and the API will accept, so a
// slipped digit cannot turn one question into a bill.
const MaxMaxTokens = 32000

// AIProvider is one row of the provider catalogue: where a provider lives, how to
// talk to it, and what it serves. Data rather than code.
type AIProvider struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	// Format is the wire protocol; not surfaced to clients (it is not theirs to set).
	Format string `json:"-"`
	// BaseURL and DefaultModel are a starting point offered to the user, never
	// imposed. The wire keys are defaultUrl / defaultModel (the shape the UI reads).
	BaseURL      string   `json:"defaultUrl"`
	DefaultModel string   `json:"defaultModel"`
	Models       []string `json:"models"`
	// KeyHint is what this provider's key looks like, for the form's placeholder.
	KeyHint string `json:"keyHint"`
}

// AIConnection is one account's access to a model.
type AIConnection struct {
	ID string `json:"id"`
	// OwnerID is whose connection it is (users.id).
	OwnerID string `json:"ownerId"`
	// Name is what the user calls it; blank falls back to the provider on screen.
	Name     string `json:"name"`
	Provider string `json:"provider"`
	// Format is the provider's wire protocol, read from the catalogue only when a
	// connection is about to be spent. Never on the wire.
	Format  string `json:"-"`
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
	// MaxTokens caps one answer; zero means the deployment default.
	MaxTokens int `json:"maxTokens"`
	// Token is write-only: set through the form, never sent back.
	Token string `json:"-"`
	// HasToken is what a client renders instead - whether a key is stored.
	HasToken bool `json:"hasToken"`
	// Role is the caller's own role on this connection (owner | user).
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Configured reports whether this connection has enough to attempt a call - a shape
// check, not a reachability one.
func (c AIConnection) Configured() bool {
	return strings.TrimSpace(c.Provider) != "" &&
		strings.TrimSpace(c.BaseURL) != "" &&
		strings.TrimSpace(c.Token) != ""
}

// AIConnectionACL is one line of a connection's access list, as its owner reads it:
// who it is shared with, with what role, and since when.
type AIConnectionACL struct {
	UserID    string    `json:"userId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// AIMessage is one turn of a conversation as the chat endpoint takes it.
type AIMessage struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

// Chat roles.
const (
	AIChatRoleUser      = "user"
	AIChatRoleAssistant = "assistant"
	AIChatRoleSystem    = "system"
)

// AIConversation is a thread of questions and answers.
type AIConversation struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// Count is how many exchanges it holds, for a list that shows the thread
	// without loading it.
	Count int `json:"count"`
	// Exchanges is the thread itself, oldest first; absent on the list.
	Exchanges []AIPrompt `json:"exchanges"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// AIPrompt is one stored exchange: a question and the answer it got, with the
// credit of what produced it.
type AIPrompt struct {
	ID         string         `json:"id"`
	Question   string         `json:"question"`
	Answer     string         `json:"answer"`
	Provider   string         `json:"provider"`
	Model      string         `json:"model"`
	Connection string         `json:"connection"`
	Usage      map[string]any `json:"usage"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// AIConversationTitle is what a thread is called before anybody renames it: the
// first question, cut at a word boundary.
func AIConversationTitle(question string) string {
	const limit = 60
	title := strings.Join(strings.Fields(question), " ")
	if title == "" {
		return "New conversation"
	}
	if len(title) <= limit {
		return title
	}
	cut := strings.LastIndex(title[:limit], " ")
	if cut < 20 {
		cut = limit
	}
	return strings.TrimSpace(title[:cut]) + "..."
}
