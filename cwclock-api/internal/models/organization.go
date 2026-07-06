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
	Currency   string    `json:"currency"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// DefaultCurrency is applied to an organization when none is provided.
const DefaultCurrency = "EUR"

// AllowedCurrencies is the fixed, ordered list of ISO 4217 currency codes an
// organization may be billed in. It's a closed list (not free text) so
// invoices/exports only ever deal with currencies the business supports.
var AllowedCurrencies = []string{
	"EUR", "USD", "GBP", "CAD", "CHF", "TND", "DZD", "MAD", "TRY", "EGP",
	"SAR", "AED", "QAR", "CNY", "HKD", "SGD", "JPY", "AUD", "NZD",
}

// IsAllowedCurrency reports whether code is one of AllowedCurrencies.
func IsAllowedCurrency(code string) bool {
	for _, c := range AllowedCurrencies {
		if c == code {
			return true
		}
	}
	return false
}

// OrganizationWithOwner adds the owner's email to an Organization, for the
// superuser's organization-management screen (which lists orgs the caller
// isn't necessarily a member of, so it can't resolve the owner client-side).
type OrganizationWithOwner struct {
	Organization
	OwnerEmail string `json:"ownerEmail"`
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
