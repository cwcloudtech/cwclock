package organization

import (
	"cwclock/cmd/admin/organization/delete"
	"cwclock/cmd/admin/organization/ls"
	"cwclock/cmd/admin/organization/transfert"

	"github.com/spf13/cobra"
)

var OrganizationCmd = &cobra.Command{
	Use:   "organization",
	Short: "Manage any organization",
	Long:  `This command lets you list, delete and transfer the ownership of any organization, regardless of membership. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	OrganizationCmd.DisableFlagsInUseLine = true
	OrganizationCmd.AddCommand(ls.LsCmd)
	OrganizationCmd.AddCommand(delete.DeleteCmd)
	OrganizationCmd.AddCommand(transfert.TransfertCmd)
}
