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

// defaultAllowedCurrencies is the fixed, ordered list of ISO 4217 currency
// codes an organization may be billed in, used unless the deployment
// overrides it via the CWCLOCK_ALLOWED_CURRENCIES env var (see
// SetAllowedCurrencies). It's a closed list (not free text) so invoices/
// exports only ever deal with currencies the business supports.
var defaultAllowedCurrencies = []string{
	"EUR", "USD", "GBP", "CAD", "CHF", "TND", "DZD", "MAD", "TRY", "EGP",
	"SAR", "AED", "QAR", "CNY", "HKD", "SGD", "JPY", "AUD", "NZD",
}

// AllowedCurrencies is the effective, ordered list of ISO 4217 currency
// codes organizations may be billed in. It starts out as
// defaultAllowedCurrencies and can be overridden at startup with
// SetAllowedCurrencies.
var AllowedCurrencies = append([]string(nil), defaultAllowedCurrencies...)

// SetAllowedCurrencies overrides AllowedCurrencies, e.g. from the
// CWCLOCK_ALLOWED_CURRENCIES env var. A nil or empty list is ignored, so the
// default list always stays in effect unless a valid override is given.
func SetAllowedCurrencies(codes []string) {
	if len(codes) == 0 {
		return
	}
	AllowedCurrencies = codes
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

// DefaultCurrency is applied to an organization when none is provided: the
// first entry of AllowedCurrencies, so it stays valid even when the list is
// overridden without "EUR" in it.
func DefaultCurrency() string {
	return AllowedCurrencies[0]
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
