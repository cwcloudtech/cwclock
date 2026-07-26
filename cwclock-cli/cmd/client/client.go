package client

import (
	"cwclock/cmd/client/create"
	"cwclock/cmd/client/delete"
	"cwclock/cmd/client/ls"
	"cwclock/cmd/client/update"

	"github.com/spf13/cobra"
)

var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Manage an organization's clients",
	Long:  `This command lets you list, create, update and delete an organization's clients.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	ClientCmd.DisableFlagsInUseLine = true
	ClientCmd.AddCommand(ls.LsCmd)
	ClientCmd.AddCommand(create.CreateCmd)
	ClientCmd.AddCommand(update.UpdateCmd)
	ClientCmd.AddCommand(delete.DeleteCmd)
}
