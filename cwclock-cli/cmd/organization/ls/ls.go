package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var format string

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List organizations",
	Long:  `This command lets you list the organizations you belong to.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleOrganizationList(format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
