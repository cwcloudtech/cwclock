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
	UpdateCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Organization ID to update (required)")
	UpdateCmd.Flags().StringVar(&name, "name", utils.EMPTY, "Organization name")
	UpdateCmd.Flags().StringVar(&country, "country", utils.EMPTY, "Country code")
	UpdateCmd.Flags().StringVar(&accountingEmail, "accounting-email", utils.EMPTY, "Accounting email, cc'd on invoice emails")
	UpdateCmd.Flags().StringVar(&address, "address", utils.EMPTY, "Address")
	UpdateCmd.Flags().StringVar(&postalCode, "postal-code", utils.EMPTY, "Postal code")
	UpdateCmd.Flags().StringVar(&city, "city", utils.EMPTY, "City")
	UpdateCmd.Flags().StringVar(&vatNumber, "vat-number", utils.EMPTY, "VAT number")
	UpdateCmd.Flags().StringVar(&siren, "siren", utils.EMPTY, "SIREN number")
	UpdateCmd.Flags().StringVar(&siret, "siret", utils.EMPTY, "SIRET number")
	UpdateCmd.Flags().StringVar(&naf, "naf", utils.EMPTY, "NAF code")
	UpdateCmd.Flags().StringVar(&mf, "mf", utils.EMPTY, "MF number")
	UpdateCmd.Flags().StringVar(&identificationNumber, "identification-number", utils.EMPTY, "Identification number")
	UpdateCmd.Flags().StringVar(&iban, "iban", utils.EMPTY, "IBAN")
	UpdateCmd.Flags().StringVar(&bic, "bic", utils.EMPTY, "BIC")
	UpdateCmd.Flags().StringVar(&currency, "currency", utils.EMPTY, "Currency code")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
