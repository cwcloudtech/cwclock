package importConfig

import (
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configFilePath string
)

var ImportConfigCmd = &cobra.Command{
	Use:   "import [config file path]",
	Short: "Import the config file from the given path",
	Long:  `This command lets you import the config file from the given path`,
	Run: func(cmd *cobra.Command, args []string) {
		errorMessage := "Please provide a valid config file path"
		if len(args) == 0 {
			fmt.Println(errorMessage)
			return
		}
		configFilePath = args[0]
		if utils.IsBlank(configFilePath) {
			fmt.Println(errorMessage)
		}
		handlers.HandleImportConfigFile(configFilePath)
	},
}
