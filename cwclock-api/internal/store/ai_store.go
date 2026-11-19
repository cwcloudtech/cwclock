package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

// AIStore keeps users' model connections, who may spend them, the provider
// catalogue, and the conversations users have through them (ai-instruct-132).
// Authorization is baked into the SQL: a connection is usable iff the caller owns it
// or holds a share row for it. Payloads live in the JSONB `data` columns of the
// ai_* tables (see V12).
type AIStore struct {
	pool *pgxpool.Pool
}

// NewAIStore builds the AI store over the shared pool.
func NewAIStore(pool *pgxpool.Pool) *AIStore { return &AIStore{pool: pool} }

// ---------------------------------------------------------------------------
// provider catalogue (ai_providers)
// ---------------------------------------------------------------------------

type aiProviderData struct {
	Label   string   `json:"label,omitempty"`
	KeyHint string   `json:"keyHint,omitempty"`
	Models  []string `json:"models,omitempty"`
}

// ListProviders is the catalogue: every provider this deployment knows how to talk
// to, in the order a picker lists them.
func (s *AIStore) ListProviders(ctx context.Context) ([]models.AIProvider, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, format, base_url, default_model, data
		FROM ai_providers ORDER BY position, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	providers := []models.AIProvider{}
	for rows.Next() {
		provider, err := scanProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, rows.Err()
}

// FindProvider is one row of the catalogue. ErrNotFound when the name is unknown.
func (s *AIStore) FindProvider(ctx context.Context, name string) (models.AIProvider, error) {
	return scanProvider(s.pool.QueryRow(ctx, `
		SELECT name, format, base_url, default_model, data
		FROM ai_providers WHERE name = $1`, name))
}

func scanProvider(row pgx.Row) (models.AIProvider, error) {
	var provider models.AIProvider
	var raw []byte
	if err := row.Scan(&provider.Name, &provider.Format, &provider.BaseURL,
		&provider.DefaultModel, &raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AIProvider{}, ErrNotFound
		}
		return models.AIProvider{}, err
	}

	var data aiProviderData
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &data)
	}
	provider.Label, provider.KeyHint, provider.Models = data.Label, data.KeyHint, data.Models
	if provider.Label == "" {
		provider.Label = provider.Name
	}
	if provider.Models == nil {
		provider.Models = []string{}
	}
	return provider, nil
}

// ---------------------------------------------------------------------------
// connections (ai_connections, ai_connection_shares)
// ---------------------------------------------------------------------------

// aiConnectionData is the JSONB payload of ai_connections.
type aiConnectionData struct {
	Name      string `json:"name,omitempty"`
	Provider  string `json:"provider"`
	BaseURL   string `json:"baseUrl"`
	Model     string `json:"model,omitempty"`
	MaxTokens int    `json:"maxTokens,omitempty"`
	Token     string `json:"token,omitempty"`
}

const aiConnectionSelect = `
	SELECT c.id, c.owner_id, c.data, c.created_at, c.updated_at`

func scanAIConnection(row pgx.Row, withRole, withFormat bool) (models.AIConnection, error) {
	var connection models.AIConnection
	var raw []byte
	var created, updated time.Time
	targets := []any{&connection.ID, &connection.OwnerID, &raw, &created, &updated}
	if withRole {
		targets = append(targets, &connection.Role)
	}
	if withFormat {
		targets = append(targets, &connection.Format)
	}
	if err := row.Scan(targets...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AIConnection{}, ErrNotFound
		}
		return models.AIConnection{}, err
	}

	var data aiConnectionData
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &data)
	}
	connection.Name = data.Name
	connection.Provider = data.Provider
	connection.BaseURL = data.BaseURL
	connection.Model = data.Model
	connection.MaxTokens = data.MaxTokens
	connection.Token = data.Token
	connection.HasToken = strings.TrimSpace(data.Token) != ""
	connection.CreatedAt = created
	connection.UpdatedAt = updated
	return connection, nil
}

// FindByID reads one connection, token included.
func (s *AIStore) FindByID(ctx context.Context, id string) (models.AIConnection, error) {
	return scanAIConnection(s.pool.QueryRow(ctx,
		aiConnectionSelect+` FROM ai_connections c WHERE c.id = $1`, id), false, false)
}

// ListFor is every connection a user may spend: their own first, then the ones
// shared with them. Never with a token.
func (s *AIStore) ListFor(ctx context.Context, userID string) ([]models.AIConnection, error) {
	rows, err := s.pool.Query(ctx, aiConnectionSelect+`,
		       CASE WHEN c.owner_id = $1 THEN 'owner' ELSE coalesce(a.role, '') END
		FROM ai_connections c
		LEFT JOIN ai_connection_shares a ON a.connection_id = c.id AND a.user_id = $1
		WHERE c.owner_id = $1 OR a.user_id = $1
		ORDER BY (c.owner_id = $1) DESC, c.position, c.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	connections := []models.AIConnection{}
	for rows.Next() {
		connection, err := scanAIConnection(rows, true, false)
		if err != nil {
			return nil, err
		}
		connection.Token = ""
		connections = append(connections, connection)
	}
	return connections, rows.Err()
}

// RoleFor is what userID may do with this connection, or "" for nothing.
func (s *AIStore) RoleFor(ctx context.Context, connectionID, userID string) (string, error) {
	var role string
	err := s.pool.QueryRow(ctx, `
		SELECT CASE WHEN c.owner_id = $2 THEN 'owner' ELSE coalesce(a.role, '') END
		FROM ai_connections c
		LEFT JOIN ai_connection_shares a ON a.connection_id = c.id AND a.user_id = $2
		WHERE c.id = $1`, connectionID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return role, nil
}

// usableFrom is the join behind both pickers: a connection the user may spend, with
// the wire format of the provider it points at (read from the catalogue).
const usableFrom = `
	FROM ai_connections c
	LEFT JOIN ai_providers p ON p.name = c.data->>'provider'
	LEFT JOIN ai_connection_shares a ON a.connection_id = c.id AND a.user_id = $1`

// Usable picks the connection a user's questions go through when they did not say.
func (s *AIStore) Usable(ctx context.Context, userID string) (models.AIConnection, error) {
	return scanAIConnection(s.pool.QueryRow(ctx, aiConnectionSelect+`, coalesce(p.format, '')`+
		usableFrom+`
		WHERE c.owner_id = $1 OR a.user_id = $1
		ORDER BY (c.owner_id = $1) DESC, c.position, c.created_at
		LIMIT 1`, userID), false, true)
}

// UsableByID is the one the user asked for, if they may spend it. Scoped in the
// statement rather than checked after.
func (s *AIStore) UsableByID(ctx context.Context, userID, connectionID string) (models.AIConnection, error) {
	return scanAIConnection(s.pool.QueryRow(ctx, aiConnectionSelect+`, coalesce(p.format, '')`+
		usableFrom+`
		WHERE c.id = $2 AND (c.owner_id = $1 OR a.user_id = $1)`,
		userID, connectionID), false, true)
}

// Create stores a new connection for ownerID. Appended rather than inserted first.
func (s *AIStore) Create(ctx context.Context, ownerID string, connection models.AIConnection) (models.AIConnection, error) {
	data, err := json.Marshal(aiConnectionData{
		Name: connection.Name, Provider: connection.Provider,
		BaseURL: connection.BaseURL, Model: connection.Model,
		MaxTokens: connection.MaxTokens, Token: connection.Token,
	})
	if err != nil {
		return models.AIConnection{}, err
	}
	return scanAIConnection(s.pool.QueryRow(ctx, `
		INSERT INTO ai_connections (owner_id, data, position)
		VALUES ($1, $2, coalesce((SELECT max(position) + 1 FROM ai_connections WHERE owner_id = $1), 0))
		RETURNING id, owner_id, data, created_at, updated_at`, ownerID, data), false, false)
}

// Update replaces a connection's settings. A blank token keeps the stored one.
func (s *AIStore) Update(ctx context.Context, id string, connection models.AIConnection) (models.AIConnection, error) {
	data, err := json.Marshal(aiConnectionData{
		Name: connection.Name, Provider: connection.Provider,
		BaseURL: connection.BaseURL, Model: connection.Model,
		MaxTokens: connection.MaxTokens,
	})
	if err != nil {
		return models.AIConnection{}, err
	}

	token := strings.TrimSpace(connection.Token)
	return scanAIConnection(s.pool.QueryRow(ctx, `
		UPDATE ai_connections
		SET data = (c.data || $2::jsonb) ||
		           CASE WHEN $3 = '' THEN '{}'::jsonb
		                ELSE jsonb_build_object('token', $3::text) END,
		    updated_at = now()
		FROM ai_connections c
		WHERE ai_connections.id = $1 AND c.id = $1
		RETURNING ai_connections.id, ai_connections.owner_id, ai_connections.data,
		          ai_connections.created_at, ai_connections.updated_at`, id, data, token), false, false)
}

// Delete removes a connection. ErrNotFound when the id is absent.
func (s *AIStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM ai_connections WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Reorder sets which connection is first - the default, and the first row of every
// "ask with" menu. The ids named take the front in the order given; whatever was not
// named follows, keeping its own order.
func (s *AIStore) Reorder(ctx context.Context, ownerID string, ids []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	position := 0
	for _, id := range ids {
		tag, err := tx.Exec(ctx, `
			UPDATE ai_connections SET position = $3, updated_at = now()
			WHERE id = $1 AND owner_id = $2
		`, id, ownerID, position)
		if err != nil {
			return err
		}
		if tag.RowsAffected() > 0 {
			position++
		}
	}

	if _, err := tx.Exec(ctx, `
		WITH rest AS (
			SELECT id, row_number() OVER (ORDER BY position, created_at) - 1 AS offset
			FROM ai_connections
			WHERE owner_id = $1 AND NOT (id::text = ANY($2))
		)
		UPDATE ai_connections c SET position = $3 + rest.offset
		FROM rest WHERE rest.id = c.id
	`, ownerID, ids, position); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Grant gives somebody else the right to spend this connection.
func (s *AIStore) Grant(ctx context.Context, connectionID, userID, role string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ai_connection_shares (connection_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (connection_id, user_id) DO UPDATE SET role = excluded.role, updated_at = now()
	`, connectionID, userID, role)
	return err
}

// Revoke takes it back. Revoking what was never granted is not an error.
func (s *AIStore) Revoke(ctx context.Context, connectionID, userID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM ai_connection_shares WHERE connection_id = $1 AND user_id = $2`, connectionID, userID)
	return err
}

// ListGrants is everybody an owner has shared this connection with (id + email +
// name/surname + role).
func (s *AIStore) ListGrants(ctx context.Context, connectionID string) ([]models.AIConnectionACL, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.user_id, coalesce(u.email, ''),
		       coalesce(u.data->>'name', ''), coalesce(u.data->>'surname', ''),
		       a.role, a.created_at
		FROM ai_connection_shares a JOIN users u ON u.id = a.user_id
		WHERE a.connection_id = $1 ORDER BY a.created_at`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grants := []models.AIConnectionACL{}
	for rows.Next() {
		var grant models.AIConnectionACL
		if err := rows.Scan(&grant.UserID, &grant.Email, &grant.Name, &grant.Surname,
			&grant.Role, &grant.CreatedAt); err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	return grants, rows.Err()
}

// ---------------------------------------------------------------------------
// conversations (ai_conversations, ai_prompts)
// ---------------------------------------------------------------------------

type aiConversationData struct {
	Title string `json:"title,omitempty"`
}

type aiPromptData struct {
	Provider   string         `json:"provider,omitempty"`
	Model      string         `json:"model,omitempty"`
	Connection string         `json:"connection,omitempty"`
	Usage      map[string]any `json:"usage,omitempty"`
}

// ListConversations is a user's threads, newest activity first, without the
// exchanges.
func (s *AIStore) ListConversations(ctx context.Context, userID string) ([]models.AIConversation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.data, c.created_at, c.updated_at,
		       (SELECT count(*) FROM ai_prompts p WHERE p.conversation_id = c.id)
		FROM ai_conversations c
		WHERE c.user_id = $1
		ORDER BY c.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := []models.AIConversation{}
	for rows.Next() {
		var conversation models.AIConversation
		var raw []byte
		if err := rows.Scan(&conversation.ID, &raw, &conversation.CreatedAt,
			&conversation.UpdatedAt, &conversation.Count); err != nil {
			return nil, err
		}
		var data aiConversationData
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &data)
		}
		conversation.Title = data.Title
		conversations = append(conversations, conversation)
	}
	return conversations, rows.Err()
}

// FindConversation reads one thread with its exchanges, for its owner. ErrNotFound
// when the thread is not the caller's or does not exist.
func (s *AIStore) FindConversation(ctx context.Context, id, userID string) (models.AIConversation, error) {
	var conversation models.AIConversation
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT c.id, c.data, c.created_at, c.updated_at
		FROM ai_conversations c WHERE c.id = $1 AND c.user_id = $2`,
		id, userID).Scan(&conversation.ID, &raw, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AIConversation{}, ErrNotFound
		}
		return models.AIConversation{}, err
	}

	var data aiConversationData
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &data)
	}
	conversation.Title = data.Title

	rows, err := s.pool.Query(ctx, `
		SELECT id, question, answer, data, created_at FROM ai_prompts
		WHERE conversation_id = $1 ORDER BY created_at`, conversation.ID)
	if err != nil {
		return models.AIConversation{}, err
	}
	defer rows.Close()

	conversation.Exchanges = []models.AIPrompt{}
	for rows.Next() {
		var exchange models.AIPrompt
		var exchangeRaw []byte
		if err := rows.Scan(&exchange.ID, &exchange.Question, &exchange.Answer,
			&exchangeRaw, &exchange.CreatedAt); err != nil {
			return models.AIConversation{}, err
		}
		var meta aiPromptData
		if len(exchangeRaw) > 0 {
			_ = json.Unmarshal(exchangeRaw, &meta)
		}
		exchange.Provider, exchange.Model = meta.Provider, meta.Model
		exchange.Connection, exchange.Usage = meta.Connection, meta.Usage
		conversation.Exchanges = append(conversation.Exchanges, exchange)
	}
	conversation.Count = len(conversation.Exchanges)
	return conversation, rows.Err()
}

// CreateConversation starts a thread, titled from its first question. Returns its uuid.
func (s *AIStore) CreateConversation(ctx context.Context, userID, title string) (string, error) {
	data, err := json.Marshal(aiConversationData{Title: title})
	if err != nil {
		return "", err
	}
	var id string
	err = s.pool.QueryRow(ctx,
		`INSERT INTO ai_conversations (user_id, data) VALUES ($1, $2) RETURNING id`,
		userID, data).Scan(&id)
	return id, err
}

// AppendExchange stores a question and its answer, and marks the thread as just used
// so the list orders by real activity. The two writes are one transaction.
func (s *AIStore) AppendExchange(ctx context.Context, conversationID string, exchange models.AIPrompt) error {
	data, err := json.Marshal(aiPromptData{
		Provider: exchange.Provider, Model: exchange.Model,
		Connection: exchange.Connection, Usage: exchange.Usage,
	})
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO ai_prompts (conversation_id, question, answer, data) VALUES ($1, $2, $3, $4)
	`, conversationID, exchange.Question, exchange.Answer, data); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE ai_conversations SET updated_at = now() WHERE id = $1`, conversationID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// DeleteConversations removes threads, by id or all of them. Scoped to the owner in
// the statement rather than checked first.
func (s *AIStore) DeleteConversations(ctx context.Context, userID string, ids []string, all bool) (int64, error) {
	if all {
		tag, err := s.pool.Exec(ctx, `DELETE FROM ai_conversations WHERE user_id = $1`, userID)
		if err != nil {
			return 0, err
		}
		return tag.RowsAffected(), nil
	}
	if len(ids) == 0 {
		return 0, nil
	}
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM ai_conversations WHERE user_id = $1 AND id::text = ANY($2)`, userID, ids)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
