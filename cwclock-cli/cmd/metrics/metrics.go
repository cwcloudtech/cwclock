package metrics

import (
	"cwclock/cmd/metrics/get"
	"cwclock/cmd/metrics/ls"

	"github.com/spf13/cobra"
)

var MetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Get metrics from CWClock",
	Long:  `Get metrics from CWClock in Prometheus format`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	MetricsCmd.DisableFlagsInUseLine = true
	MetricsCmd.AddCommand(ls.ListCmd)
	MetricsCmd.AddCommand(get.GetCmd)
}
