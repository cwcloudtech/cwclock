package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

type ExportJobStore struct {
	pool *pgxpool.Pool
}

func NewExportJobStore(pool *pgxpool.Pool) *ExportJobStore {
	return &ExportJobStore{pool: pool}
}

type exportJobData struct {
	Name             string                `json:"name"`
	CronExpression   string                `json:"cronExpression"`
	Targets          []models.ExportTarget `json:"targets"`
	ReportTypes      []string              `json:"reportTypes"`
	TimePeriod       string                `json:"timePeriod"`
	ClientIDs        []string              `json:"clientIds"`
	ProjectIDs       []string              `json:"projectIds"`
	IncludeFinancial bool                  `json:"includeFinancial"`
	Enabled          bool                  `json:"enabled"`
}

type ExportJobFields struct {
	Name             string
	CronExpression   string
	Targets          []models.ExportTarget
	ReportTypes      []string
	TimePeriod       string
	ClientIDs        []string
	ProjectIDs       []string
	IncludeFinancial bool
	Enabled          bool
}

func toExportJobData(f ExportJobFields) exportJobData {
	return exportJobData{
		Name:             f.Name,
		CronExpression:   f.CronExpression,
		Targets:          f.Targets,
		ReportTypes:      f.ReportTypes,
		TimePeriod:       f.TimePeriod,
		ClientIDs:        f.ClientIDs,
		ProjectIDs:       f.ProjectIDs,
		IncludeFinancial: f.IncludeFinancial,
		Enabled:          f.Enabled,
	}
}

func scanExportJob(row pgx.Row) (models.ExportJob, error) {
	var j models.ExportJob
	var raw []byte
	if err := row.Scan(&j.ID, &j.OrganizationID, &raw, &j.CreatedAt, &j.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ExportJob{}, ErrNotFound
		}
		return models.ExportJob{}, err
	}
	var d exportJobData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.ExportJob{}, err
	}
	j.Name = d.Name
	j.CronExpression = d.CronExpression
	j.Targets = d.Targets
	j.ReportTypes = d.ReportTypes
	j.TimePeriod = d.TimePeriod
	j.ClientIDs = d.ClientIDs
	j.ProjectIDs = d.ProjectIDs
	j.IncludeFinancial = d.IncludeFinancial
	j.Enabled = d.Enabled
	return j, nil
}

func (s *ExportJobStore) Create(ctx context.Context, orgID string, f ExportJobFields) (models.ExportJob, error) {
	data, err := json.Marshal(toExportJobData(f))
	if err != nil {
		return models.ExportJob{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO export_jobs (organization_id, data)
		VALUES ($1, $2)
		RETURNING id, organization_id, data, created_at, updated_at
	`, orgID, data)
	return scanExportJob(row)
}

func (s *ExportJobStore) FindByID(ctx context.Context, id string) (models.ExportJob, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, data, created_at, updated_at
		FROM export_jobs WHERE id = $1
	`, id)
	return scanExportJob(row)
}

func (s *ExportJobStore) List(ctx context.Context, orgID string) ([]models.ExportJob, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, data, created_at, updated_at
		FROM export_jobs WHERE organization_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []models.ExportJob{}
	for rows.Next() {
		j, err := scanExportJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (s *ExportJobStore) ListEnabled(ctx context.Context, orgID string) ([]models.ExportJob, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, data, created_at, updated_at
		FROM export_jobs WHERE organization_id = $1 AND (data->>'enabled')::boolean = true
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []models.ExportJob{}
	for rows.Next() {
		j, err := scanExportJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (s *ExportJobStore) Update(ctx context.Context, id string, f ExportJobFields) (models.ExportJob, error) {
	data, err := json.Marshal(toExportJobData(f))
	if err != nil {
		return models.ExportJob{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE export_jobs SET data = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, organization_id, data, created_at, updated_at
	`, id, data)
	return scanExportJob(row)
}

func (s *ExportJobStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM export_jobs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
