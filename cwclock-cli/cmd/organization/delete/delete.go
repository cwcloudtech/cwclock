package delete

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var id string

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an organization",
	Long:  `This command lets you delete an organization by its ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleOrganizationDelete(id)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Organization ID or name to delete (required)")
}
