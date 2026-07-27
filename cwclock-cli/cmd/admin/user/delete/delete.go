package delete

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var id string

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  `This command lets you delete a user account by its ID. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleAdminUserDelete(id)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "User ID to delete (required)")
}
