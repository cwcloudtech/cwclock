package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

// CountryStore reads the countries table, which replaced free-text country
// input on organizations/clients (ai-instruct-35) with a closed list of
// ISO 3166-1 alpha-2 codes, each with a default billing currency.
type CountryStore struct {
	pool *pgxpool.Pool
}

func NewCountryStore(pool *pgxpool.Pool) *CountryStore {
	return &CountryStore{pool: pool}
}

// List returns every country ordered by its curated sort_order
// (ai-instruct-36), falling back to name to break ties between countries
// sharing the same rank (e.g. every country billed in the same currency).
func (s *CountryStore) List(ctx context.Context) ([]models.Country, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT iso_code, name, currency_iso_code FROM countries ORDER BY sort_order, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	countries := []models.Country{}
	for rows.Next() {
		var c models.Country
		if err := rows.Scan(&c.ISO, &c.Name, &c.Currency); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}
	return countries, rows.Err()
}

// Exists reports whether iso is a known ISO 3166-1 alpha-2 country code.
func (s *CountryStore) Exists(ctx context.Context, iso string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM countries WHERE iso_code = $1)`, iso).Scan(&exists)
	return exists, err
}

// DefaultCurrency returns iso's default billing currency. Blank/unknown
// countries resolve to models.FallbackCurrency, so callers never need a
// second fallback layer of their own.
func (s *CountryStore) DefaultCurrency(ctx context.Context, iso string) (string, error) {
	if utils.IsBlank(iso) {
		return models.FallbackCurrency, nil
	}
	var currency string
	err := s.pool.QueryRow(ctx, `SELECT currency_iso_code FROM countries WHERE iso_code = $1`, iso).Scan(&currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.FallbackCurrency, nil
		}
		return utils.EMPTY, err
	}
	return currency, nil
}
