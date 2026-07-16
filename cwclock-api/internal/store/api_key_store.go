package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

type ApiKeyStore struct {
	pool *pgxpool.Pool
}

func NewApiKeyStore(pool *pgxpool.Pool) *ApiKeyStore {
	return &ApiKeyStore{pool: pool}
}

// generateToken returns a random 32-byte, hex-encoded plaintext API key.
func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return utils.EMPTY, err
	}
	return hex.EncodeToString(buf), nil
}

func scanApiKey(row pgx.Row) (models.ApiKey, error) {
	var k models.ApiKey
	if err := row.Scan(&k.ID, &k.UserID, &k.Description, &k.ExpiresAt, &k.CreatedAt, &k.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ApiKey{}, ErrNotFound
		}
		return models.ApiKey{}, err
	}
	return k, nil
}

// Create mints a new API key for userID and returns both the stored record
// and the plaintext token - the only time the plaintext is ever available.
func (s *ApiKeyStore) Create(ctx context.Context, userID, description string, expiresAt *time.Time) (models.ApiKey, string, error) {
	token, err := generateToken()
	if err != nil {
		return models.ApiKey{}, "", err
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO api_keys (user_id, key_hash, description, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, description, expires_at, created_at, updated_at
	`, userID, utils.HashToken(token), description, expiresAt)
	k, err := scanApiKey(row)
	if err != nil {
		return models.ApiKey{}, "", err
	}
	return k, token, nil
}

// ListByUser returns userID's API keys, most recently created first.
func (s *ApiKeyStore) ListByUser(ctx context.Context, userID string) ([]models.ApiKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, description, expires_at, created_at, updated_at
		FROM api_keys WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []models.ApiKey{}
	for rows.Next() {
		k, err := scanApiKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// Delete revokes an API key, scoped to its owner so a user can never revoke
// someone else's key.
func (s *ApiKeyStore) Delete(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM api_keys WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// VerifyHash looks up the user a still-valid (non-expired) API key belongs
// to, by the sha256 hash of its plaintext token.
func (s *ApiKeyStore) VerifyHash(ctx context.Context, hash string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx, `
		SELECT user_id FROM api_keys
		WHERE key_hash = $1 AND (expires_at IS NULL OR expires_at > now())
	`, hash).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return utils.EMPTY, ErrNotFound
		}
		return utils.EMPTY, err
	}
	return userID, nil
}
