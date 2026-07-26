package project

import (
	"cwclock/cmd/project/create"
	"cwclock/cmd/project/delete"
	"cwclock/cmd/project/ls"
	"cwclock/cmd/project/update"

	"github.com/spf13/cobra"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage a client's projects",
	Long:  `This command lets you list, create, update and delete projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	ProjectCmd.DisableFlagsInUseLine = true
	ProjectCmd.AddCommand(ls.LsCmd)
	ProjectCmd.AddCommand(create.CreateCmd)
	ProjectCmd.AddCommand(update.UpdateCmd)
	ProjectCmd.AddCommand(delete.DeleteCmd)
}
