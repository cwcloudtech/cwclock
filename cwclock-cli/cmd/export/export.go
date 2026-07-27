package export

import (
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	orgID      string
	clientIDs  []string
	projectIDs []string
	begin      string
	end        string
	to         string
	reportType string
	fileFormat string
	output     string
)

var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a time report",
	Long:  `This command lets you download a summary or detailed time report as a PDF or CSV file over a date range, optionally filtered to one or more clients/projects (everything is included when omitted).`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleExport(orgID, clientIDs, projectIDs, begin, end, to, reportType, fileFormat, output)
		utils.ExitIfError(err)
	},
}

func init() {
	ExportCmd.DisableFlagsInUseLine = true
	ExportCmd.Flags().StringVar(&orgID, "org", utils.EMPTY, "Organization ID (overrides configured org_id)")
	ExportCmd.Flags().StringArrayVarP(&clientIDs, "client", "c", nil, "Client ID filter (repeatable; every client is included when omitted)")
	ExportCmd.Flags().StringArrayVarP(&projectIDs, "project", "p", nil, "Project ID filter (repeatable; every project is included when omitted)")
	ExportCmd.Flags().StringVar(&begin, "begin", utils.EMPTY, "Begin date/time: ISO-8601 or now()/now()-1h/now()-1d style expression (required)")
	ExportCmd.Flags().StringVar(&end, "end", utils.EMPTY, "End date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	ExportCmd.Flags().StringVar(&to, "to", utils.EMPTY, "Alias of --end")
	ExportCmd.Flags().StringVar(&reportType, "type", "summary", "Report type: summary or detailed")
	ExportCmd.Flags().StringVar(&fileFormat, "file-format", "pdf", "File format: pdf or csv")
	ExportCmd.Flags().StringVarP(&output, "output", "o", utils.EMPTY, "Output file path (defaults to the server-provided filename, or <type>.<file-format> in the current directory)")

	ExportCmd.Example = fmt.Sprintf(
		"%s\n%s\n%s",
		"cwclock export --client <client uuid> --begin 2024-01-15T09:00:00 --end 2024-01-15T12:00:00",
		"cwclock export --client <client uuid> --begin 2024-01-15T09:00:00 --to 2024-01-15T12:00:00",
		"cwclock export --type detailed --file-format csv --client <id1> --client <id2> --project <id3> --begin now()-30d --end now()",
	)
}
