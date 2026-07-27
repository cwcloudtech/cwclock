package handlers

import (
	"cwclock/config"
	"cwclock/utils"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const MaskedValue = "*******"

var availableConfigKeys = []string{
	"anthropic_api_key",
	"anthropic_base_url",
	"api_key",
	"api_url",
	"deepseek_api_key",
	"deepseek_base_url",
	"default_helm_directory",
	"format",
	"openai_api_key",
	"openai_base_url",
	"openrouter_api_key",
	"openrouter_base_url",
	"gemini_api_key",
	"gemini_base_url",
	"mistral_api_key",
	"mistral_base_url",
	"org_id",
	"repo_branch",
	"repo_url",
}

func HandlerGetConfigKey(key string) {
	value := config.GetConfigValue(key, utils.EMPTY)
	fmt.Printf("%v = %v\n", key, value)
}

func HandleSwitchConfigFile(configFileName *string) {
	availableFiles, err := getFilesInFolder(".cwclock")
	utils.ExitIfError(err)

	found := false
	for _, fileName := range availableFiles {
		if fileName == *configFileName {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Config file '%s' not found\n", *configFileName)
	}

	configFilePath := filepath.Join(getHomeDir(), ".cwclock", *configFileName)
	HandleImportConfigFile(configFilePath)
}

func HandleImportConfigFile(configFilePath string) {
	configContent, err := readConfigFile(configFilePath)
	utils.ExitIfError(err)

	lines := strings.Split(configContent, "\n")

	configMap := make(map[string]string)

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			configMap[key] = value
		}
	}

	apiURL, apiURLExists := configMap["api_url"]
	apiKey, apiKeyExists := configMap["api_key"]
	orgID, orgIDExists := configMap["org_id"]
	clientID, clientIDExists := configMap["client_id"]
	projectID, projectIDExists := configMap["project_id"]
	format, formatExists := configMap["format"]

	if apiURLExists && utils.IsNotBlank(apiURL) {
		config.SetApiURL(apiURL)
		fmt.Printf("API URL = %v\n", apiURL)
	}

	if apiKeyExists && utils.IsNotBlank(apiKey) {
		config.SetApiKey(apiKey)
		fmt.Printf("API key = %v\n", MaskedValue)
	}

	if orgIDExists && utils.IsNotBlank(orgID) {
		config.SetOrgID(orgID)
		fmt.Printf("Organization ID = %v\n", orgID)
	}

	if clientIDExists && utils.IsNotBlank(clientID) {
		config.SetClientID(clientID)
		fmt.Printf("Client ID = %v\n", clientID)
	}

	if projectIDExists && utils.IsNotBlank(projectID) {
		config.SetProjectID(projectID)
		fmt.Printf("Project ID = %v\n", projectID)
	}

	if formatExists && utils.IsNotBlank(format) {
		config.SetDefaultFormat(format)
		fmt.Printf("Default output format = %v\n", format)
	}

	fmt.Println("Config is set successfully")
}

func HandleGetConfigFiles() {
	fileNames, err := getFilesInFolder(".cwclock")
	utils.ExitIfError(err)

	println("available config files:")
	for _, fileName := range fileNames {
		println(fileName)
	}
}

func HandleGetConfigKeys(formatOverride string) {
	keys := make([]string, len(availableConfigKeys))
	copy(keys, availableConfigKeys)
	sort.Strings(keys)

	switch config.GetDefaultFormat(formatOverride) {
	case "json":
		utils.PrintJson(keys)
	default:
		utils.PrintPrettyArray("Available configuration keys", keys)
	}
}

func getFilesInFolder(folderName string) ([]string, error) {
	homeDir := getHomeDir()
	folderPath := filepath.Join(homeDir, folderName)

	file, err := os.Open(folderPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	names, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	return names, nil
}

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	return home
}

func readConfigFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return utils.EMPTY, err
	}

	return string(data), nil
}
