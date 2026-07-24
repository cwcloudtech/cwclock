package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

var ErrNotFound = errors.New("not found")

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

type userData struct {
	Password            string   `json:"password"`
	Name                string   `json:"name,omitempty"`
	Surname             string   `json:"surname,omitempty"`
	Role                string   `json:"role,omitempty"`
	Picture             string   `json:"picture,omitempty"`
	PictureX            *float64 `json:"pictureX,omitempty"`
	PictureY            *float64 `json:"pictureY,omitempty"`
	MFAEnabled          bool     `json:"mfaEnabled,omitempty"`
	MFATOTPSecret       string   `json:"mfaTotpSecret,omitempty"`
	CalendarFeedToken   string   `json:"calendarFeedToken,omitempty"`
	CalendarFeedEnabled bool     `json:"calendarFeedEnabled,omitempty"`
}

// defaultImagePosition centers a picture/stamp when no position was ever
// stored for it (a never-repositioned image, or one saved before this field
// existed).
const defaultImagePosition = 50.0

func resolveImagePosition(v *float64) float64 {
	if v == nil {
		return defaultImagePosition
	}
	return *v
}

func scanUser(row pgx.Row) (models.User, error) {
	var u models.User
	var raw []byte
	if err := row.Scan(&u.ID, &u.Email, &raw, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	var d userData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.User{}, err
	}
	u.PasswordHash = d.Password
	u.Name = d.Name
	u.Surname = d.Surname
	u.Role = models.GlobalRole(d.Role)
	u.Picture = d.Picture
	u.PictureX = resolveImagePosition(d.PictureX)
	u.PictureY = resolveImagePosition(d.PictureY)
	u.MFAEnabled = d.MFAEnabled
	u.MFATOTPSecret = d.MFATOTPSecret
	u.CalendarFeedToken = d.CalendarFeedToken
	u.CalendarFeedEnabled = d.CalendarFeedEnabled
	return u, nil
}

// Count returns the total number of registered users, used to decide
// whether a newly registering user is the very first (and thus superuser).
func (s *UserStore) Count(ctx context.Context) (int, error) {
	var count int
	row := s.pool.QueryRow(ctx, `SELECT count(*) FROM users`)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// insertUser inserts a brand-new account with an already-decided role.
func (s *UserStore) insertUser(ctx context.Context, email, passwordHash, name, surname string, role models.GlobalRole) (models.User, error) {
	data, err := json.Marshal(userData{Password: passwordHash, Name: name, Surname: surname, Role: string(role)})
	if err != nil {
		return models.User{}, err
	}

	var u models.User
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, data)
		VALUES ($1, $2)
		RETURNING id, email, created_at, updated_at
	`, email, data)
	if err := row.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return models.User{}, err
	}
	u.PasswordHash = passwordHash
	u.Name = name
	u.Surname = surname
	u.Role = role
	return u, nil
}

// Create registers a user. The very first account ever created becomes the
// superuser; every other account starts disabled until the superuser
// confirms it (or, in activation mode "email", until they follow their
// emailed confirmation link - see handlers.UserHandler.Register).
func (s *UserStore) Create(ctx context.Context, email, passwordHash, name, surname string) (models.User, error) {
	count, err := s.Count(ctx)
	if err != nil {
		return models.User{}, err
	}
	role := utils.If(count == 0, models.GlobalRoleSuperuser, models.GlobalRoleDisabled)
	return s.insertUser(ctx, email, passwordHash, name, surname, role)
}

// FindOrCreateOIDC logs in a user authenticated via an OIDC provider: an
// existing account is matched by email (linking it regardless of how it was
// originally created), otherwise a new one is registered with no password
// hash. The very first account ever created still becomes superuser; a
// later one is confirmed immediately when activationMode is "email" (the
// identity provider already verified the address, so there's nothing left
// to confirm by email, unlike a password registration), otherwise it starts
// disabled and needs the superuser's approval like every other mode.
func (s *UserStore) FindOrCreateOIDC(ctx context.Context, email, name, surname, activationMode string) (models.User, error) {
	user, err := s.FindByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return models.User{}, err
	}

	count, err := s.Count(ctx)
	if err != nil {
		return models.User{}, err
	}
	role := models.GlobalRoleDisabled
	switch {
	case count == 0:
		role = models.GlobalRoleSuperuser
	case activationMode == models.ActivationModeEmail:
		role = models.GlobalRoleConfirmed
	}
	return s.insertUser(ctx, email, utils.EMPTY, name, surname, role)
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (models.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, data, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	return scanUser(row)
}

func (s *UserStore) FindByID(ctx context.Context, id string) (models.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, data, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

// List returns every registered user, for the superuser's user management screen.
func (s *UserStore) List(ctx context.Context) ([]models.User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, email, data, created_at, updated_at
		FROM users
		ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdatePicture sets the user's avatar picture (base64) and its x/y display
// position via a shallow merge, so it can't clobber the password hash and
// other fields stored alongside it in the same column.
func (s *UserStore) UpdatePicture(ctx context.Context, id, picture string, x, y float64) (models.User, error) {
	patch, err := json.Marshal(map[string]any{"picture": picture, "pictureX": x, "pictureY": y})
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, patch)
	return scanUser(row)
}

// UpdateProfile sets the user's name and surname, and optionally their
// password hash (nil leaves the current password untouched), leaving the
// picture stored alongside them untouched either way.
func (s *UserStore) UpdateProfile(ctx context.Context, id, name, surname string, passwordHash *string) (models.User, error) {
	patch := map[string]any{"name": name, "surname": surname}
	if passwordHash != nil {
		patch["password"] = *passwordHash
	}
	data, err := json.Marshal(patch)
	if err != nil {
		return models.User{}, err
	}

	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, data)
	return scanUser(row)
}

// Confirm sets a disabled account's role to confirmed, used when a user
// follows their emailed confirmation link (activation mode "email").
func (s *UserStore) Confirm(ctx context.Context, id string) (models.User, error) {
	patch, err := json.Marshal(map[string]any{"role": string(models.GlobalRoleConfirmed)})
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, patch)
	return scanUser(row)
}

// SetPendingTOTPSecret stores a freshly generated TOTP secret for id without
// enabling MFA yet - it only takes effect once ConfirmTOTP verifies the user
// actually scanned it, so an abandoned setup never locks anyone into MFA.
func (s *UserStore) SetPendingTOTPSecret(ctx context.Context, id, secret string) (models.User, error) {
	patch, err := json.Marshal(map[string]any{"mfaTotpSecret": secret})
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, patch)
	return scanUser(row)
}

// ConfirmTOTP turns on MFA for id once its pending TOTP secret (set by
// SetPendingTOTPSecret) has been verified.
func (s *UserStore) ConfirmTOTP(ctx context.Context, id string) (models.User, error) {
	return s.SetMFAEnabled(ctx, id, true)
}

// DisableTOTP removes id's TOTP secret and, when keepEnabled is false (the
// user has no other MFA factor left), turns MFA back off entirely.
func (s *UserStore) DisableTOTP(ctx context.Context, id string, keepEnabled bool) (models.User, error) {
	patch, err := json.Marshal(map[string]any{"mfaTotpSecret": "", "mfaEnabled": keepEnabled})
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, patch)
	return scanUser(row)
}

// SetMFAEnabled sets id's aggregate MFA flag, used directly by WebAuthn
// enrollment/removal (to reflect whether any factor is still registered) and
// by the superuser's disable-MFA action.
func (s *UserStore) SetMFAEnabled(ctx context.Context, id string, enabled bool) (models.User, error) {
	patch, err := json.Marshal(map[string]any{"mfaEnabled": enabled})
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, patch)
	return scanUser(row)
}

// SetCalendarFeedEnabled turns the caller's calendar sharing feed on or off.
// Enabling it for the first time also mints a token (kept stable across
// later enable/disable toggles, so a URL the user already shared with
// Outlook/Google keeps working once re-enabled); disabling leaves the token
// in place but FindByCalendarFeedToken/the feed handler must still refuse it.
func (s *UserStore) SetCalendarFeedEnabled(ctx context.Context, id string, enabled bool) (models.User, error) {
	patch := map[string]any{"calendarFeedEnabled": enabled}
	if enabled {
		user, err := s.FindByID(ctx, id)
		if err != nil {
			return models.User{}, err
		}
		if utils.IsBlank(user.CalendarFeedToken) {
			token, err := generateToken()
			if err != nil {
				return models.User{}, err
			}
			patch["calendarFeedToken"] = token
		}
	}
	data, err := json.Marshal(patch)
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, data)
	return scanUser(row)
}

// RegenerateCalendarFeedToken replaces the caller's calendar feed token,
// invalidating any URL previously shared with Outlook/Google Calendar.
func (s *UserStore) RegenerateCalendarFeedToken(ctx context.Context, id string) (models.User, error) {
	token, err := generateToken()
	if err != nil {
		return models.User{}, err
	}
	patch, err := json.Marshal(map[string]any{"calendarFeedToken": token})
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, patch)
	return scanUser(row)
}

// FindByCalendarFeedToken resolves the public ICS feed URL's token back to
// its owner. Disabled feeds are reported as ErrNotFound too (not a separate
// "forbidden"), so a disabled feed's URL looks indistinguishable from an
// unknown one to whoever's polling it.
func (s *UserStore) FindByCalendarFeedToken(ctx context.Context, token string) (models.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, data, created_at, updated_at
		FROM users WHERE data->>'calendarFeedToken' = $1
	`, token)
	user, err := scanUser(row)
	if err != nil {
		return models.User{}, err
	}
	if !user.CalendarFeedEnabled {
		return models.User{}, ErrNotFound
	}
	return user, nil
}

// ErrDuplicateEmail is returned when an admin edit would collide with
// another account's email address.
var ErrDuplicateEmail = errors.New("email already in use")

// AdminUserFields holds the fields the superuser may set on any account.
// PasswordHash and Picture are pointers so a nil value means "leave
// unchanged" (the superuser edit form never displays the current password).
type AdminUserFields struct {
	Email        string
	Name         string
	Surname      string
	Role         string
	PasswordHash *string
	Picture      *string
	PictureX     *float64
	PictureY     *float64
}

// AdminUpdate lets the superuser edit any account's email, profile, role,
// picture and/or password. It merges only the provided keys into the stored
// JSON so omitted fields (like an unset password) are left untouched.
func (s *UserStore) AdminUpdate(ctx context.Context, id string, f AdminUserFields) (models.User, error) {
	patch := map[string]any{
		"name":    f.Name,
		"surname": f.Surname,
		"role":    f.Role,
	}
	if f.PasswordHash != nil {
		patch["password"] = *f.PasswordHash
	}
	if f.Picture != nil {
		patch["picture"] = *f.Picture
	}
	if f.PictureX != nil {
		patch["pictureX"] = *f.PictureX
	}
	if f.PictureY != nil {
		patch["pictureY"] = *f.PictureY
	}
	data, err := json.Marshal(patch)
	if err != nil {
		return models.User{}, err
	}

	row := s.pool.QueryRow(ctx, `
		UPDATE users SET email = $3, data = data || $2::jsonb, updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, data, f.Email)
	user, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, ErrDuplicateEmail
		}
		return models.User{}, err
	}
	return user, nil
}

// Delete removes an account entirely (cascading to organizations it owns,
// per the schema's ON DELETE CASCADE).
func (s *UserStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SearchByEmail returns users whose email contains query, for invite
// autocomplete. It is capped by limit and ordered alphabetically.
func (s *UserStore) SearchByEmail(ctx context.Context, query string, limit int) ([]models.User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, email, data, created_at, updated_at
		FROM users
		WHERE email ILIKE '%' || $1 || '%'
		ORDER BY email
		LIMIT $2
	`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// CreateDisabled registers a user directly as disabled with no password hash,
// used when importing time entries for an unknown user so they exist in the
// system but cannot log in until confirmed and given a password.
func (s *UserStore) CreateDisabled(ctx context.Context, email, name, surname string) (models.User, error) {
	data, err := json.Marshal(userData{Name: name, Surname: surname, Role: string(models.GlobalRoleDisabled)})
	if err != nil {
		return models.User{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, data)
		VALUES ($1, $2)
		RETURNING id, email, data, created_at, updated_at
	`, email, data)
	return scanUser(row)
}
