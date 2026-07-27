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
	Short: "Delete a project",
	Long:  `This command lets you delete a project by its ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleProjectDelete(orgID, id)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	DeleteCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Project ID or name to delete (required)")
}
