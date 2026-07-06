package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

type OrgStore struct {
	pool *pgxpool.Pool
}

func NewOrgStore(pool *pgxpool.Pool) *OrgStore {
	return &OrgStore{pool: pool}
}

type orgData struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	PostalCode string `json:"postalCode"`
	City       string `json:"city"`
	Country    string `json:"country"`
	VATNumber  string `json:"vatNumber"`
	SIREN      string `json:"siren"`
	SIRET      string `json:"siret"`
	Picture    string `json:"picture,omitempty"`
	Currency   string `json:"currency,omitempty"`
}

// OrganizationFields holds the editable, non-identifying fields of an
// organization (used by Create/Update to avoid long positional args).
type OrganizationFields struct {
	Name       string
	Address    string
	PostalCode string
	City       string
	Country    string
	VATNumber  string
	SIREN      string
	SIRET      string
	Picture    string
	Currency   string
}

// applyOrgData unmarshals the stored JSON blob onto an organization already
// populated with its scanned columns (id, owner, timestamps).
func applyOrgData(o *models.Organization, raw []byte) error {
	var d orgData
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	o.Name = d.Name
	o.Address = d.Address
	o.PostalCode = d.PostalCode
	o.City = d.City
	o.Country = d.Country
	o.VATNumber = d.VATNumber
	o.SIREN = d.SIREN
	o.SIRET = d.SIRET
	o.Picture = d.Picture
	o.Currency = utils.If(utils.IsBlank(d.Currency), models.DefaultCurrency(), d.Currency)
	return nil
}

func scanOrganization(row pgx.Row) (models.Organization, error) {
	var o models.Organization
	var raw []byte
	if err := row.Scan(&o.ID, &o.OwnerID, &raw, &o.CreatedAt, &o.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Organization{}, ErrNotFound
		}
		return models.Organization{}, err
	}
	if err := applyOrgData(&o, raw); err != nil {
		return models.Organization{}, err
	}
	return o, nil
}

func toOrgData(f OrganizationFields) orgData {
	return orgData{
		Name:       f.Name,
		Address:    f.Address,
		PostalCode: f.PostalCode,
		City:       f.City,
		Country:    f.Country,
		VATNumber:  f.VATNumber,
		SIREN:      f.SIREN,
		SIRET:      f.SIRET,
		Picture:    f.Picture,
		Currency:   utils.If(utils.IsBlank(f.Currency), models.DefaultCurrency(), f.Currency),
	}
}

// Create creates the organization and gives its owner an explicit "owner"
// membership row, so they have somewhere to store things like a daily rate
// just like any other member.
func (s *OrgStore) Create(ctx context.Context, ownerID string, f OrganizationFields) (models.Organization, error) {
	data, err := json.Marshal(toOrgData(f))
	if err != nil {
		return models.Organization{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Organization{}, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		INSERT INTO organizations (owner_id, data)
		VALUES ($1, $2)
		RETURNING id, owner_id, data, created_at, updated_at
	`, ownerID, data)
	org, err := scanOrganization(row)
	if err != nil {
		return models.Organization{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, org.ID, ownerID); err != nil {
		return models.Organization{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Organization{}, err
	}
	return org, nil
}

func (s *OrgStore) FindByID(ctx context.Context, id string) (models.Organization, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, owner_id, data, created_at, updated_at
		FROM organizations WHERE id = $1
	`, id)
	return scanOrganization(row)
}

func (s *OrgStore) ListForUser(ctx context.Context, userID string) ([]models.Organization, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT o.id, o.owner_id, o.data, o.created_at, o.updated_at
		FROM organizations o
		LEFT JOIN organization_members m ON m.organization_id = o.id
		WHERE o.owner_id = $1 OR m.user_id = $1
		ORDER BY o.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orgs := []models.Organization{}
	for rows.Next() {
		o, err := scanOrganization(rows)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, rows.Err()
}

// ListAll returns every organization, for the superuser's org-management screen.
// ListAllWithOwner returns every organization with its owner's email, for
// the superuser's organization-management screen.
func (s *OrgStore) ListAllWithOwner(ctx context.Context) ([]models.OrganizationWithOwner, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT o.id, o.owner_id, o.data, o.created_at, o.updated_at, u.email
		FROM organizations o
		JOIN users u ON u.id = o.owner_id
		ORDER BY o.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orgs := []models.OrganizationWithOwner{}
	for rows.Next() {
		var o models.Organization
		var raw []byte
		var ownerEmail string
		if err := rows.Scan(&o.ID, &o.OwnerID, &raw, &o.CreatedAt, &o.UpdatedAt, &ownerEmail); err != nil {
			return nil, err
		}
		if err := applyOrgData(&o, raw); err != nil {
			return nil, err
		}
		orgs = append(orgs, models.OrganizationWithOwner{Organization: o, OwnerEmail: ownerEmail})
	}
	return orgs, rows.Err()
}

func (s *OrgStore) Update(ctx context.Context, id string, f OrganizationFields) (models.Organization, error) {
	data, err := json.Marshal(toOrgData(f))
	if err != nil {
		return models.Organization{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE organizations SET data = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, owner_id, data, created_at, updated_at
	`, id, data)
	return scanOrganization(row)
}

func (s *OrgStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetRole returns the effective role of a user within an organization.
// The owner always resolves to RoleOwner regardless of membership rows,
// as a safety net even though the owner also has an explicit "owner" row.
func (s *OrgStore) GetRole(ctx context.Context, orgID, userID string) (models.Role, error) {
	org, err := s.FindByID(ctx, orgID)
	if err != nil {
		return "", err
	}
	if org.OwnerID == userID {
		return models.RoleOwner, nil
	}

	var role string
	row := s.pool.QueryRow(ctx, `
		SELECT role FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`, orgID, userID)
	if err := row.Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return models.Role(role), nil
}

type memberData struct {
	DailyRate *float64 `json:"dailyRate,omitempty"`
	Currency  string   `json:"currency,omitempty"`
}

func scanMember(row pgx.Row) (models.Member, error) {
	var m models.Member
	var raw []byte
	if err := row.Scan(&m.ID, &m.OrganizationID, &m.UserID, &m.Email, &m.Name, &m.Surname, &m.Role, &raw, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Member{}, ErrNotFound
		}
		return models.Member{}, err
	}
	var d memberData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.Member{}, err
	}
	m.DailyRate = d.DailyRate
	m.Currency = d.Currency
	return m, nil
}

const memberColumns = `m.id, m.organization_id, m.user_id, u.email,
	COALESCE(u.data->>'name', ''), COALESCE(u.data->>'surname', ''),
	m.role, m.data, m.created_at, m.updated_at`

// memberUserLookup fetches a single user's identity fields by id, used by
// mutations below that RETURN a member row without joining organization_members.
const memberUserLookup = `
	(SELECT email FROM users WHERE id = $2),
	(SELECT COALESCE(data->>'name', '') FROM users WHERE id = $2),
	(SELECT COALESCE(data->>'surname', '') FROM users WHERE id = $2),
`

func (s *OrgStore) AddMember(ctx context.Context, orgID, userID string, role models.Role) (models.Member, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, user_id,
			`+memberUserLookup+`
			role, data, created_at, updated_at
	`, orgID, userID, string(role))
	return scanMember(row)
}

func (s *OrgStore) UpdateMemberRole(ctx context.Context, orgID, userID string, role models.Role) (models.Member, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE organization_members SET role = $3, updated_at = now()
		WHERE organization_id = $1 AND user_id = $2
		RETURNING id, organization_id, user_id,
			`+memberUserLookup+`
			role, data, created_at, updated_at
	`, orgID, userID, string(role))
	return scanMember(row)
}

// SetMemberRate sets a member's daily rate and currency.
func (s *OrgStore) SetMemberRate(ctx context.Context, orgID, userID string, dailyRate float64, currency string) (models.Member, error) {
	data, err := json.Marshal(memberData{DailyRate: &dailyRate, Currency: currency})
	if err != nil {
		return models.Member{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE organization_members SET data = $3, updated_at = now()
		WHERE organization_id = $1 AND user_id = $2
		RETURNING id, organization_id, user_id,
			`+memberUserLookup+`
			role, data, created_at, updated_at
	`, orgID, userID, data)
	return scanMember(row)
}

func (s *OrgStore) RemoveMember(ctx context.Context, orgID, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2
	`, orgID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListMembers returns the organization's members, owner included since the
// owner always has an explicit "owner" membership row.
func (s *OrgStore) ListMembers(ctx context.Context, orgID string) ([]models.Member, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+memberColumns+`
		FROM organization_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.organization_id = $1
		ORDER BY m.created_at
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []models.Member{}
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// TransferOwnership makes newOwnerID the organization's owner. The previous
// owner is demoted to admin (keeping meaningful access) and the new owner's
// membership row (which must already exist) is promoted to owner.
func (s *OrgStore) TransferOwnership(ctx context.Context, orgID, previousOwnerID, newOwnerID string) (models.Organization, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Organization{}, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		UPDATE organizations SET owner_id = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, owner_id, data, created_at, updated_at
	`, orgID, newOwnerID)
	org, err := scanOrganization(row)
	if err != nil {
		return models.Organization{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = 'admin', updated_at = now()
	`, orgID, previousOwnerID); err != nil {
		return models.Organization{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = 'owner', updated_at = now()
	`, orgID, newOwnerID); err != nil {
		return models.Organization{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Organization{}, err
	}
	return org, nil
}
