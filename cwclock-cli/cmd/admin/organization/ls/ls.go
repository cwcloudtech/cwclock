package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var format string

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List every organization",
	Long:  `This command lets you list every organization, regardless of membership. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleAdminOrganizationList(format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
