package switchConfig

import (
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configFileName string
)

var SwitchConfigCmd = &cobra.Command{
	Use:   "switch [config file name]",
	Short: "Switch the config file",
	Long:  `This command lets you switch between different config files`,
	Run: func(cmd *cobra.Command, args []string) {
		errorMessage := "Please provide a valid config file path"
		if len(args) == 0 {
			fmt.Println(errorMessage)
			return
		}
		configFileName = args[0]
		if utils.IsBlank(configFileName) {
			fmt.Println(errorMessage)
		}
		handlers.HandleSwitchConfigFile(&configFileName)
	},
}
