package report

import (
	"fmt"
	"sort"
	"strings"
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
// amount (daily rate * quantity, project rate taking priority over the
// client's - see effectiveDailyRate) exactly as reports do. A project with
// subdivisions contributes one line per subdivision instead of one line for
// the whole project, each getting an even share of that project's
// quantity/amount (the project's amount divided by its number of
// subdivisions) - entries aren't tagged with a subdivision, so this is a
// deliberate even split rather than a per-subdivision breakdown. Lines are
// ordered by project name for a stable, readable invoice.
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

// InvoiceVATTotals sums an invoice's line items - the total without taxes
// is the sum of every line's amount, project or subdivision alike - and
// applies the client's VAT rate on top (0 when the client is VAT-exempt,
// i.e. VATRate <= 0).
func InvoiceVATTotals(items []InvoiceLineItem, client models.Client) (totalHT, totalVAT, totalTTC float64) {
	for _, it := range items {
		totalHT += it.TotalHT
	}
	if client.VATRate > 0 {
		totalVAT = totalHT * client.VATRate / 100
	}
	return totalHT, totalVAT, totalHT + totalVAT
}

// tableUnsafe strips embedded newlines from a cell value - drawTable wraps
// long text on its own (see pdftable.go), so a literal newline would only
// ever be redundant, and body-cell splitting doesn't treat it specially.
var tableUnsafe = strings.NewReplacer("\n", " ", "\r", " ")

func cell(s string) string {
	return tableUnsafe.Replace(s)
}

// formatVAT formats the VAT rate and total as a string, or "no vat" when the
// client is VAT-exempt (VATRate <= 0). When the client is VAT-exempt, the
// VAT discharge motive is appended when set.
func formatVAT(client models.Client, totalVAT float64, currency string) string {
	if client.VATRate <= 0 {
		if utils.IsNotBlank(client.VATDischargeMotive) {
			return "no vat - " + cell(client.VATDischargeMotive)
		}
		return "no vat"
	}

	return fmt.Sprintf("%s %s (%.0f%%)", formatAmount(totalVAT), currency, client.VATRate)
}

// formatContact formats a "Contact" row as "name: email" (ai-instruct-37),
// falling back to a bare email when name is blank, and to "" (dropped
// entirely by addRow) when email itself is blank.
func formatContact(name, email string) string {
	if utils.IsBlank(email) {
		return utils.EMPTY
	}

	if utils.IsBlank(name) {
		return cell(email)
	}

	return cell(fmt.Sprintf("%s: %s", name, email))
}

// ibanRow builds the bank-details row's label/value (ai-instruct-67): both
// IBAN and BIC set combines them under "IBAN / BIC"; IBAN alone still shows
// under plain "IBAN" (a BIC identifies the bank but isn't itself enough to
// receive a transfer, so BIC alone shows nothing). Both blank returns ("",
// "") - addRow drops the row entirely on the blank value regardless of
// label.
func ibanRow(iban, bic string) (label, value string) {
	switch {
	case utils.IsNotBlank(iban) && utils.IsNotBlank(bic):
		return "IBAN / BIC", cell(fmt.Sprintf("%s / %s", iban, bic))
	case utils.IsNotBlank(iban):
		return "IBAN", cell(iban)
	default:
		return utils.EMPTY, utils.EMPTY
	}
}

// formatAddress joins the non-blank parts of an address into one line,
// returning "" (dropped entirely by addRow) when nothing is set.
func formatAddress(address, postalCode, city, country string) string {
	parts := make([]string, 0, 3)
	if utils.IsNotBlank(address) {
		parts = append(parts, address)
	}
	if line := strings.TrimSpace(postalCode + " " + city); utils.IsNotBlank(line) {
		parts = append(parts, line)
	}
	if utils.IsNotBlank(country) {
		parts = append(parts, country)
	}
	return cell(strings.Join(parts, ", "))
}

// addRow appends a {label, value} row, or drops it entirely when value is
// blank, so an organization/client with e.g. no SIRET set doesn't get a
// dangling empty row on its invoice.
func addRow(rows [][]string, label, value string) [][]string {
	if utils.IsBlank(value) {
		return rows
	}
	return append(rows, []string{label, value})
}

var issuerColumns = []tableColumn{{Header: "Issuer", Weight: 30}, {Header: "Details", Weight: 35}}
var clientColumns = []tableColumn{{Header: "Customer", Weight: 30}, {Header: "Details", Weight: 35}}
var lineItemColumns = []tableColumn{
	{Header: "Description", Weight: 50},
	{Header: "Quantity", Weight: 20},
	{Header: "Amount (without taxes)", Weight: 30},
}
var totalsColumns = []tableColumn{{Header: "Totals", Weight: 60}, {Header: "Amount", Weight: 40}}

func issuerRows(org models.Organization, invoiceDate, ownerContact string) [][]string {
	rows := [][]string{}
	rows = addRow(rows, "Name", cell(org.Name))
	rows = addRow(rows, "Address", formatAddress(org.Address, org.PostalCode, org.City, org.Country))
	rows = addRow(rows, "SIREN", cell(org.SIREN))
	rows = addRow(rows, "SIRET", cell(org.SIRET))
	rows = addRow(rows, "VAT / TVA IC", cell(org.VATNumber))
	rows = addRow(rows, "NAF", cell(org.NAF))
	rows = addRow(rows, "Date", invoiceDate)
	rows = addRow(rows, "Contact", ownerContact)
	ibanLabel, ibanValue := ibanRow(org.IBAN, org.BIC)
	rows = addRow(rows, ibanLabel, ibanValue)
	return rows
}

func clientRows(client models.Client) [][]string {
	rows := [][]string{}
	rows = addRow(rows, "Name", cell(client.Name))
	rows = addRow(rows, "Address", formatAddress(client.Address, client.PostalCode, client.City, client.Country))
	rows = addRow(rows, "Contact", formatContact(client.ContactName, client.Email))
	rows = addRow(rows, "VAT / TVA IC", cell(client.VATNumber))
	return rows
}

func lineItemRows(items []InvoiceLineItem) [][]string {
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		rows = append(rows, []string{cell(it.Description), formatAmount(it.Quantity), formatAmount(it.TotalHT)})
	}
	return rows
}

func totalsRows(client models.Client, totalHT, totalVAT, totalTTC, vatRate float64, currency string) [][]string {
	return [][]string{
		{"Total HT (without taxes)", fmt.Sprintf("%s %s", formatAmount(totalHT), currency)},
		{"VAT / TVA", formatVAT(client, totalVAT, currency)},
		{"Total TTC (with taxes)", fmt.Sprintf("%s %s", formatAmount(totalTTC), currency)},
	}
}

// RenderInvoicePDF renders a generated invoice as a PDF: the organization's
// logo top-right (see ResolveLogo), an intro (title/location/date/billed
// period/purchase order), issuer/client info tables, the billable line
// items, and a totals table -
// every table drawn directly with fpdf (see drawTable), exactly like report
// PDFs, rather than through a markdown table, so a long value wraps onto a
// second line inside its cell instead of overflowing it, and a blank
// row/cell can never crash the renderer. Rows with a blank value (e.g. no
// SIRET set) are dropped entirely. When set, the organization's stamp is
// placed below everything.
func RenderInvoicePDF(org models.Organization, client models.Client, owner models.User, invoiceNumber string, items []InvoiceLineItem, totalHT, totalVAT, totalTTC float64, startDay, endDay string) ([]byte, error) {
	renderer := newPdfRenderer("P")
	addFooter(renderer.Pdf)

	logoData, logoType := ResolveLogo(org.Picture)
	if len(logoData) > 0 {
		placeLogo(renderer.Pdf, logoData, logoType)
	}

	invoiceDate := time.Now().Format(USDateLayout)
	ownerName := strings.TrimSpace(owner.Name + " " + owner.Surname)
	ownerContact := formatContact(ownerName, owner.Email)

	intro := fmt.Sprintf(
		"# Invoice N°%s\n\n%s, the %s\n\nPeriod: %s - %s\n\n\n\n",
		cell(invoiceNumber), cell(org.City), invoiceDate, formatUSDate(startDay), formatUSDate(endDay),
	)
	if err := renderer.Run([]byte(intro)); err != nil {
		return nil, err
	}

	// Dropped entirely when blank, rather than printing a dangling "Purchase
	// Order:" line with nothing after it; an explicit Ln (not just a blank
	// markdown line, which mdtopdf collapses) adds real breathing room below
	// it when it is printed.
	if utils.IsNotBlank(client.PurchaseOrder) {
		poLine := fmt.Sprintf("\nPurchase Order: %s\n", cell(client.PurchaseOrder))
		if err := renderer.Run([]byte(poLine)); err != nil {
			return nil, err
		}
		renderer.Pdf.Ln(8)
	}

	translate := renderer.Pdf.UnicodeTranslatorFromDescriptor("cp1252")
	drawTable(renderer.Pdf, translate, issuerColumns, issuerRows(org, invoiceDate, ownerContact))
	renderer.Pdf.Ln(8)
	drawTable(renderer.Pdf, translate, clientColumns, clientRows(client))
	renderer.Pdf.Ln(8)

	if err := renderer.Run([]byte("## Object\n")); err != nil {
		return nil, err
	}
	drawTable(renderer.Pdf, translate, lineItemColumns, lineItemRows(items))
	renderer.Pdf.Ln(8)
	drawTable(renderer.Pdf, translate, totalsColumns, totalsRows(client, totalHT, totalVAT, totalTTC, client.VATRate, org.Currency))

	if utils.IsNotBlank(org.Stamp) {
		if decoded, dt, ok := decodeDataURI(org.Stamp); ok {
			placeStamp(renderer.Pdf, decoded, dt)
		}
	}

	return outputPDF(renderer.Pdf)
}
