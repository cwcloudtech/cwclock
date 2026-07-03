package models

import "time"

type Client struct {
	ID                 string    `json:"id"`
	OrganizationID     string    `json:"organizationId"`
	Name               string    `json:"name"`
	Address            string    `json:"address"`
	PostalCode         string    `json:"postalCode"`
	City               string    `json:"city"`
	Country            string    `json:"country"`
	VATNumber          string    `json:"vatNumber"`
	VATRate            float64   `json:"vatRate"`
	VATDischargeMotive string    `json:"vatDischargeMotive"`
	PurchaseOrder      string    `json:"purchaseOrder"`
	HoursPerDay        float64   `json:"hoursPerDay"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}
