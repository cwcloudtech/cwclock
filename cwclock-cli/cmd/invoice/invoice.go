package invoice

import (
	"cwclock/cmd/invoice/delete"
	"cwclock/cmd/invoice/generate"
	"cwclock/cmd/invoice/preview"
	"cwclock/cmd/invoice/send"
	"cwclock/cmd/invoice/upload"

	"github.com/spf13/cobra"
)

var InvoiceCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Preview, generate and manage client invoices",
	Long:  `This command lets you preview and generate an invoice for a client over a date range, and send, delete or upload/reupload an already-generated invoice to external connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	InvoiceCmd.DisableFlagsInUseLine = true
	InvoiceCmd.AddCommand(preview.PreviewCmd)
	InvoiceCmd.AddCommand(generate.GenerateCmd)
	InvoiceCmd.AddCommand(send.SendCmd)
	InvoiceCmd.AddCommand(upload.UploadCmd)
	InvoiceCmd.AddCommand(delete.DeleteCmd)
}
