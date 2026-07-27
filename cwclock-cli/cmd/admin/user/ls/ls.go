package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var format string

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List every user",
	Long:  `This command lets you list every registered user. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleAdminUserList(format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
