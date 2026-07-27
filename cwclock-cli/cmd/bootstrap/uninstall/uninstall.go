package uninstall

import (
	"cwclock/handlers"

	"github.com/spf13/cobra"
)

var (
	releaseName string
	nameSpace   string
	force       bool
	openshift   bool
)

var UninstallCmd = &cobra.Command{
	Use:   "uninstall [flags]",
	Short: "Uninstall the Helm release for cwclock application",
	Long:  `Uninstall the Helm release from Kubernetes.`,
	Run: func(cmd *cobra.Command, args []string) {
		handlers.HandleUninstall(cmd, releaseName, nameSpace, force, openshift)
	},
}

func init() {
	UninstallCmd.Flags().StringVarP(&nameSpace, "namespace", "n", "cwclock", "Namespace to use for uninstalling deployment (default: cwclock)")
	UninstallCmd.Flags().StringVarP(&releaseName, "release", "r", "cwclock", "Release name for deployment (default: cwclock)")
	UninstallCmd.Flags().BoolVarP(&force, "force", "f", false, "Force remove every resources on the namespace")
	UninstallCmd.Flags().BoolVarP(&openshift, "openshift", "o", false, "Use openshift cli instead of kubectl")
}
