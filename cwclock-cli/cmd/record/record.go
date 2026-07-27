package record

import (
	"cwclock/cmd/record/delete"
	"cwclock/cmd/record/ls"
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	start     bool
	stop      bool
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
	Long:  `This command lets you start/stop a timer, create a time record for a given range, list records and delete a record.`,
	Run: func(cmd *cobra.Command, args []string) {
		effectiveBegin := utils.If(utils.IsNotBlank(begin), begin, from)
		effectiveEnd := utils.If(utils.IsNotBlank(end), end, to)
		hasRange := utils.IsNotBlank(effectiveBegin) || utils.IsNotBlank(effectiveEnd)

		actionCount := 0
		if start {
			actionCount++
		}
		if stop {
			actionCount++
		}
		if hasRange {
			actionCount++
		}

		if actionCount == 0 {
			cmd.Help()
			return
		}

		if actionCount > 1 {
			fmt.Println("Error: specify only one of --start, --stop or --begin/--end (--from/--to)")
			cmd.Help()
			utils.ExitIfNeeded(utils.EMPTY, true)
			return
		}

		var err error
		switch {
		case start:
			err = handlers.HandleRecordStart(text, clientID, projectID)
		case stop:
			err = handlers.HandleRecordStop(orgID, text, format)
		case hasRange:
			if utils.IsBlank(effectiveBegin) || utils.IsBlank(effectiveEnd) {
				fmt.Println("Error: both a begin (--begin/--from) and an end (--end/--to) date are required")
				cmd.Help()
				utils.ExitIfNeeded(utils.EMPTY, true)
				return
			}
			err = handlers.HandleRecordCreateRange(orgID, clientID, projectID, text, effectiveBegin, effectiveEnd, format)
		}

		utils.ExitIfError(err)
	},
}

func init() {
	RecordCmd.DisableFlagsInUseLine = true
	RecordCmd.Flags().BoolVar(&start, "start", false, "Start a timer for a new time record")
	RecordCmd.Flags().BoolVar(&stop, "stop", false, "Stop the running timer and send the time record")
	RecordCmd.Flags().StringVar(&begin, "begin", utils.EMPTY, "Begin date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	RecordCmd.Flags().StringVar(&end, "end", utils.EMPTY, "End date/time: ISO-8601 or now()/now()-1h/now()-1d style expression")
	RecordCmd.Flags().StringVar(&from, "from", utils.EMPTY, "Alias of --begin")
	RecordCmd.Flags().StringVar(&to, "to", utils.EMPTY, "Alias of --end")
	RecordCmd.Flags().StringVarP(&text, "text", "t", utils.EMPTY, "Time record description")
	RecordCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Client ID (overrides configured client_id)")
	RecordCmd.Flags().StringVarP(&projectID, "project", "p", utils.EMPTY, "Project ID (overrides configured project_id)")
	RecordCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	RecordCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")

	RecordCmd.Example = fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s",
		"cwclock record --start -t 'working on the CLI' -c <client-id> -p <project-id>",
		"cwclock record --stop",
		"cwclock record --begin now()-1h --end now() -t 'catching up' -c <client-id> -p <project-id>",
		"cwclock record --from 2024-01-15T09:00:00 --to 2024-01-15T12:00:00 -t 'meeting' -c <client-id> -p <project-id>",
		"cwclock record ls --max 10",
		"cwclock record delete -i <record-id>",
	)

	RecordCmd.AddCommand(ls.LsCmd)
	RecordCmd.AddCommand(delete.DeleteCmd)
}
