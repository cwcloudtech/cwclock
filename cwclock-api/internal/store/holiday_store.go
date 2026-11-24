package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

// HolidayStore is the country-keyed public-holiday decision table. Rows are
// deployment-wide (superuser-managed), not per organization.
type HolidayStore struct {
	pool *pgxpool.Pool
}

func NewHolidayStore(pool *pgxpool.Pool) *HolidayStore {
	return &HolidayStore{pool: pool}
}

// HolidayFields holds the editable fields of a decision-table rule.
type HolidayFields struct {
	Name         string
	Type         models.HolidayType
	Month        int
	Day          int
	EasterOffset int
	Weekday      string
	Nth          int
}

// holidayData is the jsonb payload stored in public_holidays.data.
type holidayData struct {
	Name         string             `json:"name"`
	Type         models.HolidayType `json:"type"`
	Month        int                `json:"month,omitempty"`
	Day          int                `json:"day,omitempty"`
	EasterOffset int                `json:"easterOffset,omitempty"`
	Weekday      string             `json:"weekday,omitempty"`
	Nth          int                `json:"nth,omitempty"`
}

func toHolidayData(f HolidayFields) holidayData {
	d := holidayData{Name: f.Name, Type: f.Type}
	switch f.Type {
	case models.HolidayFixed:
		d.Month = f.Month
		d.Day = f.Day
	case models.HolidayEaster:
		d.EasterOffset = f.EasterOffset
	case models.HolidayWeekend:
		d.Weekday = f.Weekday
	case models.HolidayNthWeekday:
		d.Month = f.Month
		d.Weekday = f.Weekday
		d.Nth = f.Nth
	}
	return d
}

func scanHoliday(row pgx.Row) (models.PublicHoliday, error) {
	var h models.PublicHoliday
	var raw []byte
	if err := row.Scan(&h.ID, &h.Country, &raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PublicHoliday{}, ErrNotFound
		}
		return models.PublicHoliday{}, err
	}
	var d holidayData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.PublicHoliday{}, err
	}
	h.Name = d.Name
	h.Type = d.Type
	h.Month = d.Month
	h.Day = d.Day
	h.EasterOffset = d.EasterOffset
	h.Weekday = d.Weekday
	h.Nth = d.Nth
	return h, nil
}

// orderedHolidaySQL sorts rules by their place in the civil year so the admin
// screen reads top-to-bottom: fixed rules by month/day, movable (easter) rules
// by their offset, then name.
const orderedHolidaySQL = `
	ORDER BY country,
	         COALESCE((data->>'month')::int, 99),
	         COALESCE((data->>'day')::int, 99),
	         COALESCE((data->>'easterOffset')::int, 0),
	         data->>'name'`

// List returns the holiday rules, optionally filtered to one country (empty =
// all). Used by the admin screen.
func (s *HolidayStore) List(ctx context.Context, country string) ([]models.PublicHoliday, error) {
	country = strings.ToUpper(strings.TrimSpace(country))
	query := `SELECT id, country, data FROM public_holidays`
	args := []any{}
	if country != "" {
		query += ` WHERE country = $1`
		args = append(args, country)
	}
	query += orderedHolidaySQL
	return s.query(ctx, query, args...)
}

// ListByCountry returns the holiday rules for one ISO country code, for the
// working-day resolver.
func (s *HolidayStore) ListByCountry(ctx context.Context, country string) ([]models.PublicHoliday, error) {
	country = strings.ToUpper(strings.TrimSpace(country))
	if country == "" {
		return nil, nil
	}
	return s.query(ctx, `SELECT id, country, data FROM public_holidays WHERE country = $1`+orderedHolidaySQL, country)
}

func (s *HolidayStore) query(ctx context.Context, query string, args ...any) ([]models.PublicHoliday, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.PublicHoliday{}
	for rows.Next() {
		h, err := scanHoliday(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *HolidayStore) FindByID(ctx context.Context, id string) (models.PublicHoliday, error) {
	return scanHoliday(s.pool.QueryRow(ctx, `SELECT id, country, data FROM public_holidays WHERE id = $1`, id))
}

func (s *HolidayStore) Create(ctx context.Context, country string, f HolidayFields) (models.PublicHoliday, error) {
	data, err := json.Marshal(toHolidayData(f))
	if err != nil {
		return models.PublicHoliday{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO public_holidays (country, data)
		VALUES ($1, $2)
		RETURNING id, country, data
	`, strings.ToUpper(strings.TrimSpace(country)), data)
	return scanHoliday(row)
}

func (s *HolidayStore) Update(ctx context.Context, id, country string, f HolidayFields) (models.PublicHoliday, error) {
	data, err := json.Marshal(toHolidayData(f))
	if err != nil {
		return models.PublicHoliday{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE public_holidays
		SET country = $2, data = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, country, data
	`, id, strings.ToUpper(strings.TrimSpace(country)), data)
	return scanHoliday(row)
}

func (s *HolidayStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM public_holidays WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
