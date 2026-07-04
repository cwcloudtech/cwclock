package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

var ErrNotFound = errors.New("not found")

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

type userData struct {
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	Surname  string `json:"surname,omitempty"`
	Picture  string `json:"picture,omitempty"`
}

func (s *UserStore) Create(ctx context.Context, email, passwordHash, name, surname string) (models.User, error) {
	data, err := json.Marshal(userData{Password: passwordHash, Name: name, Surname: surname})
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
	return u, nil
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	var raw []byte
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, data, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
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
	u.Picture = d.Picture
	return u, nil
}

func (s *UserStore) FindByID(ctx context.Context, id string) (models.User, error) {
	var u models.User
	var raw []byte
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, data, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
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
	u.Name = d.Name
	u.Surname = d.Surname
	u.Picture = d.Picture
	return u, nil
}

// UpdatePicture sets the user's avatar picture (base64) via jsonb_set, so it
// can't clobber the password hash stored alongside it in the same column.
func (s *UserStore) UpdatePicture(ctx context.Context, id, picture string) (models.User, error) {
	var u models.User
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET data = jsonb_set(data, '{picture}', to_jsonb($2::text), true), updated_at = now()
		WHERE id = $1
		RETURNING id, email, created_at, updated_at
	`, id, picture)
	if err := row.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	u.Picture = picture
	return u, nil
}

// UpdateProfile sets the user's name and surname via jsonb_set, leaving the
// password hash and picture stored alongside them untouched.
func (s *UserStore) UpdateProfile(ctx context.Context, id, name, surname string) (models.User, error) {
	var u models.User
	var raw []byte
	row := s.pool.QueryRow(ctx, `
		UPDATE users
		SET data = jsonb_set(jsonb_set(data, '{name}', to_jsonb($2::text), true), '{surname}', to_jsonb($3::text), true),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, email, data, created_at, updated_at
	`, id, name, surname)
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
	u.Name = d.Name
	u.Surname = d.Surname
	u.Picture = d.Picture
	return u, nil
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
		var u models.User
		var raw []byte
		if err := rows.Scan(&u.ID, &u.Email, &raw, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		var d userData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, err
		}
		u.Name = d.Name
		u.Surname = d.Surname
		users = append(users, u)
	}
	return users, rows.Err()
}
