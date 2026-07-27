package keys

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	format string
)

var KeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List available configuration keys",
	Long:  `This command lets you list all available configuration keys`,
	Run: func(cmd *cobra.Command, args []string) {
		handlers.HandleGetConfigKeys(format)
	},
}

func init() {
	KeysCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
