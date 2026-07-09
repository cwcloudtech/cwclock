package report

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

// InvoiceLineItem is one billable row of an invoice: either a whole
// project, or (when the project has subdivisions) one row per subdivision,
// with the project's quantity/amount split evenly across them.
type InvoiceLineItem struct {
	Description string
	Quantity    float64
	TotalHT     float64
}

// BuildInvoiceLineItems groups entries by project, computing each project's
// billable quantity (hours worked divided by the client's HoursPerDay) and
// amount (quantity * effective daily rate, project rate taking priority
// over the client's - see effectiveDailyRate) exactly as reports do. A
// project with subdivisions contributes one line per subdivision instead of
// one line for the whole project, each getting an even share of that
// project's quantity/amount - entries aren't tagged with a subdivision, so
// this is a deliberate even split rather than a per-subdivision breakdown.
// Lines are ordered by project name for a stable, readable invoice.
func BuildInvoiceLineItems(entries []models.TimeEntry, projects map[string]models.Project, client models.Client) []InvoiceLineItem {
	durationByProject := map[string]int{}
	for _, e := range entries {
		durationByProject[e.ProjectID] += DurationSecs(e, client)
	}

	projectIDs := make([]string, 0, len(durationByProject))
	for id := range durationByProject {
		projectIDs = append(projectIDs, id)
	}
	sort.Slice(projectIDs, func(i, j int) bool {
		return projects[projectIDs[i]].Name < projects[projectIDs[j]].Name
	})

	items := make([]InvoiceLineItem, 0, len(projectIDs))
	for _, id := range projectIDs {
		project := projects[id]
		hours := float64(durationByProject[id]) / 3600
		quantity := hours / hoursPerDay(client)

		rate := effectiveDailyRate(client, project, models.Member{})
		dailyRate := 0.0
		if rate != nil {
			dailyRate = *rate
		}
		totalHT := quantity * dailyRate

		if len(project.Subdivisions) == 0 {
			items = append(items, InvoiceLineItem{Description: project.Name, Quantity: quantity, TotalHT: totalHT})
			continue
		}

		n := float64(len(project.Subdivisions))
		for _, sub := range project.Subdivisions {
			items = append(items, InvoiceLineItem{
				Description: project.Name + " - " + sub,
				Quantity:    quantity / n,
				TotalHT:     totalHT / n,
			})
		}
	}
	return items
}

// InvoiceVATTotals sums an invoice's line items and applies the client's
// VAT rate (0 when the client is VAT-exempt, i.e. VATRate <= 0).
func InvoiceVATTotals(items []InvoiceLineItem, client models.Client) (totalHT, totalVAT, totalTTC float64) {
	for _, it := range items {
		totalHT += it.TotalHT
	}
	if client.VATRate > 0 {
		totalVAT = totalHT * client.VATRate / 100
	}
	return totalHT, totalVAT, totalHT + totalVAT
}

// tableUnsafe strips characters that would corrupt a markdown table cell -
// notably "|", which would otherwise split a name/address/description
// containing one into extra columns.
var tableUnsafe = strings.NewReplacer("|", "/", "\n", " ", "\r", " ")

func cell(s string) string {
	return tableUnsafe.Replace(s)
}

// clientVATLine is the Client table's "TVA IC" cell: the client's VAT
// number when they're charged VAT, or an exemption note (plus the discharge
// motive, when given) when VATRate is 0 or negative.
func clientVATLine(client models.Client) string {
	if client.VATRate <= 0 {
		if utils.IsNotBlank(client.VATDischargeMotive) {
			return "no tva - " + cell(client.VATDischargeMotive)
		}
		return "no tva"
	}
	return cell(client.VATNumber)
}

func lineItemsMarkdown(items []InvoiceLineItem) string {
	var b strings.Builder
	b.WriteString("| Description | Quantité (Quantity) | Montant HT (Amount without taxes) |\n")
	b.WriteString("| ------------ | -------------------- | ---------------------------------- |\n")
	for _, it := range items {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", cell(it.Description), formatAmount(it.Quantity), formatAmount(it.TotalHT))
	}
	return b.String()
}

func formatAddress(address, postalCode, city, country string) string {
	return cell(fmt.Sprintf("%s, %s %s, %s", address, postalCode, city, country))
}

// invoiceMarkdownTpl is the invoice's markdown template, externalized under
// templates/ (with a .tpl.md extension) like the report header, so invoice
// markdown lives in its own reviewable file instead of a Go string literal.
//
//go:embed templates/invoice.tpl.md
var invoiceMarkdownTpl string

var invoiceTemplate = template.Must(template.New("invoice").Parse(invoiceMarkdownTpl))

// invoiceTemplateData is entirely precomputed, sanitized plain strings
// (rather than raw model structs) so every value going into a markdown
// table cell has already been through cell() - matching how reportHeader
// works for the report templates.
type invoiceTemplateData struct {
	InvoiceNumber       string
	InvoiceDate         string
	OrgCity             string
	OrgName             string
	OrgAddress          string
	OrgSIREN            string
	OrgSIRET            string
	OrgVATNumber        string
	OrgNAF              string
	OwnerContact        string
	ClientPurchaseOrder string
	ClientName          string
	ClientAddress       string
	ClientEmail         string
	ClientVATLine       string
	LineItemsMarkdown   string
	TotalHT             string
	TotalVATLine        string
	TotalTTC            string
	Currency            string
}

// RenderInvoicePDF renders a generated invoice as a PDF: the organization's
// logo top-right (see ResolveLogo) and, when set, its stamp image placed
// below the content.
func RenderInvoicePDF(org models.Organization, client models.Client, owner models.User, invoiceNumber string, items []InvoiceLineItem, totalHT, totalVAT, totalTTC float64) ([]byte, error) {
	ownerContact := cell(fmt.Sprintf("%s %s: %s", owner.Surname, owner.Name, owner.Email))
	data := invoiceTemplateData{
		InvoiceNumber:       cell(invoiceNumber),
		InvoiceDate:         time.Now().Format("02/01/2006"),
		OrgCity:             cell(org.City),
		OrgName:             cell(org.Name),
		OrgAddress:          formatAddress(org.Address, org.PostalCode, org.City, org.Country),
		OrgSIREN:            cell(org.SIREN),
		OrgSIRET:            cell(org.SIRET),
		OrgVATNumber:        cell(org.VATNumber),
		OrgNAF:              cell(org.NAF),
		OwnerContact:        ownerContact,
		ClientPurchaseOrder: cell(client.PurchaseOrder),
		ClientName:          cell(client.Name),
		ClientAddress:       formatAddress(client.Address, client.PostalCode, client.City, client.Country),
		ClientEmail:         cell(client.Email),
		ClientVATLine:       clientVATLine(client),
		LineItemsMarkdown:   lineItemsMarkdown(items),
		TotalHT:             formatAmount(totalHT),
		TotalVATLine:        fmt.Sprintf("%s %s (%.0f%%)", formatAmount(totalVAT), org.Currency, client.VATRate),
		TotalTTC:            formatAmount(totalTTC),
		Currency:            org.Currency,
	}

	var buf strings.Builder
	if err := invoiceTemplate.Execute(&buf, data); err != nil {
		return nil, err
	}

	logoData, logoType := ResolveLogo(org.Picture)
	var stampData []byte
	var stampType string
	if org.Stamp != "" {
		if decoded, dt, ok := decodeDataURI(org.Stamp); ok {
			stampData, stampType = decoded, dt
		}
	}

	return RenderMarkdownPDF(buf.String(), logoData, logoType, stampData, stampType)
}
