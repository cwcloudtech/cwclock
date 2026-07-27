package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID  string
	format string
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List export jobs",
	Long:  `This command lets you list an organization's scheduled export jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleJobList(orgID, format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID or name (overrides configured org_id)")
	LsCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
