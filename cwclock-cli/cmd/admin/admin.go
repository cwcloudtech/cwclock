package admin

import (
	"cwclock/cmd/admin/organization"
	"cwclock/cmd/admin/user"

	"github.com/spf13/cobra"
)

var AdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Superuser administration",
	Long:  `This command lets a superuser manage every organization and user, regardless of membership.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	AdminCmd.DisableFlagsInUseLine = true
	AdminCmd.AddCommand(organization.OrganizationCmd)
	AdminCmd.AddCommand(user.UserCmd)
}
