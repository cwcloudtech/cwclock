package transfert

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	id     string
	owner  string
	format string
)

var TransfertCmd = &cobra.Command{
	Use:   "transfert",
	Short: "Transfer any organization's ownership",
	Long:  `This command lets you transfer any organization's ownership to another user, regardless of membership. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleAdminOrganizationTransfer(id, owner, format)
		utils.ExitIfError(err)
	},
}

func init() {
	TransfertCmd.DisableFlagsInUseLine = true
	TransfertCmd.Flags().StringVarP(&id, "id", "i", "", "Organization ID (required)")
	TransfertCmd.Flags().StringVar(&owner, "owner", "", "New owner's user ID (required)")
	TransfertCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
