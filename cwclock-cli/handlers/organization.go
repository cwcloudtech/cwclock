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

type OrganizationFields struct {
	Name                 string
	AccountingEmail      string
	Address              string
	PostalCode           string
	City                 string
	Country              string
	VATNumber            string
	SIREN                string
	SIRET                string
	NAF                  string
	MF                   string
	IdentificationNumber string
	IBAN                 string
	BIC                  string
	Currency             string
}

func (f OrganizationFields) toPayload() client.OrganizationPayload {
	return client.OrganizationPayload{
		Name:                 f.Name,
		AccountingEmail:      f.AccountingEmail,
		Address:              f.Address,
		PostalCode:           f.PostalCode,
		City:                 f.City,
		Country:              f.Country,
		VATNumber:            f.VATNumber,
		SIREN:                f.SIREN,
		SIRET:                f.SIRET,
		NAF:                  f.NAF,
		MF:                   f.MF,
		IdentificationNumber: f.IdentificationNumber,
		IBAN:                 f.IBAN,
		BIC:                  f.BIC,
		Currency:             f.Currency,
	}
}

func mergeOrganizationFields(current client.Organization, fields OrganizationFields, changed map[string]bool) client.OrganizationPayload {
	payload := client.OrganizationPayload{
		Name:                 current.Name,
		AccountingEmail:      current.AccountingEmail,
		Address:              current.Address,
		PostalCode:           current.PostalCode,
		City:                 current.City,
		Country:              current.Country,
		VATNumber:            current.VATNumber,
		SIREN:                current.SIREN,
		SIRET:                current.SIRET,
		NAF:                  current.NAF,
		MF:                   current.MF,
		IdentificationNumber: current.IdentificationNumber,
		IBAN:                 current.IBAN,
		BIC:                  current.BIC,
		Currency:             current.Currency,
	}

	if changed["name"] {
		payload.Name = fields.Name
	}
	if changed["accounting-email"] {
		payload.AccountingEmail = fields.AccountingEmail
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
	if changed["iban"] {
		payload.IBAN = fields.IBAN
	}
	if changed["bic"] {
		payload.BIC = fields.BIC
	}
	if changed["currency"] {
		payload.Currency = fields.Currency
	}

	return payload
}

func HandleOrganizationList(formatOverride string) error {
	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	orgs, err := cli.ListOrganizations()
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(orgs)
		return nil
	}

	displayOrganizationsAsTable(orgs)
	return nil
}

func HandleOrganizationCreate(fields OrganizationFields, formatOverride string) error {
	if utils.IsBlank(fields.Name) {
		return fmt.Errorf("name is required: use --name")
	}
	if utils.IsBlank(fields.Country) {
		return fmt.Errorf("country is required: use --country")
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	org, err := cli.CreateOrganization(fields.toPayload())
	if err != nil {
		return err
	}

	renderOrganization(org, formatOverride)
	return nil
}

func HandleOrganizationUpdate(id string, fields OrganizationFields, changed map[string]bool, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("organization id is required: use -i or --id")
	}

	current, err := resolveOrganization(trimmedID)
	if err != nil {
		return err
	}

	payload := mergeOrganizationFields(current, fields, changed)
	if utils.IsBlank(payload.Name) {
		return fmt.Errorf("name is required: use --name")
	}
	if utils.IsBlank(payload.Country) {
		return fmt.Errorf("country is required: use --country")
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	updated, err := cli.UpdateOrganization(current.ID, payload)
	if err != nil {
		return err
	}

	renderOrganization(updated, formatOverride)
	return nil
}

func HandleOrganizationDelete(id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("organization id is required: use -i or --id")
	}

	current, err := resolveOrganization(trimmedID)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.DeleteOrganization(current.ID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", current.ID)
	return nil
}

func renderOrganization(org client.Organization, formatOverride string) {
	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(org)
		return
	}
	displayOrganizationsAsTable([]client.Organization{org})
}

func displayOrganizationsAsTable(orgs []client.Organization) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Country", "Currency"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(orgs) {
		table.Append([]string{"No organizations available", utils.EMPTY, utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for _, org := range orgs {
		table.Append([]string{org.ID, org.Name, org.Country, org.Currency})
	}
	table.Render()
}
