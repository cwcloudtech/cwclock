package delete

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	id    string
	orgID string
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a time record",
	Long:  `This command lets you delete a time record by its ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleRecordDelete(orgID, id)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Time record ID to delete")
	DeleteCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
}
