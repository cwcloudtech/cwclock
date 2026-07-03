package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Picture      string    `json:"picture,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type UserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Token   string `json:"token"`
	Picture string `json:"picture,omitempty"`
}

type UserMeResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Picture   string    `json:"picture,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
