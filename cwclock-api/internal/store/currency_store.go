package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

// CurrencyStore reads the currencies table, which replaced the
// CWCLOCK_ALLOWED_CURRENCIES env var (ai-instruct-35) as the source of
// ISO 4217 codes organizations may be billed in.
type CurrencyStore struct {
	pool *pgxpool.Pool
}

func NewCurrencyStore(pool *pgxpool.Pool) *CurrencyStore {
	return &CurrencyStore{pool: pool}
}

func (s *CurrencyStore) List(ctx context.Context) ([]models.Currency, error) {
	rows, err := s.pool.Query(ctx, `SELECT iso_code, name FROM currencies ORDER BY iso_code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	currencies := []models.Currency{}
	for rows.Next() {
		var c models.Currency
		if err := rows.Scan(&c.ISO, &c.Name); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}
	return currencies, rows.Err()
}

// Exists reports whether code is a known ISO 4217 currency code.
func (s *CurrencyStore) Exists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM currencies WHERE iso_code = $1)`, code).Scan(&exists)
	return exists, err
}
