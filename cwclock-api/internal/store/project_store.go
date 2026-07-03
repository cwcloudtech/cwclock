package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

type ProjectStore struct {
	pool *pgxpool.Pool
}

func NewProjectStore(pool *pgxpool.Pool) *ProjectStore {
	return &ProjectStore{pool: pool}
}

type projectData struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func scanProject(row pgx.Row) (models.Project, error) {
	var p models.Project
	var raw []byte
	if err := row.Scan(&p.ID, &p.OrganizationID, &p.ClientID, &raw, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Project{}, ErrNotFound
		}
		return models.Project{}, err
	}
	var d projectData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.Project{}, err
	}
	p.Name = d.Name
	p.Color = d.Color
	return p, nil
}

func (s *ProjectStore) Create(ctx context.Context, orgID, clientID, name, color string) (models.Project, error) {
	data, err := json.Marshal(projectData{Name: name, Color: color})
	if err != nil {
		return models.Project{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO projects (organization_id, client_id, data)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, client_id, data, created_at, updated_at
	`, orgID, clientID, data)
	return scanProject(row)
}

func (s *ProjectStore) FindByID(ctx context.Context, id string) (models.Project, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, client_id, data, created_at, updated_at
		FROM projects WHERE id = $1
	`, id)
	return scanProject(row)
}

// List returns projects for an organization, optionally filtered by client.
func (s *ProjectStore) List(ctx context.Context, orgID, clientID string) ([]models.Project, error) {
	var rows pgx.Rows
	var err error
	if clientID == "" {
		rows, err = s.pool.Query(ctx, `
			SELECT id, organization_id, client_id, data, created_at, updated_at
			FROM projects WHERE organization_id = $1
			ORDER BY created_at DESC
		`, orgID)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, organization_id, client_id, data, created_at, updated_at
			FROM projects WHERE organization_id = $1 AND client_id = $2
			ORDER BY created_at DESC
		`, orgID, clientID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []models.Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *ProjectStore) Update(ctx context.Context, id, name, color string) (models.Project, error) {
	data, err := json.Marshal(projectData{Name: name, Color: color})
	if err != nil {
		return models.Project{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE projects SET data = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, organization_id, client_id, data, created_at, updated_at
	`, id, data)
	return scanProject(row)
}

func (s *ProjectStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
