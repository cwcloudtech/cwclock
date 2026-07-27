package stop

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID  string
	text   string
	format string
)

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running timer and send the time record",
	Long:  `This command stops the timer started with "record start" and creates the time record over the elapsed range.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleRecordStop(orgID, text, format)
		utils.ExitIfError(err)
	},
}

func init() {
	StopCmd.DisableFlagsInUseLine = true
	StopCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	StopCmd.Flags().StringVarP(&text, "text", "t", utils.EMPTY, "Time record description (overrides the one given at start)")
	StopCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
