package models

import "time"

type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ClientID       string    `json:"clientId"`
	Name           string    `json:"name"`
	Color          string    `json:"color"`
	DailyRate      *float64  `json:"dailyRate,omitempty"`
	Subdivisions   []string  `json:"subdivisions,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
