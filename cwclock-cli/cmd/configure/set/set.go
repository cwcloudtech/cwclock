package set

import (
	"cwclock/config"
	"fmt"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var SetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value by key",
	Long:  `This command lets you update any configuration key in the default configuration file`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		keyName := args[0]
		value := args[1]
		config.UpdateFileKeyValue("config", keyName, value)
		fmt.Printf("%s = %v\n", keyName, value)
	},
}

func init() {
	SetCmd.DisableFlagsInUseLine = true
}
