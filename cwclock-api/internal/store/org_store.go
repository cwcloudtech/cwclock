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
	pool      *pgxpool.Pool
	countries *CountryStore
}

func NewOrgStore(pool *pgxpool.Pool, countries *CountryStore) *OrgStore {
	return &OrgStore{pool: pool, countries: countries}
}

type orgData struct {
	Name                 string                      `json:"name"`
	AccountingEmail      string                      `json:"accountingEmail,omitempty"`
	Address              string                      `json:"address"`
	PostalCode           string                      `json:"postalCode"`
	City                 string                      `json:"city"`
	Country              string                      `json:"country"`
	VATNumber            string                      `json:"vatNumber"`
	SIREN                string                      `json:"siren"`
	SIRET                string                      `json:"siret"`
	NAF                  string                      `json:"naf,omitempty"`
	MF                   string                      `json:"mf,omitempty"`
	IdentificationNumber string                      `json:"identificationNumber,omitempty"`
	IBAN                 string                      `json:"iban,omitempty"`
	BIC                  string                      `json:"bic,omitempty"`
	Picture              string                      `json:"picture,omitempty"`
	PictureX             *float64                    `json:"pictureX,omitempty"`
	PictureY             *float64                    `json:"pictureY,omitempty"`
	Stamp                string                      `json:"stamp,omitempty"`
	StampX               *float64                    `json:"stampX,omitempty"`
	StampY               *float64                    `json:"stampY,omitempty"`
	Currency             string                      `json:"currency,omitempty"`
	ExternalConnections  []models.ExternalConnection `json:"externalConnections,omitempty"`
}

// OrganizationFields holds the editable, non-identifying fields of an
// organization (used by Create/Update to avoid long positional args).
type OrganizationFields struct {
	Name                 string
	AccountingEmail      string
	Address              string
	PostalCode           string
	City                 string
	Country              string
	VATNumber            string
	SIREN                string
	SIRET                string
	NAF                  string
	MF                   string
	IdentificationNumber string
	IBAN                 string
	BIC                  string
	Picture              string
	PictureX             float64
	PictureY             float64
	Stamp                string
	StampX               float64
	StampY               float64
	Currency             string
	ExternalConnections  []models.ExternalConnection
}

// applyOrgData unmarshals the stored JSON blob onto an organization already
// populated with its scanned columns (id, owner, timestamps). Currency
// falls back to models.FallbackCurrency only defensively, for rows written
// before Create/Update started always resolving and persisting one.
func applyOrgData(o *models.Organization, raw []byte) error {
	var d orgData
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	o.Name = d.Name
	o.AccountingEmail = d.AccountingEmail
	o.Address = d.Address
	o.PostalCode = d.PostalCode
	o.City = d.City
	o.Country = d.Country
	o.VATNumber = d.VATNumber
	o.SIREN = d.SIREN
	o.SIRET = d.SIRET
	o.NAF = d.NAF
	o.MF = d.MF
	o.IdentificationNumber = d.IdentificationNumber
	o.IBAN = d.IBAN
	o.BIC = d.BIC
	o.Picture = d.Picture
	o.PictureX = resolveImagePosition(d.PictureX)
	o.PictureY = resolveImagePosition(d.PictureY)
	o.Stamp = d.Stamp
	o.StampX = resolveImagePosition(d.StampX)
	o.StampY = resolveImagePosition(d.StampY)
	o.Currency = utils.If(utils.IsBlank(d.Currency), models.FallbackCurrency, d.Currency)
	o.ExternalConnections = d.ExternalConnections
	if o.ExternalConnections == nil {
		o.ExternalConnections = []models.ExternalConnection{}
	}
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
		Name:                 f.Name,
		AccountingEmail:      f.AccountingEmail,
		Address:              f.Address,
		PostalCode:           f.PostalCode,
		City:                 f.City,
		Country:              f.Country,
		VATNumber:            f.VATNumber,
		SIREN:                f.SIREN,
		SIRET:                f.SIRET,
		NAF:                  f.NAF,
		MF:                   f.MF,
		IdentificationNumber: f.IdentificationNumber,
		IBAN:                 f.IBAN,
		BIC:                  f.BIC,
		Picture:              f.Picture,
		PictureX:             &f.PictureX,
		PictureY:             &f.PictureY,
		Stamp:                f.Stamp,
		StampX:               &f.StampX,
		StampY:               &f.StampY,
		Currency:             f.Currency,
		ExternalConnections:  f.ExternalConnections,
	}
}

// resolveCurrency keeps an explicit currency as-is; a blank one defaults to
// the country's own currency (ai-instruct-35: "the default currency should
// be selected according to the country but let the user decide"), falling
// back to models.FallbackCurrency when the country can't resolve one either.
func (s *OrgStore) resolveCurrency(ctx context.Context, f OrganizationFields) (string, error) {
	if utils.IsNotBlank(f.Currency) {
		return f.Currency, nil
	}
	return s.countries.DefaultCurrency(ctx, f.Country)
}

// Create creates the organization and gives its owner an explicit "owner"
// membership row, so they have somewhere to store things like a daily rate
// just like any other member.
func (s *OrgStore) Create(ctx context.Context, ownerID string, f OrganizationFields) (models.Organization, error) {
	currency, err := s.resolveCurrency(ctx, f)
	if err != nil {
		return models.Organization{}, err
	}
	f.Currency = currency

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
	currency, err := s.resolveCurrency(ctx, f)
	if err != nil {
		return models.Organization{}, err
	}
	f.Currency = currency

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

// AddExternalConnection appends conn to the organization's stored
// externalConnections list via a single jsonb_set/concat update (rather
// than fetch-modify-Update, which would clobber a concurrent whole-org edit
// touching any other field), per ai-instruct-40's "add a connection should
// automatically save the organization" PATCH endpoint.
func (s *OrgStore) AddExternalConnection(ctx context.Context, id string, conn models.ExternalConnection) (models.Organization, error) {
	connJSON, err := json.Marshal(conn)
	if err != nil {
		return models.Organization{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE organizations
		SET data = jsonb_set(
			data,
			'{externalConnections}',
			COALESCE(data->'externalConnections', '[]'::jsonb) || jsonb_build_array($2::jsonb)
		), updated_at = now()
		WHERE id = $1
		RETURNING id, owner_id, data, created_at, updated_at
	`, id, connJSON)
	return scanOrganization(row)
}

// RemoveExternalConnection removes the connection with the given id from
// the organization's stored externalConnections list, the same atomic
// jsonb-only update as AddExternalConnection (rather than fetch-filter-
// Update, which would clobber a concurrent whole-org edit touching any
// other field). Removing an id that isn't present is a no-op, not an error.
func (s *OrgStore) RemoveExternalConnection(ctx context.Context, id, connID string) (models.Organization, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE organizations
		SET data = jsonb_set(
			data,
			'{externalConnections}',
			(
				SELECT COALESCE(jsonb_agg(elem), '[]'::jsonb)
				FROM jsonb_array_elements(COALESCE(data->'externalConnections', '[]'::jsonb)) elem
				WHERE elem->>'id' != $2
			)
		), updated_at = now()
		WHERE id = $1
		RETURNING id, owner_id, data, created_at, updated_at
	`, id, connID)
	return scanOrganization(row)
}

// IsOwnedBySuperuser reports whether orgID's owner has the superuser global
// role, used to exempt superuser-owned organizations from the monthly
// invoice/export-job mail counter (ai-instruct-83).
func (s *OrgStore) IsOwnedBySuperuser(ctx context.Context, orgID string) (bool, error) {
	var isSuperuser bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM organizations o
			JOIN users u ON u.id = o.owner_id
			WHERE o.id = $1 AND u.data->>'role' = $2
		)
	`, orgID, string(models.GlobalRoleSuperuser)).Scan(&isSuperuser)
	return isSuperuser, err
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
		return utils.EMPTY, err
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
			return utils.EMPTY, ErrNotFound
		}
		return utils.EMPTY, err
	}
	return models.Role(role), nil
}

type memberData struct {
	DailyRate *float64 `json:"dailyRate,omitempty"`
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

// SetMemberRate sets a member's daily rate. Currency isn't stored per
// member - it's always the organization's own currency.
func (s *OrgStore) SetMemberRate(ctx context.Context, orgID, userID string, dailyRate float64) (models.Member, error) {
	data, err := json.Marshal(memberData{DailyRate: &dailyRate})
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

// CountMembersByRole returns the number of organization memberships per
// role (owner/admin/member/reader), for the "counter of users per role"
// metric. A user belonging to several organizations counts once per role
// held, since roles are per-membership rather than global to the user.
func (s *OrgStore) CountMembersByRole(ctx context.Context) (map[string]int64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT role, count(*) FROM organization_members GROUP BY role
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int64{}
	for rows.Next() {
		var role string
		var count int64
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		counts[role] = count
	}
	return counts, rows.Err()
}

// ErrCannotRemoveOwner is returned by RemoveMember when asked to remove the
// organization's own owner - see ai-instruct-32: doing so used to leave
// owner_id pointing at a user with no organization_members row at all,
// which then dropped the owner out of every members-derived list (the
// report/invoice user filter, the entry-reassignment picker, ...).
var ErrCannotRemoveOwner = errors.New("cannot remove the organization owner")

func (s *OrgStore) RemoveMember(ctx context.Context, orgID, userID string) error {
	org, err := s.FindByID(ctx, orgID)
	if err != nil {
		return err
	}
	if org.OwnerID == userID {
		return ErrCannotRemoveOwner
	}

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

// ListMembers returns the organization's members. The owner's own
// membership row is healed back in first if it's ever missing (see
// ErrCannotRemoveOwner - this covers data predating that guard), so the
// owner is always present here and in everything this list feeds.
func (s *OrgStore) ListMembers(ctx context.Context, orgID string) ([]models.Member, error) {
	org, err := s.FindByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
		ON CONFLICT (organization_id, user_id) DO NOTHING
	`, orgID, org.OwnerID); err != nil {
		return nil, err
	}

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
