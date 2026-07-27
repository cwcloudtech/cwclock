package handlers

import (
	"crypto/rand"
	"cwclock/config"
	"cwclock/env"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"cwclock/utils"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var keyValuePattern = regexp.MustCompile(`^[^=]+=[^=]+$`)
var yamlFilePattern = regexp.MustCompile(`(?i)^.+\.ya?ml$`)

//go:embed assets/pfw.yml
var defaultPortForwardConfig []byte

type RepoConfig struct {
	RepoURL  string
	Branch   string
	Username string
	Password string
}

func GetRepoConfig() RepoConfig {
	repoURL := config.GetRepoURL()
	branch := config.GetRepoBranch()

	if utils.IsBlank(repoURL) {
		repoURL = env.REPO_URL
	} else {
		env.REPO_URL = repoURL
	}

	if utils.IsBlank(branch) {
		branch = env.BRANCH
	} else {
		env.BRANCH = branch
	}

	return RepoConfig{
		RepoURL: repoURL,
		Branch:  branch,
	}
}

func (c *RepoConfig) SetRepoURL(url string) {
	config.UpdateFileKeyValue("config", "repo_url", url)
	env.REPO_URL = url
}

func (c *RepoConfig) SetRepoBranch(branch string) {
	config.UpdateFileKeyValue("config", "repo_branch", branch)
	env.BRANCH = branch
}

func SaveRepoConfig(config *RepoConfig) {
	if utils.IsNotBlank(config.RepoURL) {
		config.SetRepoURL(config.RepoURL)
	}

	if utils.IsNotBlank(config.Branch) {
		config.SetRepoBranch(config.Branch)
	}

	if utils.IsNotBlank(config.Username) {
		env.REPO_USERNAME = config.Username
	}

	if utils.IsNotBlank(config.Password) {
		env.REPO_PASSWORD = config.Password
	}
}

func HandleBootstrap(cmd *cobra.Command, releaseName, nameSpace string, clusterName string, otherValues []string, flagVerbose bool, directoryOverride string, keepDir bool, recreateNs bool, openshift bool) {
	repoConfig := GetRepoConfig()
	repoURL := repoConfig.RepoURL
	directory := config.GetDefaultHelmDirectory()
	branch := repoConfig.Branch
	externalDirectory := cmd.Flags().Changed("directory")

	if utils.IsNotBlank(directoryOverride) {
		directory = strings.TrimSpace(directoryOverride)
	}

	if externalDirectory {
		if _, err := os.Stat(directory); err != nil {
			log.Printf("Directory %q is not accessible: %v", directory, err)
			return
		}
		log.Printf("Using local helm directory %q; skipping repository clone.", directory)
	} else {
		if err := CloneRepo(repoURL, directory, branch, keepDir, env.REPO_USERNAME, env.REPO_PASSWORD); err != nil {
			log.Printf("Error cloning repository: %v", err)
			return
		}
	}

	subdir := fmt.Sprintf("%s/%s", directory, env.SUBDIR)
	password, err := replaceVariablesInValues(subdir, nameSpace)
	if err != nil || utils.IsBlank(password) {
		log.Printf("Error updating values.yaml: %v", err)
		return
	} else {
		log.Printf("Generated random password for tenant: %s", password)
	}

	if err := runKindRecreateCluster(clusterName); err != nil {
		log.Printf("Error recreating kind cluster: %v", err)
		return
	}

	log.Println("Starting Helm chart installation...")

	if err := runHelmDependancyUpdate(subdir, keepDir); err != nil {
		log.Fatalf("Error running helm command: %v", err)
	}

	if err := runDeleteNS(nameSpace, recreateNs, openshift); err != nil {
		log.Printf("Not able to delete the namespace: %s, error: %v", nameSpace, err)
	}

	if err := runCreateNS(nameSpace, openshift); err != nil {
		log.Printf("Not able to create the namespace: %s, error: %v", nameSpace, err)
	}

	if err := runHelmInstall(releaseName, subdir, nameSpace, otherValues, openshift); err != nil {
		log.Fatalf("Error running helm command: %v", err)
	}

	if !externalDirectory {
		if err := runGitClean(directory, repoConfig); err != nil {
			log.Printf("Error running git clean commands: %v", err)
		}
	}

	log.Println("Helm chart installation started in background.")
}

func runDeleteNS(nameSpace string, recreateNs bool, openshift bool) error {
	if !recreateNs {
		return nil
	}

	kubectlCommand := utils.If(openshift, "oc", "kubectl")

	kubectlArgs := []string{
		"delete",
		"ns",
		nameSpace,
	}

	log.Printf("Executing %s command: %s %s", kubectlCommand, kubectlCommand, strings.Join(kubectlArgs, " "))

	kubectlDeleteNs := exec.Command(kubectlCommand, kubectlArgs...)
	kubectlDeleteNs.Stdout = os.Stdout
	kubectlDeleteNs.Stderr = os.Stderr

	return kubectlDeleteNs.Run()
}

func runCreateNS(nameSpace string, openshift bool) error {
	kubectlCommand := utils.If(openshift, "oc", "kubectl")

	kubectlArgs := []string{
		"create",
		"ns",
		nameSpace,
	}

	log.Printf("Executing %s command: %s %s", kubectlCommand, kubectlCommand, strings.Join(kubectlArgs, " "))

	kubectlDeleteNs := exec.Command(kubectlCommand, kubectlArgs...)
	kubectlDeleteNs.Stdout = os.Stdout
	kubectlDeleteNs.Stderr = os.Stderr

	return kubectlDeleteNs.Run()
}

func runKindRecreateCluster(clusterName string) error {
	if utils.IsBlank(clusterName) {
		return nil
	}

	kindCommand := "kind"
	kindArgs := []string{
		"delete",
		"cluster",
		"--name", clusterName,
	}

	log.Printf("Executing kind command: %s %s", kindCommand, strings.Join(kindArgs, " "))

	kindDeleteCluster := exec.Command(kindCommand, kindArgs...)
	kindDeleteCluster.Stdout = os.Stdout
	kindDeleteCluster.Stderr = os.Stderr

	if err := kindDeleteCluster.Run(); err != nil {
		log.Printf("error deleting kind cluster: %v", err)
	}

	kindArgs = []string{
		"create",
		"cluster",
		"--name", clusterName,
	}

	log.Printf("Executing kind command: %s %s", kindCommand, strings.Join(kindArgs, " "))

	kindCreateCluster := exec.Command(kindCommand, kindArgs...)
	kindCreateCluster.Stdout = os.Stdout
	kindCreateCluster.Stderr = os.Stderr

	return kindCreateCluster.Run()
}

func runHelmDependancyUpdate(directory string, keepDir bool) error {
	if _, err := os.Stat(directory + "/charts"); !os.IsNotExist(err) {
		if keepDir {
			return nil
		}
	}

	helmCommand := "helm"
	helmArgs := []string{
		"dependency",
		"update",
	}

	log.Printf("Executing helm command: %s %s", helmCommand, strings.Join(helmArgs, " "))

	helmDepUdpate := exec.Command(helmCommand, helmArgs...)
	helmDepUdpate.Dir = directory
	helmDepUdpate.Stdout = os.Stdout
	helmDepUdpate.Stderr = os.Stderr

	return helmDepUdpate.Run()
}

func runHelmInstall(releaseName string, directory string, nameSpace string, otherValues []string, openshift bool) error {
	helmCommand := "helm"
	helmArgs := utils.If(openshift, []string{
		"install",
		releaseName,
		directory,
		"--namespace", nameSpace,
		"--set", "s3.enabled=false",
	}, []string{
		"install",
		releaseName,
		directory,
		"--namespace", nameSpace,
	})

	for _, value := range otherValues {
		trimmedValue := strings.TrimSpace(value)

		switch {
		case keyValuePattern.MatchString(trimmedValue):
			helmArgs = append(helmArgs, "--set", trimmedValue)
		case yamlFilePattern.MatchString(trimmedValue):
			helmArgs = append(helmArgs, "-f", trimmedValue)
		default:
			return fmt.Errorf("invalid --value %q: expected key=value or .yaml/.yml file", value)
		}
	}

	log.Printf("Executing helm command: %s %s", helmCommand, strings.Join(helmArgs, " "))

	helmInstallation := exec.Command(helmCommand, helmArgs...)
	helmInstallation.Stdout = os.Stdout
	helmInstallation.Stderr = os.Stderr
	helmInstallation.SysProcAttr = helmInstallSysProcAttr()

	if err := helmInstallation.Start(); err != nil {
		return err
	}

	log.Printf("Helm install is running in background with PID %d", helmInstallation.Process.Pid)

	return helmInstallation.Process.Release()
}

func CloneRepo(repoURL, directory, branch string, keepDir bool, username, password string) error {
	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		if keepDir {
			return nil
		}

		fmt.Printf("Deleting existing directory: %s\n", directory)
		if err := os.RemoveAll(directory); err != nil {
			return fmt.Errorf("failed to delete existing directory: %v", err)
		}
	}

	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Progress:      os.Stdout,
	}

	if utils.IsNotBlank(username) || utils.IsNotBlank(password) {
		cloneOptions.Auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	_, err := git.PlainClone(directory, false, cloneOptions)

	if err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	fmt.Println("Repository cloned successfully.")
	return nil
}

func runGitClean(directory string, config RepoConfig) error {
	gitCommand := "git"

	commandArgs := [][]string{{"add", "."}, {"stash"}, {"stash", "clear"}, {"reset", "--hard", "origin/" + config.Branch}}

	for _, args := range commandArgs {
		log.Printf("Executing git command: %s %s", gitCommand, strings.Join(args, " "))
		gitCmd := exec.Command(gitCommand, args...)
		gitCmd.Dir = directory
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr
		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("failed to execute git command %v: %w", args, err)
		}
	}

	return nil
}

func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("password length must be greater than 0")
	}

	byteLen := (length*3 + 3) / 4
	randomBytes := make([]byte, byteLen)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes)[:length], nil
}

func replaceVariablesInValues(chartDirectory string, nameSpace string) (string, error) {
	patterns := []string{
		"values.yaml",
		"values.yml",
		"values-*.yaml",
		"values-*.yml",
		"values.*.yaml",
		"values.*.yml",
	}

	valuesFilePaths := make(map[string]struct{})
	for _, pattern := range patterns {
		matchedPaths, err := filepath.Glob(filepath.Join(chartDirectory, pattern))
		if err != nil {
			return "", fmt.Errorf("invalid values file pattern %q: %w", pattern, err)
		}

		for _, matchedPath := range matchedPaths {
			fileInfo, err := os.Stat(matchedPath)
			if err != nil {
				return "", fmt.Errorf("failed to stat values file %q: %w", matchedPath, err)
			}

			if !fileInfo.Mode().IsRegular() {
				continue
			}

			valuesFilePaths[matchedPath] = struct{}{}
		}
	}

	if len(valuesFilePaths) == 0 {
		return "", fmt.Errorf("no values files found in %q", chartDirectory)
	}

	password, err := generateRandomPassword(20)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	for valuesFilePath := range valuesFilePaths {
		content, err := os.ReadFile(valuesFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read values file %q: %w", valuesFilePath, err)
		}

		replacedContent := strings.ReplaceAll(string(content), "${{tenant-name}}", nameSpace)
		replacedContent = strings.ReplaceAll(replacedContent, "${{namespace}}", nameSpace)
		replacedContent = strings.ReplaceAll(replacedContent, "${{tenant-password}}", password)
		replacedContent = strings.ReplaceAll(replacedContent, "${{password}}", password)
		if replacedContent == string(content) {
			continue
		}

		fileInfo, err := os.Stat(valuesFilePath)
		if err != nil {
			return password, fmt.Errorf("failed to stat values file %q: %w", valuesFilePath, err)
		}

		if err := os.WriteFile(valuesFilePath, []byte(replacedContent), fileInfo.Mode()); err != nil {
			return password, fmt.Errorf("failed to update values file %q: %w", valuesFilePath, err)
		}
	}

	return password, nil
}

func HandleUninstall(cmd *cobra.Command, releaseName string, nameSpace string, force bool, openshift bool) {
	log.Println("Starting Helm chart uninstallation...")

	if err := runHelmUninstall(releaseName, nameSpace); err != nil {
		if !force {
			log.Fatalf("Error running helm uninstall command: %v", err)
		} else {
			log.Printf("Error running helm uninstall command: %v", err)
		}
	}

	if err := runDeleteAll(nameSpace, force, openshift); err != nil {
		log.Printf("Error running kubectl delete all command: %v", err)
	}

	log.Println("Helm chart uninstallation completed successfully.")
}

func runHelmUninstall(releaseName, nameSpace string) error {
	helmCommand := "helm"
	helmArgs := []string{
		"uninstall",
		releaseName,
		"--namespace", nameSpace,
	}

	log.Printf("Executing helm command: %s %s", helmCommand, strings.Join(helmArgs, " "))

	helmUninstallation := exec.Command(helmCommand, helmArgs...)
	helmUninstallation.Stdout = os.Stdout
	helmUninstallation.Stderr = os.Stderr

	return helmUninstallation.Run()
}

func runDeleteAll(nameSpace string, force bool, openshift bool) error {
	if !force {
		return nil
	}

	kubectlCommand := utils.If(openshift, "oc", "kubectl")

	kubectlArgs := []string{
		"-n",
		nameSpace,
		"delete",
		"all",
		"--all",
	}

	log.Printf("Executing %s command: %s %s", kubectlCommand, kubectlCommand, strings.Join(kubectlArgs, " "))

	kubectlDeleteAll := exec.Command(kubectlCommand, kubectlArgs...)
	kubectlDeleteAll.Stdout = os.Stdout
	kubectlDeleteAll.Stderr = os.Stderr

	return kubectlDeleteAll.Run()
}

type kubernetesSecret struct {
	Data map[string]string `json:"data"`
}

type portForwardSecret struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type portForwardConfig struct {
	ServiceName string            `yaml:"name"`
	Port        int               `yaml:"port"`
	TargetPort  int               `yaml:"targetPort"`
	Secret      portForwardSecret `yaml:"secret"`
}

type portForwardConfigFile struct {
	PortForwardConfigs []portForwardConfig `yaml:"pfwConfigs"`
}

func getSecretValue(nameSpace string, openshift bool, secretName string, entryName string) (error, string) {
	kubectlCommand := utils.If(openshift, "oc", "kubectl")
	kubectlArgs := []string{
		"-n",
		nameSpace,
		"get",
		"secrets",
		secretName,
		"-o",
		"json",
	}

	log.Printf("Executing %s command: %s %s", kubectlCommand, kubectlCommand, strings.Join(kubectlArgs, " "))

	output, err := exec.Command(kubectlCommand, kubectlArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error retrieving secret: %s", strings.TrimSpace(string(output))), ""
	}

	var secret kubernetesSecret
	if err := json.Unmarshal(output, &secret); err != nil {
		return fmt.Errorf("error parsing secret JSON: %w", err), ""
	}

	pwdBase64, ok := secret.Data[entryName]
	if !ok {
		return fmt.Errorf("%s not found in secret", entryName), ""
	}

	pwdBytes, err := base64.StdEncoding.DecodeString(pwdBase64)
	if err != nil {
		return fmt.Errorf("error decoding %s: %w", entryName, err), ""
	}

	return nil, string(pwdBytes)
}

func HandlePortForward(cmd *cobra.Command, nameSpace string, openshift bool, configPath string) {
	log.Println("Starting tunnels...")

	portForwardConfigs, err := loadPortForwardConfigs(configPath, nameSpace)
	if err != nil {
		log.Fatalf("Error loading port-forward config: %v", err)
	}

	for _, cfg := range portForwardConfigs {
		if err := runPortForward(nameSpace, cfg.ServiceName, cfg.Port, cfg.TargetPort, openshift); err != nil {
			log.Fatalf("Error running kubectl: %v", err)
		}

		if utils.IsNotBlank(cfg.Secret.Name) && utils.IsNotBlank(cfg.Secret.User) && utils.IsNotBlank(cfg.Secret.Password) {
			uerr, user := getSecretValue(nameSpace, openshift, cfg.Secret.Name, cfg.Secret.User)
			if uerr != nil {
				log.Fatalf("Error retrieving user: %v", uerr)
			}

			perr, password := getSecretValue(nameSpace, openshift, cfg.Secret.Name, cfg.Secret.Password)
			if perr != nil {
				log.Fatalf("Error retrieving password: %v", perr)
			}

			log.Printf("Going to %s: http://localhost:%d (%s / %s)", cfg.ServiceName, cfg.TargetPort, user, password)
		} else {
			log.Printf("Going to %s: http://localhost:%d", cfg.ServiceName, cfg.TargetPort)
		}
	}
}

func loadPortForwardConfigs(configPath string, nameSpace string) ([]portForwardConfig, error) {
	configContent := defaultPortForwardConfig
	configSource := "embedded handlers/assets/pfw.yml"

	if utils.IsNotBlank(configPath) {
		fileContent, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config %q: %w", configPath, err)
		}

		configContent = fileContent
		configSource = configPath
	}

	var wrappedConfig portForwardConfigFile
	if err := yaml.Unmarshal(configContent, &wrappedConfig); err != nil {
		return nil, fmt.Errorf("failed to parse port-forward config %q: %w", configSource, err)
	}

	configs := wrappedConfig.PortForwardConfigs
	if len(configs) == 0 {
		if err := yaml.Unmarshal(configContent, &configs); err != nil {
			return nil, fmt.Errorf("failed to parse port-forward config list %q: %w", configSource, err)
		}
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no port-forward configs found in %q", configSource)
	}

	for index := range configs {
		configs[index].ServiceName = expandPortForwardValue(configs[index].ServiceName, nameSpace)
		configs[index].Secret.Name = expandPortForwardValue(configs[index].Secret.Name, nameSpace)

		if utils.IsBlank(configs[index].ServiceName) {
			return nil, fmt.Errorf("config entry %d has an empty serviceName", index)
		}

		if configs[index].Port <= 0 || configs[index].TargetPort <= 0 {
			return nil, fmt.Errorf("config entry %d must define positive port and targetPort values", index)
		}
	}

	return configs, nil
}

func expandPortForwardValue(value string, nameSpace string) string {
	value = strings.ReplaceAll(value, "{{namespace}}", nameSpace)
	value = strings.ReplaceAll(value, "${namespace}", nameSpace)

	return value
}

func runPortForward(nameSpace string, service string, port int, targetPort int, openshift bool) error {
	kubectlCommand := utils.If(openshift, "oc", "kubectl")

	kubectlArgs := []string{
		"-n",
		nameSpace,
		"port-forward",
		"svc/" + service,
		"" + strconv.Itoa(targetPort) + ":" + strconv.Itoa(port),
	}

	log.Printf("Executing %s command: %s %s", kubectlCommand, kubectlCommand, strings.Join(kubectlArgs, " "))

	kubectlPortForward := exec.Command(kubectlCommand, kubectlArgs...)

	return kubectlPortForward.Start()
}
