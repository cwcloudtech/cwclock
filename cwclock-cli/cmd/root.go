package cmd

import (
	"cwclock/cmd/admin"
	"cwclock/cmd/ai"
	"cwclock/cmd/bootstrap"
	"cwclock/cmd/client"
	"cwclock/cmd/configure"
	"cwclock/cmd/job"
	"cwclock/cmd/metrics"
	"cwclock/cmd/organization"
	"cwclock/cmd/project"
	"cwclock/cmd/record"
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
	"moul.io/banner"
)

var (
	fversion    bool
	cli_version string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cwclock",
	Short: "\nA Command Line interface to manage CWClock",
	Long:  "\nA Command Line interface to manage CWClock.\nComplete documentation is available here: https://www.cwcloud.tech/docs/tutorials/cwclock",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.Inline("cwclock cli"))
		if fversion {
			handlers.HandleVersion(cli_version)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	cli_version = version
	err := rootCmd.Execute()
	utils.ExitIfErrorWithouMsg(err)
}

func init() {
	rootCmd.Flags().BoolVarP(&fversion, "version", "v", false, "The cli version")

	rootCmd.AddCommand(ai.AiCmd)
	rootCmd.AddCommand(configure.ConfigureCmd)
	rootCmd.AddCommand(bootstrap.BootstrapCmd)
	rootCmd.AddCommand(metrics.MetricsCmd)
	rootCmd.AddCommand(record.RecordCmd)
	rootCmd.AddCommand(organization.OrganizationCmd)
	rootCmd.AddCommand(client.ClientCmd)
	rootCmd.AddCommand(project.ProjectCmd)
	rootCmd.AddCommand(job.JobCmd)
	rootCmd.AddCommand(admin.AdminCmd)
}
