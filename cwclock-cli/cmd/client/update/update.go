package update

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID                  string
	id                     string
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

var fieldFlagNames = []string{
	"name", "email", "invoice-emails", "contact-name", "address", "postal-code",
	"city", "country", "vat-number", "vat-rate", "vat-discharge-motive",
	"siren", "siret", "naf", "mf", "identification-number", "purchase-order",
	"hours-per-day", "daily-rate", "send-reports-with-invoice",
}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a client",
	Long:  `This command lets you update an existing client. Only the flags you pass are changed; everything else keeps its current value.`,
	Run: func(cmd *cobra.Command, args []string) {
		changed := make(map[string]bool, len(fieldFlagNames))
		for _, flagName := range fieldFlagNames {
			changed[flagName] = cmd.Flags().Changed(flagName)
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
		err := handlers.HandleOrgClientUpdate(orgID, id, fields, changed, format)
		utils.ExitIfError(err)
	},
}

func init() {
	UpdateCmd.DisableFlagsInUseLine = true
	UpdateCmd.Flags().StringVarP(&orgID, "org", "o", "", "Organization ID (overrides configured org_id)")
	UpdateCmd.Flags().StringVarP(&id, "id", "i", "", "Client ID to update (required)")
	UpdateCmd.Flags().StringVar(&name, "name", "", "Client name")
	UpdateCmd.Flags().StringVar(&country, "country", "", "Country code")
	UpdateCmd.Flags().StringVar(&email, "email", "", "Email")
	UpdateCmd.Flags().StringVar(&invoiceEmails, "invoice-emails", "", "Comma/semicolon-separated invoice recipient emails")
	UpdateCmd.Flags().StringVar(&contactName, "contact-name", "", "Contact name")
	UpdateCmd.Flags().StringVar(&address, "address", "", "Address")
	UpdateCmd.Flags().StringVar(&postalCode, "postal-code", "", "Postal code")
	UpdateCmd.Flags().StringVar(&city, "city", "", "City")
	UpdateCmd.Flags().StringVar(&vatNumber, "vat-number", "", "VAT number")
	UpdateCmd.Flags().Float64Var(&vatRate, "vat-rate", 0, "VAT rate percentage")
	UpdateCmd.Flags().StringVar(&vatDischargeMotive, "vat-discharge-motive", "", "VAT discharge motive")
	UpdateCmd.Flags().StringVar(&siren, "siren", "", "SIREN number")
	UpdateCmd.Flags().StringVar(&siret, "siret", "", "SIRET number")
	UpdateCmd.Flags().StringVar(&naf, "naf", "", "NAF code")
	UpdateCmd.Flags().StringVar(&mf, "mf", "", "MF number")
	UpdateCmd.Flags().StringVar(&identificationNumber, "identification-number", "", "Identification number")
	UpdateCmd.Flags().StringVar(&purchaseOrder, "purchase-order", "", "Purchase order")
	UpdateCmd.Flags().Float64Var(&hoursPerDay, "hours-per-day", 0, "Hours per day")
	UpdateCmd.Flags().Float64Var(&dailyRate, "daily-rate", 0, "Daily rate")
	UpdateCmd.Flags().BoolVar(&sendReportsWithInvoice, "send-reports-with-invoice", false, "Attach time reports to invoice emails sent to this client")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
