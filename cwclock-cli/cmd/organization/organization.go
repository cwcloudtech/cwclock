package organization

import (
	"cwclock/cmd/organization/connection"
	"cwclock/cmd/organization/create"
	"cwclock/cmd/organization/delete"
	"cwclock/cmd/organization/ls"
	"cwclock/cmd/organization/update"

	"github.com/spf13/cobra"
)

var OrganizationCmd = &cobra.Command{
	Use:   "organization",
	Short: "Manage organizations",
	Long:  `This command lets you list, create, update and delete organizations.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	OrganizationCmd.DisableFlagsInUseLine = true
	OrganizationCmd.AddCommand(ls.LsCmd)
	OrganizationCmd.AddCommand(create.CreateCmd)
	OrganizationCmd.AddCommand(update.UpdateCmd)
	OrganizationCmd.AddCommand(delete.DeleteCmd)
	OrganizationCmd.AddCommand(connection.ConnectionCmd)
}
