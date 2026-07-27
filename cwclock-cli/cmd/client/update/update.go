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
	UpdateCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	UpdateCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Client ID or name to update (required)")
	UpdateCmd.Flags().StringVar(&name, "name", utils.EMPTY, "Client name")
	UpdateCmd.Flags().StringVar(&country, "country", utils.EMPTY, "Country code")
	UpdateCmd.Flags().StringVar(&email, "email", utils.EMPTY, "Email")
	UpdateCmd.Flags().StringVar(&invoiceEmails, "invoice-emails", utils.EMPTY, "Comma/semicolon-separated invoice recipient emails")
	UpdateCmd.Flags().StringVar(&contactName, "contact-name", utils.EMPTY, "Contact name")
	UpdateCmd.Flags().StringVar(&address, "address", utils.EMPTY, "Address")
	UpdateCmd.Flags().StringVar(&postalCode, "postal-code", utils.EMPTY, "Postal code")
	UpdateCmd.Flags().StringVar(&city, "city", utils.EMPTY, "City")
	UpdateCmd.Flags().StringVar(&vatNumber, "vat-number", utils.EMPTY, "VAT number")
	UpdateCmd.Flags().Float64Var(&vatRate, "vat-rate", 0, "VAT rate percentage")
	UpdateCmd.Flags().StringVar(&vatDischargeMotive, "vat-discharge-motive", utils.EMPTY, "VAT discharge motive")
	UpdateCmd.Flags().StringVar(&siren, "siren", utils.EMPTY, "SIREN number")
	UpdateCmd.Flags().StringVar(&siret, "siret", utils.EMPTY, "SIRET number")
	UpdateCmd.Flags().StringVar(&naf, "naf", utils.EMPTY, "NAF code")
	UpdateCmd.Flags().StringVar(&mf, "mf", utils.EMPTY, "MF number")
	UpdateCmd.Flags().StringVar(&identificationNumber, "identification-number", utils.EMPTY, "Identification number")
	UpdateCmd.Flags().StringVar(&purchaseOrder, "purchase-order", utils.EMPTY, "Purchase order")
	UpdateCmd.Flags().Float64Var(&hoursPerDay, "hours-per-day", 0, "Hours per day")
	UpdateCmd.Flags().Float64Var(&dailyRate, "daily-rate", 0, "Daily rate")
	UpdateCmd.Flags().BoolVar(&sendReportsWithInvoice, "send-reports-with-invoice", false, "Attach time reports to invoice emails sent to this client")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
