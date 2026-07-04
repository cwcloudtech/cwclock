package models

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleReader Role = "reader"
)

type Organization struct {
	ID         string    `json:"id"`
	OwnerID    string    `json:"ownerId"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	PostalCode string    `json:"postalCode"`
	City       string    `json:"city"`
	Country    string    `json:"country"`
	VATNumber  string    `json:"vatNumber"`
	SIREN      string    `json:"siren"`
	SIRET      string    `json:"siret"`
	Picture    string    `json:"picture,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Member struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	UserID         string    `json:"userId"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Surname        string    `json:"surname"`
	Role           Role      `json:"role"`
	DailyRate      *float64  `json:"dailyRate,omitempty"`
	Currency       string    `json:"currency,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
