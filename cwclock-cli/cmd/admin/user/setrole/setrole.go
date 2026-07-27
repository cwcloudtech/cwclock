package setrole

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	id     string
	format string
)

var SetRoleCmd = &cobra.Command{
	Use:   "set-role <role>",
	Short: "Set a user's global role",
	Long:  `This command lets you set a user's global role: superuser, confirmed, disabled or ban. Requires superuser.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleAdminUserSetRole(id, args[0], format)
		utils.ExitIfError(err)
	},
}

func init() {
	SetRoleCmd.DisableFlagsInUseLine = true
	SetRoleCmd.Flags().StringVarP(&id, "id", "i", "", "User ID (required)")
	SetRoleCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
