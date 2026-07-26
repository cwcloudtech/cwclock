package configure

import (
	"cwclock/cmd/configure/get"
	importConfig "cwclock/cmd/configure/import"
	"cwclock/cmd/configure/keys"
	"cwclock/cmd/configure/ls"
	"cwclock/cmd/configure/set"
	switchConfig "cwclock/cmd/configure/switch"
	"cwclock/config"
	"cwclock/handlers"

	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

// configureCmd represents the configure command
var ConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure CLI defaults such as API URL and output format",
	Long: `This command lets you configure CLI defaults such as API URL and output format.
The configure command takes no arguments it will prompt you for each default value`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			default_api_url := config.GetApiURL()
			fmt.Printf("API URL [%s]: ", default_api_url)
			new_api_url := utils.PromptUserForValue()
			if utils.IsNotBlank(new_api_url) {
				handlers.HandlerSetApiURL(new_api_url)
			}

			defaultApiKey := config.GetApiKey()
			fmt.Printf("API key [%s]: ", utils.If(utils.IsNotBlank(defaultApiKey), handlers.MaskedValue, ""))
			newApiKey := utils.PromptUserForValue()
			if utils.IsNotBlank(newApiKey) {
				config.SetApiKey(newApiKey)
			}

			defaultOrgID := config.GetOrgID()
			fmt.Printf("Organization ID [%s]: ", defaultOrgID)
			newOrgID := utils.PromptUserForValue()
			if utils.IsNotBlank(newOrgID) {
				config.SetOrgID(newOrgID)
			}

			defaultClientID := config.GetClientID()
			fmt.Printf("Client ID [%s]: ", defaultClientID)
			newClientID := utils.PromptUserForValue()
			if utils.IsNotBlank(newClientID) {
				config.SetClientID(newClientID)
			}

			defaultProjectID := config.GetProjectID()
			fmt.Printf("Project ID [%s]: ", defaultProjectID)
			newProjectID := utils.PromptUserForValue()
			if utils.IsNotBlank(newProjectID) {
				config.SetProjectID(newProjectID)
			}

			default_format := config.GetDefaultFormat("")
			fmt.Printf("Default output format [%s]: ", default_format)
			new_default_format := utils.PromptUserForValue()
			if utils.IsNotBlank(new_default_format) {
				handlers.HandlerSetDefaultFormat(new_default_format)
			}
		}
	},
}

func init() {
	ConfigureCmd.DisableFlagsInUseLine = true
	ConfigureCmd.AddCommand(set.SetCmd)
	ConfigureCmd.AddCommand(get.GetCmd)
	ConfigureCmd.AddCommand(keys.KeysCmd)
	ConfigureCmd.AddCommand(ls.LsCmd)
	ConfigureCmd.AddCommand(switchConfig.SwitchConfigCmd)
	ConfigureCmd.AddCommand(importConfig.ImportConfigCmd)
}
