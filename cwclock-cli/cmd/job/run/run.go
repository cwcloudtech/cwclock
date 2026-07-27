package run

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID string
	id    string
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an export job immediately",
	Long:  `This command lets you run an export job now, outside its normal cron schedule, using the same delivery path a scheduled run uses.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleJobRun(orgID, id)
		utils.ExitIfError(err)
	},
}

func init() {
	RunCmd.DisableFlagsInUseLine = true
	RunCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	RunCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Job ID to run (required)")
}
