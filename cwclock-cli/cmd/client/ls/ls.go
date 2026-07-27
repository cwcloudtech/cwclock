package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID  string
	format string
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List an organization's clients",
	Long:  `This command lets you list the clients of an organization.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleOrgClientList(orgID, format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	LsCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
