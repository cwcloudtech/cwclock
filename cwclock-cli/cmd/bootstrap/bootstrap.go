package bootstrap

import (
	"cwclock/cmd/bootstrap/pfw"
	"cwclock/cmd/bootstrap/uninstall"
	"cwclock/config"
	"cwclock/env"
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagVerbose  bool
	nameSpace    string
	otherValues  []string
	releaseName  string
	directory    string
	keepDir      bool
	clusterName  string
	recreateNs   bool
	openshift    bool
	tempRepoURL  string
	tempBranch   string
	tempUsername string
	tempPassword string
)

var BootstrapCmd = &cobra.Command{
	Use:   "bootstrap [flags]",
	Short: "CWClock installation on Kubernetes",
	Long:  `CWClock installation on Kubernetes.`,
	Run: func(cmd *cobra.Command, args []string) {
		handlers.HandleBootstrap(cmd, releaseName, nameSpace, clusterName, otherValues, flagVerbose, directory, keepDir, recreateNs, openshift)
	},
}

func init() {
	BootstrapCmd.DisableFlagsInUseLine = true
	BootstrapCmd.Flags().StringVarP(&releaseName, "release", "r", "cwclock", "Release name for deployment (default: cwclock)")
	BootstrapCmd.Flags().StringVarP(&nameSpace, "namespace", "n", "cwclock", "Namespace to use for deployment (default: cwclock)")
	BootstrapCmd.Flags().StringVarP(&clusterName, "kind-cluster", "c", utils.EMPTY, "Kind cluster name (optional)")
	BootstrapCmd.Flags().StringVarP(&directory, "directory", "D", utils.EMPTY, "Helm chart directory (overrides default_helm_directory/env.DIRECTORY and skips git clone)")
	BootstrapCmd.Flags().BoolVarP(&keepDir, "keep-dir", "k", false, "Keep the local helm directory")
	BootstrapCmd.Flags().BoolVarP(&recreateNs, "recreate-ns", "d", false, "Recreate the namespace")
	BootstrapCmd.Flags().BoolVarP(&openshift, "openshift", "o", false, "Use openshift cli instead of kubectl")
	BootstrapCmd.Flags().StringArrayVarP(&otherValues, "value", "p", []string{}, `Values to override other configurations (e.g. --value key=value --value key2=value2)`)
	BootstrapCmd.Flags().StringArrayVar(&otherValues, "values", []string{}, `Alias of --value`)

	configureCmd := &cobra.Command{
		Use:   "configure",
		Short: "Bootstrap CWClock with custom repository configuration",
		Long: `Bootstrap CWClock installation with temporary repository configuration overrides. 
		This command allows you to specify custom repository URL and branch for a single deployment.

		You can either:
		1. Use flags to specify all values directly
		2. Run without flags for an interactive prompt`,
		Example: `  # Interactive mode:
		cwclock bootstrap configure

		# Flag-based mode:
		cwclock bootstrap configure --repo-url=https://custom-repo.git --branch=dev
		cwclock bootstrap configure --repo-url=https://gitlab.alternative.io/helm.git --branch=feature-123
		cwclock bootstrap configure --repo-url=https://custom-repo.git --username=myuser --password=mypassword`,
		Run: func(cmd *cobra.Command, args []string) {
			defaultRepoURL := env.REPO_URL
			defaultBranch := env.BRANCH

			if utils.IsNotBlank(config.GetRepoURL()) {
				defaultRepoURL = config.GetRepoURL()
			}
			if utils.IsNotBlank(config.GetRepoBranch()) {
				defaultBranch = config.GetRepoBranch()
			}

			if !cmd.Flags().Changed("repo-url") && !cmd.Flags().Changed("branch") &&
				!cmd.Flags().Changed("username") && !cmd.Flags().Changed("password") {

				fmt.Printf("Repository URL [%s]: ", defaultRepoURL)
				userRepoURL := utils.PromptUserForValue()
				if utils.IsNotBlank(userRepoURL) {
					tempRepoURL = userRepoURL
				} else {
					tempRepoURL = defaultRepoURL
				}

				fmt.Printf("Branch [%s]: ", defaultBranch)
				userBranch := utils.PromptUserForValue()
				if utils.IsNotBlank(userBranch) {
					tempBranch = userBranch
				} else {
					tempBranch = defaultBranch
				}

				if tempRepoURL != defaultRepoURL {
					fmt.Print("Username (optional): ")
					tempUsername = utils.PromptUserForValue()

					if utils.IsNotBlank(tempUsername) {
						fmt.Print("Password (optional): ")
						tempPassword = utils.PromptUserForValue()
					}
				}

			} else {
				if !cmd.Flags().Changed("repo-url") {
					tempRepoURL = defaultRepoURL
				}
				if !cmd.Flags().Changed("branch") {
					tempBranch = defaultBranch
				}
			}

			tempConfig := &handlers.RepoConfig{
				RepoURL:  tempRepoURL,
				Branch:   tempBranch,
				Username: tempUsername,
				Password: tempPassword,
			}
			handlers.SaveRepoConfig(tempConfig)
			fmt.Printf("Configuration saved successfully:\n")
			fmt.Printf("- Repository URL: %s\n", tempConfig.RepoURL)
			fmt.Printf("- Branch: %s\n", tempConfig.Branch)
			if utils.IsNotBlank(tempConfig.Username) {
				fmt.Printf("- Authentication: Enabled\n")
			}
			fmt.Printf("\nTo perform installation with these settings, run: cwclock bootstrap\n")
		},
	}

	configureCmd.Flags().StringVarP(&releaseName, "release", "r", "cwclock", "Release name for deployment (default: cwclock)")
	configureCmd.Flags().StringVarP(&nameSpace, "namespace", "n", "cwclock", "Namespace to use for deployment (default: cwclock)")
	configureCmd.Flags().StringVarP(&clusterName, "kind-cluster", "c", utils.EMPTY, "Kind cluster name (optional)")
	configureCmd.Flags().StringVarP(&directory, "directory", "D", utils.EMPTY, "Helm chart directory (overrides default_helm_directory/env.DIRECTORY and skips git clone)")
	configureCmd.Flags().BoolVarP(&keepDir, "keep-dir", "k", false, "Keep the local helm directory")
	configureCmd.Flags().StringArrayVarP(&otherValues, "value", "p", []string{}, `Values to override other configurations (e.g. --value key=value --value key2=value2)`)
	configureCmd.Flags().StringArrayVar(&otherValues, "values", []string{}, `Alias of --value`)

	// Configure subcommand flags
	configureCmd.Flags().StringVarP(&tempRepoURL, "repo-url", "u", utils.EMPTY, "Temporary repository URL")
	configureCmd.Flags().StringVarP(&tempBranch, "branch", "b", utils.EMPTY, "Temporary branch name")
	configureCmd.Flags().StringVar(&tempUsername, "username", "U", "Username for repository authentication")
	configureCmd.Flags().StringVar(&tempPassword, "password", "P", "Password for repository authentication")

	BootstrapCmd.AddCommand(configureCmd)
	BootstrapCmd.AddCommand(uninstall.UninstallCmd)
	BootstrapCmd.AddCommand(pfw.PfwCmd)
}
