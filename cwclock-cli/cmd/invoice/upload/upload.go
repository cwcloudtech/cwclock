package upload

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID string
	id    string
)

var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload/reupload an invoice to external connections",
	Long:  `This command lets you push an already-generated invoice's PDF to every one of its organization's external connections again (e.g. after fixing a connection's credentials).`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleInvoiceUpload(orgID, id)
		utils.ExitIfError(err)
	},
}

func init() {
	UploadCmd.DisableFlagsInUseLine = true
	UploadCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	UploadCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Invoice ID or number to upload/reupload (required)")
}
