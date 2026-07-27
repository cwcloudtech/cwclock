package generate

import (
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	orgID    string
	clientID string
	begin    string
	end      string
	to       string
	output   string
	format   string
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate and save an invoice",
	Long:  `This command renders an invoice PDF for a client over a date range, saves it under its own invoice number, and streams the same PDF back as a download.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleInvoiceGenerate(orgID, clientID, begin, end, to, output, format)
		utils.ExitIfError(err)
	},
}

func init() {
	GenerateCmd.DisableFlagsInUseLine = true
	GenerateCmd.Flags().StringVar(&orgID, "org", utils.EMPTY, "Organization ID (overrides configured org_id)")
	GenerateCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Client ID (required)")
	GenerateCmd.Flags().StringVar(&begin, "begin", utils.EMPTY, "Begin date/time: ISO-8601 or now()/now()-1h/now()-1d style expression (required)")
	GenerateCmd.Flags().StringVar(&end, "end", utils.EMPTY, "End date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	GenerateCmd.Flags().StringVar(&to, "to", utils.EMPTY, "Alias of --end")
	GenerateCmd.Flags().StringVarP(&output, "output", "o", utils.EMPTY, "Output file path (defaults to the server-provided filename in the current directory)")
	GenerateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override for the generated invoice's metadata: pretty|json")

	GenerateCmd.Example = fmt.Sprintf(
		"%s\n%s",
		"cwclock invoice generate --client <client uuid> --begin 2024-01-15T09:00:00 --end 2024-01-15T12:00:00",
		"cwclock invoice generate --client <client uuid> --begin 2024-01-15T09:00:00 --to 2024-01-15T12:00:00",
	)
}
