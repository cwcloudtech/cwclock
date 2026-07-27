package delete

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var id string

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete any organization",
	Long:  `This command lets you delete any organization by its ID, regardless of membership. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleOrganizationDelete(id)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&id, "id", "i", "", "Organization ID to delete (required)")
}
