package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

type TimeEntryStore struct {
	pool *pgxpool.Pool
}

func NewTimeEntryStore(pool *pgxpool.Pool) *TimeEntryStore {
	return &TimeEntryStore{pool: pool}
}

type timeEntryData struct {
	Text   string  `json:"text"`
	Status bool    `json:"status"`
	Day    string  `json:"day"`
	Start  *string `json:"start"`
	End    *string `json:"end"`
	AllDay bool    `json:"allDay"`
}

// TimeEntryFields holds the editable fields of a time entry. When AllDay is
// true, Start/End are always stored as null regardless of what's passed.
type TimeEntryFields struct {
	Text   string
	Status bool
	Day    string
	Start  *string
	End    *string
	AllDay bool
}

func toTimeEntryData(f TimeEntryFields) timeEntryData {
	d := timeEntryData{Text: f.Text, Status: f.Status, Day: f.Day, AllDay: f.AllDay}
	if !f.AllDay {
		d.Start = f.Start
		d.End = f.End
	}
	return d
}

func scanTimeEntry(row pgx.Row) (models.TimeEntry, error) {
	var t models.TimeEntry
	var raw []byte
	if err := row.Scan(&t.ID, &t.OrganizationID, &t.ClientID, &t.ProjectID, &t.UserID, &raw, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TimeEntry{}, ErrNotFound
		}
		return models.TimeEntry{}, err
	}
	var d timeEntryData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.TimeEntry{}, err
	}
	t.Text = d.Text
	t.Status = d.Status
	t.Day = d.Day
	t.Start = d.Start
	t.End = d.End
	t.AllDay = d.AllDay
	return t, nil
}

func (s *TimeEntryStore) Create(ctx context.Context, orgID, clientID, projectID, userID string, f TimeEntryFields) (models.TimeEntry, error) {
	data, err := json.Marshal(toTimeEntryData(f))
	if err != nil {
		return models.TimeEntry{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO time_entries (organization_id, client_id, project_id, user_id, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
	`, orgID, clientID, projectID, userID, data)
	return scanTimeEntry(row)
}

func (s *TimeEntryStore) FindByID(ctx context.Context, id string) (models.TimeEntry, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
		FROM time_entries WHERE id = $1
	`, id)
	return scanTimeEntry(row)
}

// List returns time entries for an organization. When userID is non-empty,
// results are restricted to that user's own entries (used to enforce that
// members only see their own time records).
func (s *TimeEntryStore) List(ctx context.Context, orgID, userID string) ([]models.TimeEntry, error) {
	var rows pgx.Rows
	var err error
	if userID == "" {
		rows, err = s.pool.Query(ctx, `
			SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
			FROM time_entries WHERE organization_id = $1
			ORDER BY created_at DESC
		`, orgID)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
			FROM time_entries WHERE organization_id = $1 AND user_id = $2
			ORDER BY created_at DESC
		`, orgID, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.TimeEntry{}
	for rows.Next() {
		t, err := scanTimeEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, t)
	}
	return entries, rows.Err()
}

// Update updates a time entry's fields. When reassignUserID is non-empty,
// the entry is also reassigned to that user (callers must have already
// authorized the reassignment).
func (s *TimeEntryStore) Update(ctx context.Context, id, reassignUserID string, f TimeEntryFields) (models.TimeEntry, error) {
	data, err := json.Marshal(toTimeEntryData(f))
	if err != nil {
		return models.TimeEntry{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE time_entries
		SET data = $2, user_id = COALESCE(NULLIF($3, '')::uuid, user_id), updated_at = now()
		WHERE id = $1
		RETURNING id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
	`, id, data, reassignUserID)
	return scanTimeEntry(row)
}

func (s *TimeEntryStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM time_entries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
