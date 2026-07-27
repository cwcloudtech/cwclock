package handlers

import (
	"cwclock/client"
	"cwclock/config"
	"cwclock/utils"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type OrgClientFields struct {
	Name                   string
	Email                  string
	InvoiceEmails          string
	ContactName            string
	Address                string
	PostalCode             string
	City                   string
	Country                string
	VATNumber              string
	VATRate                float64
	VATDischargeMotive     string
	SIREN                  string
	SIRET                  string
	NAF                    string
	MF                     string
	IdentificationNumber   string
	PurchaseOrder          string
	HoursPerDay            float64
	DailyRate              float64
	SendReportsWithInvoice bool
}

func (f OrgClientFields) toPayload(changed map[string]bool) client.OrgClientPayload {
	payload := client.OrgClientPayload{
		Name:                   f.Name,
		Email:                  f.Email,
		InvoiceEmails:          f.InvoiceEmails,
		ContactName:            f.ContactName,
		Address:                f.Address,
		PostalCode:             f.PostalCode,
		City:                   f.City,
		Country:                f.Country,
		VATNumber:              f.VATNumber,
		VATDischargeMotive:     f.VATDischargeMotive,
		SIREN:                  f.SIREN,
		SIRET:                  f.SIRET,
		NAF:                    f.NAF,
		MF:                     f.MF,
		IdentificationNumber:   f.IdentificationNumber,
		PurchaseOrder:          f.PurchaseOrder,
		HoursPerDay:            f.HoursPerDay,
		SendReportsWithInvoice: f.SendReportsWithInvoice,
	}
	if changed["vat-rate"] {
		payload.VATRate = &f.VATRate
	}
	if changed["daily-rate"] {
		payload.DailyRate = &f.DailyRate
	}
	return payload
}

func mergeOrgClientFields(current client.OrgClient, fields OrgClientFields, changed map[string]bool) client.OrgClientPayload {
	currentVATRate := current.VATRate
	payload := client.OrgClientPayload{
		Name:                   current.Name,
		Email:                  current.Email,
		InvoiceEmails:          current.InvoiceEmails,
		ContactName:            current.ContactName,
		Address:                current.Address,
		PostalCode:             current.PostalCode,
		City:                   current.City,
		Country:                current.Country,
		VATNumber:              current.VATNumber,
		VATRate:                &currentVATRate,
		VATDischargeMotive:     current.VATDischargeMotive,
		SIREN:                  current.SIREN,
		SIRET:                  current.SIRET,
		NAF:                    current.NAF,
		MF:                     current.MF,
		IdentificationNumber:   current.IdentificationNumber,
		PurchaseOrder:          current.PurchaseOrder,
		HoursPerDay:            current.HoursPerDay,
		DailyRate:              current.DailyRate,
		SendReportsWithInvoice: current.SendReportsWithInvoice,
	}

	if changed["name"] {
		payload.Name = fields.Name
	}
	if changed["email"] {
		payload.Email = fields.Email
	}
	if changed["invoice-emails"] {
		payload.InvoiceEmails = fields.InvoiceEmails
	}
	if changed["contact-name"] {
		payload.ContactName = fields.ContactName
	}
	if changed["address"] {
		payload.Address = fields.Address
	}
	if changed["postal-code"] {
		payload.PostalCode = fields.PostalCode
	}
	if changed["city"] {
		payload.City = fields.City
	}
	if changed["country"] {
		payload.Country = fields.Country
	}
	if changed["vat-number"] {
		payload.VATNumber = fields.VATNumber
	}
	if changed["vat-rate"] {
		payload.VATRate = &fields.VATRate
	}
	if changed["vat-discharge-motive"] {
		payload.VATDischargeMotive = fields.VATDischargeMotive
	}
	if changed["siren"] {
		payload.SIREN = fields.SIREN
	}
	if changed["siret"] {
		payload.SIRET = fields.SIRET
	}
	if changed["naf"] {
		payload.NAF = fields.NAF
	}
	if changed["mf"] {
		payload.MF = fields.MF
	}
	if changed["identification-number"] {
		payload.IdentificationNumber = fields.IdentificationNumber
	}
	if changed["purchase-order"] {
		payload.PurchaseOrder = fields.PurchaseOrder
	}
	if changed["hours-per-day"] {
		payload.HoursPerDay = fields.HoursPerDay
	}
	if changed["daily-rate"] {
		payload.DailyRate = &fields.DailyRate
	}
	if changed["send-reports-with-invoice"] {
		payload.SendReportsWithInvoice = fields.SendReportsWithInvoice
	}

	return payload
}

func HandleOrgClientList(orgOverride string, formatOverride string) error {
	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	clients, err := cli.ListOrgClients(orgID)
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(clients)
		return nil
	}

	displayOrgClientsAsTable(clients)
	return nil
}

func HandleOrgClientCreate(orgOverride string, fields OrgClientFields, changed map[string]bool, formatOverride string) error {
	if utils.IsBlank(fields.Name) {
		return fmt.Errorf("name is required: use --name")
	}
	if utils.IsBlank(fields.Country) {
		return fmt.Errorf("country is required: use --country")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	orgClient, err := cli.CreateOrgClient(orgID, fields.toPayload(changed))
	if err != nil {
		return err
	}

	renderOrgClient(orgClient, formatOverride)
	return nil
}

func HandleOrgClientUpdate(orgOverride string, id string, fields OrgClientFields, changed map[string]bool, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("client id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	clients, err := cli.ListOrgClients(orgID)
	if err != nil {
		return err
	}

	current, found := findOrgClient(clients, trimmedID)
	if !found {
		return fmt.Errorf("client %q not found", trimmedID)
	}

	payload := mergeOrgClientFields(current, fields, changed)
	if utils.IsBlank(payload.Name) {
		return fmt.Errorf("name is required: use --name")
	}
	if utils.IsBlank(payload.Country) {
		return fmt.Errorf("country is required: use --country")
	}

	updated, err := cli.UpdateOrgClient(orgID, trimmedID, payload)
	if err != nil {
		return err
	}

	renderOrgClient(updated, formatOverride)
	return nil
}

func HandleOrgClientDelete(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("client id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.DeleteOrgClient(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func findOrgClient(clients []client.OrgClient, id string) (client.OrgClient, bool) {
	for _, c := range clients {
		if c.ID == id {
			return c, true
		}
	}
	return client.OrgClient{}, false
}

func renderOrgClient(orgClient client.OrgClient, formatOverride string) {
	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(orgClient)
		return
	}
	displayOrgClientsAsTable([]client.OrgClient{orgClient})
}

func displayOrgClientsAsTable(clients []client.OrgClient) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Country", "Email"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(clients) {
		table.Append([]string{"No clients available", utils.EMPTY, utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for _, c := range clients {
		table.Append([]string{c.ID, c.Name, c.Country, c.Email})
	}
	table.Render()
}
