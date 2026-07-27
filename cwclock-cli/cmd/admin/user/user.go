package user

import (
	"cwclock/cmd/admin/user/delete"
	"cwclock/cmd/admin/user/ls"
	"cwclock/cmd/admin/user/setrole"
	"cwclock/cmd/admin/user/update"

	"github.com/spf13/cobra"
)

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `This command lets you list, update, delete and set the role of any user. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	UserCmd.DisableFlagsInUseLine = true
	UserCmd.AddCommand(ls.LsCmd)
	UserCmd.AddCommand(update.UpdateCmd)
	UserCmd.AddCommand(delete.DeleteCmd)
	UserCmd.AddCommand(setrole.SetRoleCmd)
}
