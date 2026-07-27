package target

import (
	"cwclock/cmd/job/target/create"
	"cwclock/cmd/job/target/delete"
	"cwclock/cmd/job/target/ls"

	"github.com/spf13/cobra"
)

var TargetCmd = &cobra.Command{
	Use:   "target",
	Short: "Manage an export job's targets",
	Long:  `This command lets you list, create and delete an export job's targets.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	TargetCmd.DisableFlagsInUseLine = true
	TargetCmd.AddCommand(ls.LsCmd)
	TargetCmd.AddCommand(create.CreateCmd)
	TargetCmd.AddCommand(delete.DeleteCmd)
}
