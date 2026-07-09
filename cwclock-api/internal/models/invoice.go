package models

import "time"

// InvoiceStatus tracks an invoice's payment lifecycle.
type InvoiceStatus string

const (
	InvoiceStatusUnpaid   InvoiceStatus = "unpaid"
	InvoiceStatusPaid     InvoiceStatus = "paid"
	InvoiceStatusCanceled InvoiceStatus = "canceled"
	InvoiceStatusRefunded InvoiceStatus = "refunded"
)

// IsValidInvoiceStatus reports whether status is one of the known invoice
// statuses.
func IsValidInvoiceStatus(status string) bool {
	switch InvoiceStatus(status) {
	case InvoiceStatusUnpaid, InvoiceStatusPaid, InvoiceStatusCanceled, InvoiceStatusRefunded:
		return true
	default:
		return false
	}
}

// Invoice is a generated, saved invoice for one client over a selected
// period. Its rendered PDF is stored alongside it (see InvoiceStore.GetPDF)
// but is deliberately not part of this struct - it's large binary data
// fetched only by the dedicated download endpoint, never inlined into a
// list/JSON response.
type Invoice struct {
	ID                string        `json:"id"`
	OrganizationID    string        `json:"organizationId"`
	ClientID          string        `json:"clientId"`
	Number            string        `json:"number"`
	Status            InvoiceStatus `json:"status"`
	TotalHT           float64       `json:"totalHT"`
	TotalVAT          float64       `json:"totalVAT"`
	TotalTTC          float64       `json:"totalTTC"`
	SelectedBeginDate string        `json:"selectedBeginDate"`
	SelectedEndDate   string        `json:"selectedEndDate"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}
