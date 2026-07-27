package delete

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID string
	id    string
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an export job",
	Long:  `This command lets you delete an export job by its ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleJobDelete(orgID, id)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	DeleteCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Job ID to delete (required)")
}
