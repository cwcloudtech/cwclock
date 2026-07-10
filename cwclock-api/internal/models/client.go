package models

import "time"

type Client struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organizationId"`
	Name                 string    `json:"name"`
	Email                string    `json:"email,omitempty"`
	ContactName          string    `json:"contactName"`
	Address              string    `json:"address"`
	PostalCode           string    `json:"postalCode"`
	City                 string    `json:"city"`
	Country              string    `json:"country"`
	VATNumber            string    `json:"vatNumber"`
	VATRate              float64   `json:"vatRate"`
	VATDischargeMotive   string    `json:"vatDischargeMotive"`
	SIREN                string    `json:"siren"`
	SIRET                string    `json:"siret"`
	NAF                  string    `json:"naf"`
	MF                   string    `json:"mf"`
	IdentificationNumber string    `json:"identificationNumber"`
	PurchaseOrder        string    `json:"purchaseOrder"`
	HoursPerDay          float64   `json:"hoursPerDay"`
	DailyRate            *float64  `json:"dailyRate,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}
