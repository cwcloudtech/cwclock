package get

import (
	"cwclock/client"
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var format string
var valueOnly bool
var filter string

var GetCmd = &cobra.Command{
	Use:   "get <metric_name>",
	Short: "Get all samples for a specific metric",
	Long: `This command lets you retrieve all samples (with labels and values) for a given metric name.
You can use a wildcard suffix for prefix matching, for example "cpu_*".
You can filter samples by label with --filter using either "label:value" or "label=value".
Example: cwclock metrics get cpu_all --filter instance:node-1`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		samples, err := client.GetMetricByName(name, filter)
		utils.ExitIfError(err)
		handlers.HandleGetMetric(name, samples, format, valueOnly)
	},
}

func init() {
	GetCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
	GetCmd.Flags().BoolVar(&valueOnly, "value", false, "Display only raw sample value(s), one per line")
	GetCmd.Flags().StringVar(&filter, "filter", utils.EMPTY, "Filter samples by label, format: label:value or label=value")
}
