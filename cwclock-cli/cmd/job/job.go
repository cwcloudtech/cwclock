package job

import (
	"cwclock/cmd/job/create"
	"cwclock/cmd/job/delete"
	"cwclock/cmd/job/ls"
	"cwclock/cmd/job/run"
	"cwclock/cmd/job/target"
	"cwclock/cmd/job/update"

	"github.com/spf13/cobra"
)

var JobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage scheduled export jobs",
	Long:  `This command lets you list, create, update and delete an organization's scheduled export jobs, and manage their targets.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	JobCmd.DisableFlagsInUseLine = true
	JobCmd.AddCommand(ls.LsCmd)
	JobCmd.AddCommand(create.CreateCmd)
	JobCmd.AddCommand(update.UpdateCmd)
	JobCmd.AddCommand(delete.DeleteCmd)
	JobCmd.AddCommand(run.RunCmd)
	JobCmd.AddCommand(target.TargetCmd)
}
