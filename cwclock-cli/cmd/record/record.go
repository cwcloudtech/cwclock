package record

import (
	"cwclock/cmd/record/delete"
	"cwclock/cmd/record/ls"
	"cwclock/cmd/record/start"
	"cwclock/cmd/record/stop"
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	begin     string
	end       string
	from      string
	to        string
	text      string
	clientID  string
	projectID string
	orgID     string
	format    string
)

var RecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Track and manage time records",
	Long:  `This command lets you create a time record for a given range, list records and delete a record. Use "record start"/"record stop" to time one instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		effectiveBegin := utils.If(utils.IsNotBlank(begin), begin, from)
		effectiveEnd := utils.If(utils.IsNotBlank(end), end, to)

		if utils.IsBlank(effectiveBegin) && utils.IsBlank(effectiveEnd) {
			cmd.Help()
			return
		}

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
	RecordCmd.DisableFlagsInUseLine = true
	RecordCmd.Flags().StringVar(&begin, "begin", utils.EMPTY, "Begin date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	RecordCmd.Flags().StringVar(&end, "end", utils.EMPTY, "End date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	RecordCmd.Flags().StringVar(&from, "from", utils.EMPTY, "Alias of --begin")
	RecordCmd.Flags().StringVar(&to, "to", utils.EMPTY, "Alias of --end")
	RecordCmd.Flags().StringVarP(&text, "text", "t", utils.EMPTY, "Time record description")
	RecordCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Client ID (optional; inferred from --project when omitted)")
	RecordCmd.Flags().StringVarP(&projectID, "project", "p", utils.EMPTY, "Project ID (required)")
	RecordCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	RecordCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")

	RecordCmd.Example = fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s",
		"cwclock record start -t 'working on the CLI' -p <project-id>",
		"cwclock record stop",
		"cwclock record --begin now()-1h --end now() -t 'catching up' -p <project-id>",
		"cwclock record --from 2024-01-15T09:00:00 --to 2024-01-15T12:00:00 -t 'meeting' -p <project-id>",
		"cwclock record ls --max 10",
		"cwclock record delete -i <record-id>",
	)

	RecordCmd.AddCommand(start.StartCmd)
	RecordCmd.AddCommand(stop.StopCmd)
	RecordCmd.AddCommand(ls.LsCmd)
	RecordCmd.AddCommand(delete.DeleteCmd)
}
