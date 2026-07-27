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
	Short: "Delete an invoice",
	Long:  `This command lets you delete an invoice (and its stored PDF) by its ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleInvoiceDelete(orgID, id)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	DeleteCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Invoice ID or number to delete (required)")
}
