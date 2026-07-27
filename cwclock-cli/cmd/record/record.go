package record

import (
	"cwclock/cmd/record/create"
	"cwclock/cmd/record/delete"
	"cwclock/cmd/record/ls"
	"cwclock/cmd/record/start"
	"cwclock/cmd/record/stop"

	"github.com/spf13/cobra"
)

var RecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Track and manage time records",
	Long:  `This command lets you start/stop a timer, create a time record for a given range, list records and delete a record.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RecordCmd.DisableFlagsInUseLine = true
	RecordCmd.AddCommand(start.StartCmd)
	RecordCmd.AddCommand(stop.StopCmd)
	RecordCmd.AddCommand(create.CreateCmd)
	RecordCmd.AddCommand(ls.LsCmd)
	RecordCmd.AddCommand(delete.DeleteCmd)
}
