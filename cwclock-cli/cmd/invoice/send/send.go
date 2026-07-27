package send

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID string
	id    string
)

var SendCmd = &cobra.Command{
	Use:   "send",
	Short: "Email an invoice to its client",
	Long:  `This command lets you email an already-generated invoice's PDF to its client's invoice recipients.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleInvoiceSend(orgID, id)
		utils.ExitIfError(err)
	},
}

func init() {
	SendCmd.DisableFlagsInUseLine = true
	SendCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	SendCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Invoice ID to send (required)")
}
