package update

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	id                   string
	name                 string
	accountingEmail      string
	address              string
	postalCode           string
	city                 string
	country              string
	vatNumber            string
	siren                string
	siret                string
	naf                  string
	mf                   string
	identificationNumber string
	iban                 string
	bic                  string
	currency             string
	format               string
)

var fieldFlagNames = []string{
	"name", "accounting-email", "address", "postal-code", "city", "country",
	"vat-number", "siren", "siret", "naf", "mf", "identification-number",
	"iban", "bic", "currency",
}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an organization",
	Long:  `This command lets you update an existing organization. Only the flags you pass are changed; everything else keeps its current value.`,
	Run: func(cmd *cobra.Command, args []string) {
		changed := make(map[string]bool, len(fieldFlagNames))
		for _, flagName := range fieldFlagNames {
			changed[flagName] = cmd.Flags().Changed(flagName)
		}

		fields := handlers.OrganizationFields{
			Name:                 name,
			AccountingEmail:      accountingEmail,
			Address:              address,
			PostalCode:           postalCode,
			City:                 city,
			Country:              country,
			VATNumber:            vatNumber,
			SIREN:                siren,
			SIRET:                siret,
			NAF:                  naf,
			MF:                   mf,
			IdentificationNumber: identificationNumber,
			IBAN:                 iban,
			BIC:                  bic,
			Currency:             currency,
		}
		err := handlers.HandleOrganizationUpdate(id, fields, changed, format)
		utils.ExitIfError(err)
	},
}

func init() {
	UpdateCmd.DisableFlagsInUseLine = true
	UpdateCmd.Flags().StringVarP(&id, "id", "i", "", "Organization ID to update (required)")
	UpdateCmd.Flags().StringVar(&name, "name", "", "Organization name")
	UpdateCmd.Flags().StringVar(&country, "country", "", "Country code")
	UpdateCmd.Flags().StringVar(&accountingEmail, "accounting-email", "", "Accounting email, cc'd on invoice emails")
	UpdateCmd.Flags().StringVar(&address, "address", "", "Address")
	UpdateCmd.Flags().StringVar(&postalCode, "postal-code", "", "Postal code")
	UpdateCmd.Flags().StringVar(&city, "city", "", "City")
	UpdateCmd.Flags().StringVar(&vatNumber, "vat-number", "", "VAT number")
	UpdateCmd.Flags().StringVar(&siren, "siren", "", "SIREN number")
	UpdateCmd.Flags().StringVar(&siret, "siret", "", "SIRET number")
	UpdateCmd.Flags().StringVar(&naf, "naf", "", "NAF code")
	UpdateCmd.Flags().StringVar(&mf, "mf", "", "MF number")
	UpdateCmd.Flags().StringVar(&identificationNumber, "identification-number", "", "Identification number")
	UpdateCmd.Flags().StringVar(&iban, "iban", "", "IBAN")
	UpdateCmd.Flags().StringVar(&bic, "bic", "", "BIC")
	UpdateCmd.Flags().StringVar(&currency, "currency", "", "Currency code")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
