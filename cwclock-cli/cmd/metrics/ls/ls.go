package ls

import (
	"cwclock/client"
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var format string

var ListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all metrics",
	Long: `This command lets you list all available metrics with their labels and values.
This command takes no arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		samples, err := client.GetMetrics()
		utils.ExitIfError(err)
		handlers.HandleListMetrics(samples, format)
	},
}

func init() {
	ListCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
