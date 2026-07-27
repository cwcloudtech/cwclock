package create

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
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

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an organization",
	Long:  `This command lets you create a new organization.`,
	Run: func(cmd *cobra.Command, args []string) {
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
		err := handlers.HandleOrganizationCreate(fields, format)
		utils.ExitIfError(err)
	},
}

func init() {
	CreateCmd.DisableFlagsInUseLine = true
	CreateCmd.Flags().StringVar(&name, "name", "", "Organization name (required)")
	CreateCmd.Flags().StringVar(&country, "country", "", "Country code (required)")
	CreateCmd.Flags().StringVar(&accountingEmail, "accounting-email", "", "Accounting email, cc'd on invoice emails")
	CreateCmd.Flags().StringVar(&address, "address", "", "Address")
	CreateCmd.Flags().StringVar(&postalCode, "postal-code", "", "Postal code")
	CreateCmd.Flags().StringVar(&city, "city", "", "City")
	CreateCmd.Flags().StringVar(&vatNumber, "vat-number", "", "VAT number")
	CreateCmd.Flags().StringVar(&siren, "siren", "", "SIREN number")
	CreateCmd.Flags().StringVar(&siret, "siret", "", "SIRET number")
	CreateCmd.Flags().StringVar(&naf, "naf", "", "NAF code")
	CreateCmd.Flags().StringVar(&mf, "mf", "", "MF number")
	CreateCmd.Flags().StringVar(&identificationNumber, "identification-number", "", "Identification number")
	CreateCmd.Flags().StringVar(&iban, "iban", "", "IBAN")
	CreateCmd.Flags().StringVar(&bic, "bic", "", "BIC")
	CreateCmd.Flags().StringVar(&currency, "currency", "", "Currency code")
	CreateCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
