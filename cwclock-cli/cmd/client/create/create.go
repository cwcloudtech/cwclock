package create

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID                  string
	name                   string
	email                  string
	invoiceEmails          string
	contactName            string
	address                string
	postalCode             string
	city                   string
	country                string
	vatNumber              string
	vatRate                float64
	vatDischargeMotive     string
	siren                  string
	siret                  string
	naf                    string
	mf                     string
	identificationNumber   string
	purchaseOrder          string
	hoursPerDay            float64
	dailyRate              float64
	sendReportsWithInvoice bool
	format                 string
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a client",
	Long:  `This command lets you create a new client for an organization.`,
	Run: func(cmd *cobra.Command, args []string) {
		changed := map[string]bool{
			"vat-rate":   cmd.Flags().Changed("vat-rate"),
			"daily-rate": cmd.Flags().Changed("daily-rate"),
		}

		fields := handlers.OrgClientFields{
			Name:                   name,
			Email:                  email,
			InvoiceEmails:          invoiceEmails,
			ContactName:            contactName,
			Address:                address,
			PostalCode:             postalCode,
			City:                   city,
			Country:                country,
			VATNumber:              vatNumber,
			VATRate:                vatRate,
			VATDischargeMotive:     vatDischargeMotive,
			SIREN:                  siren,
			SIRET:                  siret,
			NAF:                    naf,
			MF:                     mf,
			IdentificationNumber:   identificationNumber,
			PurchaseOrder:          purchaseOrder,
			HoursPerDay:            hoursPerDay,
			DailyRate:              dailyRate,
			SendReportsWithInvoice: sendReportsWithInvoice,
		}
		err := handlers.HandleOrgClientCreate(orgID, fields, changed, format)
		utils.ExitIfError(err)
	},
}

func init() {
	CreateCmd.DisableFlagsInUseLine = true
	CreateCmd.Flags().StringVarP(&orgID, "org", "o", "", "Organization ID (overrides configured org_id)")
	CreateCmd.Flags().StringVar(&name, "name", "", "Client name (required)")
	CreateCmd.Flags().StringVar(&country, "country", "", "Country code (required)")
	CreateCmd.Flags().StringVar(&email, "email", "", "Email")
	CreateCmd.Flags().StringVar(&invoiceEmails, "invoice-emails", "", "Comma/semicolon-separated invoice recipient emails")
	CreateCmd.Flags().StringVar(&contactName, "contact-name", "", "Contact name")
	CreateCmd.Flags().StringVar(&address, "address", "", "Address")
	CreateCmd.Flags().StringVar(&postalCode, "postal-code", "", "Postal code")
	CreateCmd.Flags().StringVar(&city, "city", "", "City")
	CreateCmd.Flags().StringVar(&vatNumber, "vat-number", "", "VAT number")
	CreateCmd.Flags().Float64Var(&vatRate, "vat-rate", 0, "VAT rate percentage (defaults to 20 server-side when omitted)")
	CreateCmd.Flags().StringVar(&vatDischargeMotive, "vat-discharge-motive", "", "VAT discharge motive")
	CreateCmd.Flags().StringVar(&siren, "siren", "", "SIREN number")
	CreateCmd.Flags().StringVar(&siret, "siret", "", "SIRET number")
	CreateCmd.Flags().StringVar(&naf, "naf", "", "NAF code")
	CreateCmd.Flags().StringVar(&mf, "mf", "", "MF number")
	CreateCmd.Flags().StringVar(&identificationNumber, "identification-number", "", "Identification number")
	CreateCmd.Flags().StringVar(&purchaseOrder, "purchase-order", "", "Purchase order")
	CreateCmd.Flags().Float64Var(&hoursPerDay, "hours-per-day", 0, "Hours per day (defaults to 7 server-side when omitted)")
	CreateCmd.Flags().Float64Var(&dailyRate, "daily-rate", 0, "Daily rate")
	CreateCmd.Flags().BoolVar(&sendReportsWithInvoice, "send-reports-with-invoice", false, "Attach time reports to invoice emails sent to this client")
	CreateCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
