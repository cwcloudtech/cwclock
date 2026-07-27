package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID  string
	id     string
	format string
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List an export job's targets",
	Long:  `This command lets you list the targets of an export job, with their offset in the array (used by "job target delete").`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleJobTargetList(orgID, id, format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	LsCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Job ID (required)")
	LsCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
