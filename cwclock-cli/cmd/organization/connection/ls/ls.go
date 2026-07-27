package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	id     string
	format string
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List an organization's external connections",
	Long:  `This command lets you list an organization's external storage connections, with their offset in the array (used by "organization connection --offset" to delete one).`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleOrgConnectionList(id, format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Organization ID (required)")
	LsCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
