package pfw

import (
	"cwclock/handlers"

	"github.com/spf13/cobra"
)

var (
	nameSpace  string
	openshift  bool
	configPath string
)

var PfwCmd = &cobra.Command{
	Use:   "pfw",
	Short: "Open tunnels and display the graphical interface",
	Long:  `Open port forwarding for the GUI and API.`,
	Run: func(cmd *cobra.Command, args []string) {
		handlers.HandlePortForward(cmd, nameSpace, openshift, configPath)
	},
}

func init() {
	PfwCmd.Flags().StringVarP(&nameSpace, "namespace", "n", "cwclock", "Namespace (default: cwclock)")
	PfwCmd.Flags().BoolVarP(&openshift, "openshift", "o", false, "Use openshift cli instead of kubectl")
	PfwCmd.Flags().StringVar(&configPath, "config", "", "Path to a YAML port-forward config file")
}
