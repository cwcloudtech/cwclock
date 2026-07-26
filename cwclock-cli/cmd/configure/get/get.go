package get

import (
	"cwclock/handlers"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var GetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value by key",
	Long:  `This command lets you retrieve any configuration value by its key`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handlers.HandlerGetConfigKey(args[0])
	},
}

func init() {
	GetCmd.DisableFlagsInUseLine = true
}
