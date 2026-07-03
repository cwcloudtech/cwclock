package models

import "time"

type TimeEntry struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ClientID       string    `json:"clientId"`
	ProjectID      string    `json:"projectId"`
	UserID         string    `json:"userId"`
	Text           string    `json:"text"`
	Status         bool      `json:"status"`
	Day            string    `json:"day"`
	Start          *string   `json:"start"`
	End            *string   `json:"end"`
	AllDay         bool      `json:"allDay"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
