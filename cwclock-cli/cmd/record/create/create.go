package create

import (
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	orgID     string
	clientID  string
	projectID string
	text      string
	begin     string
	end       string
	from      string
	to        string
	format    string
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a time record for a given range",
	Long:  `This command creates a time record over an explicit begin/end range. Only --project is required - its client is looked up automatically unless --client overrides it, and --text defaults to the project's name when omitted, matching the web app.`,
	Run: func(cmd *cobra.Command, args []string) {
		effectiveBegin := utils.If(utils.IsNotBlank(begin), begin, from)
		effectiveEnd := utils.If(utils.IsNotBlank(end), end, to)

		if utils.IsBlank(effectiveBegin) || utils.IsBlank(effectiveEnd) {
			fmt.Println("Error: both a begin (--begin/--from) and an end (--end/--to) date are required")
			cmd.Help()
			utils.ExitIfNeeded(utils.EMPTY, true)
			return
		}

		err := handlers.HandleRecordCreateRange(orgID, clientID, projectID, text, effectiveBegin, effectiveEnd, format)
		utils.ExitIfError(err)
	},
}

func init() {
	CreateCmd.DisableFlagsInUseLine = true
	CreateCmd.Flags().StringVar(&begin, "begin", utils.EMPTY, "Begin date/time: ISO-8601 or now()/now()-1h/now()-1d style expression (required)")
	CreateCmd.Flags().StringVar(&end, "end", utils.EMPTY, "End date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	CreateCmd.Flags().StringVar(&from, "from", utils.EMPTY, "Alias of --begin")
	CreateCmd.Flags().StringVar(&to, "to", utils.EMPTY, "Alias of --end")
	CreateCmd.Flags().StringVarP(&text, "text", "t", utils.EMPTY, "Time record description (optional; defaults to the project's name)")
	CreateCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Client ID or name (optional; inferred from --project when omitted)")
	CreateCmd.Flags().StringVarP(&projectID, "project", "p", utils.EMPTY, "Project ID or name (required)")
	CreateCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	CreateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")

	CreateCmd.Example = fmt.Sprintf(
		"%s\n%s",
		"cwclock record create --begin now()-1h --end now() -p <project-id>",
		"cwclock record create --from 2024-01-15T09:00:00 --to 2024-01-15T12:00:00 -t 'meeting' -p <project-id>",
	)
}
