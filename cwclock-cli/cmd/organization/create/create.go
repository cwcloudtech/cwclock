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
	CreateCmd.Flags().StringVar(&name, "name", utils.EMPTY, "Organization name (required)")
	CreateCmd.Flags().StringVar(&country, "country", utils.EMPTY, "Country code (required)")
	CreateCmd.Flags().StringVar(&accountingEmail, "accounting-email", utils.EMPTY, "Accounting email, cc'd on invoice emails")
	CreateCmd.Flags().StringVar(&address, "address", utils.EMPTY, "Address")
	CreateCmd.Flags().StringVar(&postalCode, "postal-code", utils.EMPTY, "Postal code")
	CreateCmd.Flags().StringVar(&city, "city", utils.EMPTY, "City")
	CreateCmd.Flags().StringVar(&vatNumber, "vat-number", utils.EMPTY, "VAT number")
	CreateCmd.Flags().StringVar(&siren, "siren", utils.EMPTY, "SIREN number")
	CreateCmd.Flags().StringVar(&siret, "siret", utils.EMPTY, "SIRET number")
	CreateCmd.Flags().StringVar(&naf, "naf", utils.EMPTY, "NAF code")
	CreateCmd.Flags().StringVar(&mf, "mf", utils.EMPTY, "MF number")
	CreateCmd.Flags().StringVar(&identificationNumber, "identification-number", utils.EMPTY, "Identification number")
	CreateCmd.Flags().StringVar(&iban, "iban", utils.EMPTY, "IBAN")
	CreateCmd.Flags().StringVar(&bic, "bic", utils.EMPTY, "BIC")
	CreateCmd.Flags().StringVar(&currency, "currency", utils.EMPTY, "Currency code")
	CreateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
