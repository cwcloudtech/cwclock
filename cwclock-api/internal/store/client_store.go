package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

type ClientStore struct {
	pool *pgxpool.Pool
}

func NewClientStore(pool *pgxpool.Pool) *ClientStore {
	return &ClientStore{pool: pool}
}

type clientData struct {
	Name                 string   `json:"name"`
	Email                string   `json:"email,omitempty"`
	ContactName          string   `json:"contactName,omitempty"`
	Address              string   `json:"address"`
	PostalCode           string   `json:"postalCode"`
	City                 string   `json:"city"`
	Country              string   `json:"country"`
	VATNumber            string   `json:"vatNumber"`
	VATRate              float64  `json:"vatRate"`
	VATDischargeMotive   string   `json:"vatDischargeMotive"`
	SIREN                string   `json:"siren,omitempty"`
	SIRET                string   `json:"siret,omitempty"`
	NAF                  string   `json:"naf,omitempty"`
	MF                   string   `json:"mf,omitempty"`
	IdentificationNumber string   `json:"identificationNumber,omitempty"`
	PurchaseOrder        string   `json:"purchaseOrder"`
	HoursPerDay          float64  `json:"hoursPerDay"`
	DailyRate            *float64 `json:"dailyRate,omitempty"`
}

// ClientFields holds the editable, non-identifying fields of a client.
type ClientFields struct {
	Name                 string
	Email                string
	ContactName          string
	Address              string
	PostalCode           string
	City                 string
	Country              string
	VATNumber            string
	VATRate              *float64
	VATDischargeMotive   string
	SIREN                string
	SIRET                string
	NAF                  string
	MF                   string
	IdentificationNumber string
	PurchaseOrder        string
	HoursPerDay          float64
	DailyRate            *float64
}

// defaultVATRate is applied when the client has no VAT rate set at all, or
// when it's negative; an explicit 0 (VAT-exempt) is kept as-is.
const defaultVATRate float64 = 20

func toClientData(f ClientFields) clientData {
	vatRate := defaultVATRate
	if f.VATRate != nil && *f.VATRate >= 0 {
		vatRate = *f.VATRate
	}
	hoursPerDay := f.HoursPerDay
	if hoursPerDay == 0 {
		hoursPerDay = 7
	}
	return clientData{
		Name:                 f.Name,
		Email:                f.Email,
		ContactName:          f.ContactName,
		Address:              f.Address,
		PostalCode:           f.PostalCode,
		City:                 f.City,
		Country:              f.Country,
		VATNumber:            f.VATNumber,
		VATRate:              vatRate,
		VATDischargeMotive:   f.VATDischargeMotive,
		SIREN:                f.SIREN,
		SIRET:                f.SIRET,
		NAF:                  f.NAF,
		MF:                   f.MF,
		IdentificationNumber: f.IdentificationNumber,
		PurchaseOrder:        f.PurchaseOrder,
		HoursPerDay:          hoursPerDay,
		DailyRate:            f.DailyRate,
	}
}

func scanClient(row pgx.Row) (models.Client, error) {
	var c models.Client
	var raw []byte
	if err := row.Scan(&c.ID, &c.OrganizationID, &raw, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Client{}, ErrNotFound
		}
		return models.Client{}, err
	}
	var d clientData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.Client{}, err
	}
	c.Name = d.Name
	c.Email = d.Email
	c.ContactName = d.ContactName
	c.Address = d.Address
	c.PostalCode = d.PostalCode
	c.City = d.City
	c.Country = d.Country
	c.VATNumber = d.VATNumber
	c.VATRate = d.VATRate
	c.VATDischargeMotive = d.VATDischargeMotive
	c.SIREN = d.SIREN
	c.SIRET = d.SIRET
	c.NAF = d.NAF
	c.MF = d.MF
	c.IdentificationNumber = d.IdentificationNumber
	c.PurchaseOrder = d.PurchaseOrder
	c.HoursPerDay = d.HoursPerDay
	c.DailyRate = d.DailyRate
	return c, nil
}

func (s *ClientStore) Create(ctx context.Context, orgID string, f ClientFields) (models.Client, error) {
	data, err := json.Marshal(toClientData(f))
	if err != nil {
		return models.Client{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO clients (organization_id, data)
		VALUES ($1, $2)
		RETURNING id, organization_id, data, created_at, updated_at
	`, orgID, data)
	return scanClient(row)
}

func (s *ClientStore) FindByID(ctx context.Context, id string) (models.Client, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, data, created_at, updated_at
		FROM clients WHERE id = $1
	`, id)
	return scanClient(row)
}

func (s *ClientStore) List(ctx context.Context, orgID string) ([]models.Client, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, data, created_at, updated_at
		FROM clients WHERE organization_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []models.Client{}
	for rows.Next() {
		c, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

func (s *ClientStore) Update(ctx context.Context, id string, f ClientFields) (models.Client, error) {
	data, err := json.Marshal(toClientData(f))
	if err != nil {
		return models.Client{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE clients SET data = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, organization_id, data, created_at, updated_at
	`, id, data)
	return scanClient(row)
}

// Transfer moves a client - and, to keep every table's own denormalized
// organization_id column consistent with it, all of its projects and time
// entries too - to a different organization. The caller is responsible for
// verifying the acting user owns both organizations (see ai-instruct-34).
func (s *ClientStore) Transfer(ctx context.Context, id, targetOrgID string) (models.Client, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Client{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE clients SET organization_id = $2, updated_at = now() WHERE id = $1
	`, id, targetOrgID)
	if err != nil {
		return models.Client{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Client{}, ErrNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE projects SET organization_id = $2, updated_at = now() WHERE client_id = $1
	`, id, targetOrgID); err != nil {
		return models.Client{}, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE time_entries SET organization_id = $2, updated_at = now() WHERE client_id = $1
	`, id, targetOrgID); err != nil {
		return models.Client{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Client{}, err
	}
	return s.FindByID(ctx, id)
}

func (s *ClientStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM clients WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Count returns the total number of clients across every organization, for
// the "counter of clients" metric.
func (s *ClientStore) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM clients`).Scan(&count)
	return count, err
}

// ListAll returns every client across every organization, keyed lookups
// (e.g. hoursPerDay for the task-duration metric) don't need one query per
// client.
func (s *ClientStore) ListAll(ctx context.Context) ([]models.Client, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, data, created_at, updated_at FROM clients
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []models.Client{}
	for rows.Next() {
		c, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

// FindByName returns the first client in orgID whose name exactly matches name.
func (s *ClientStore) FindByName(ctx context.Context, orgID, name string) (models.Client, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, data, created_at, updated_at
		FROM clients WHERE organization_id = $1 AND data->>'name' = $2
		LIMIT 1
	`, orgID, name)
	return scanClient(row)
}
