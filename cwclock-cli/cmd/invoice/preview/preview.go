package preview

import (
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	orgID      string
	clientID   string
	projectIDs []string
	begin      string
	end        string
	to         string
	output     string
)

var PreviewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview an invoice without saving it",
	Long:  `This command renders an invoice PDF for a client over a date range without saving anything server-side, so you can check it before generating a real one.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleInvoicePreview(orgID, clientID, projectIDs, begin, end, to, output)
		utils.ExitIfError(err)
	},
}

func init() {
	PreviewCmd.DisableFlagsInUseLine = true
	PreviewCmd.Flags().StringVar(&orgID, "org", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	PreviewCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Client ID or name (required)")
	PreviewCmd.Flags().StringArrayVarP(&projectIDs, "project", "p", nil, "Project ID or name to include (repeatable; empty = every project)")
	PreviewCmd.Flags().StringVar(&begin, "begin", utils.EMPTY, "Begin date/time: ISO-8601 or now()/now()-1h/now()-1d style expression (required)")
	PreviewCmd.Flags().StringVar(&end, "end", utils.EMPTY, "End date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	PreviewCmd.Flags().StringVar(&to, "to", utils.EMPTY, "Alias of --end")
	PreviewCmd.Flags().StringVarP(&output, "output", "o", utils.EMPTY, "Output file path (defaults to the server-provided filename in the current directory)")

	PreviewCmd.Example = fmt.Sprintf(
		"%s\n%s",
		"cwclock invoice preview --client <client uuid> --begin 2024-01-15T09:00:00 --end 2024-01-15T12:00:00",
		"cwclock invoice preview --client <client uuid> --begin 2024-01-15T09:00:00 --to 2024-01-15T12:00:00",
	)
}
