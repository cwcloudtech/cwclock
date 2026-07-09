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
	Password string   `json:"password"`
	Name     string   `json:"name,omitempty"`
	Surname  string   `json:"surname,omitempty"`
	Role     string   `json:"role,omitempty"`
	Picture  string   `json:"picture,omitempty"`
	PictureX *float64 `json:"pictureX,omitempty"`
	PictureY *float64 `json:"pictureY,omitempty"`
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

// Create registers a user. The very first account ever created becomes the
// superuser; every other account starts disabled until the superuser
// confirms it.
func (s *UserStore) Create(ctx context.Context, email, passwordHash, name, surname string) (models.User, error) {
	count, err := s.Count(ctx)
	if err != nil {
		return models.User{}, err
	}
	role := utils.If(count == 0, models.GlobalRoleSuperuser, models.GlobalRoleDisabled)

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
