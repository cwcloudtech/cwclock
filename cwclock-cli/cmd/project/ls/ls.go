package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID    string
	clientID string
	format   string
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List projects",
	Long:  `This command lets you list an organization's projects, optionally filtered by client.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleProjectList(orgID, clientID, format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&orgID, "org", "o", "", "Organization ID (overrides configured org_id)")
	LsCmd.Flags().StringVarP(&clientID, "client", "c", "", "Filter by client ID")
	LsCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
