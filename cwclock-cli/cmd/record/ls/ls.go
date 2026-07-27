package ls

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	max    int
	format string
	orgID  string
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List the most recent time records",
	Long:  `This command lets you list your most recent time records (10 by default).`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleRecordList(orgID, max, format)
		utils.ExitIfError(err)
	},
}

func init() {
	LsCmd.DisableFlagsInUseLine = true
	LsCmd.Flags().IntVar(&max, "max", 10, "Maximum number of records to display")
	LsCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
	LsCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
}
