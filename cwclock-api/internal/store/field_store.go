package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FieldStore reads the fields table: the per-country decision table of
// which business identification fields to display (ai-instruct-35), e.g.
// FR has SIRET/SIREN/NAF rows, TN has an MF row, every EU country has a
// VAT Code row.
type FieldStore struct {
	pool *pgxpool.Pool
}

func NewFieldStore(pool *pgxpool.Pool) *FieldStore {
	return &FieldStore{pool: pool}
}

// ListForCountry returns iso's raw fields rows, ordered so a country's own
// identification field(s) (e.g. FR's SIRET/SIREN/NAF, TN's MF) come before
// its EU VAT Code row, matching the order ai-instruct-35's example expects.
// Callers wanting the "Identification number" default applied on top of
// this should use models.ResolveFields.
func (s *FieldStore) ListForCountry(ctx context.Context, iso string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name FROM fields WHERE iso_code = $1
		ORDER BY CASE name
			WHEN 'SIRET' THEN 1
			WHEN 'SIREN' THEN 2
			WHEN 'NAF' THEN 3
			WHEN 'MF' THEN 4
			WHEN 'VAT Code' THEN 100
			ELSE 50
		END
	`, iso)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}
