package start

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID     string
	clientID  string
	projectID string
	text      string
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a timer for a new time record",
	Long:  `This command starts a timer; stop it with "record stop" to send the time record. Only --project is required - its client is looked up automatically unless --client overrides it.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleRecordStart(orgID, text, clientID, projectID)
		utils.ExitIfError(err)
	},
}

func init() {
	StartCmd.DisableFlagsInUseLine = true
	StartCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	StartCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Client ID (optional; inferred from --project when omitted)")
	StartCmd.Flags().StringVarP(&projectID, "project", "p", utils.EMPTY, "Project ID (required)")
	StartCmd.Flags().StringVarP(&text, "text", "t", utils.EMPTY, "Time record description (optional; defaults to the project's name)")
}
