package config

import (
	"cwclock/env"
	"cwclock/utils"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
)

func GetValueFromFile(content_file string, key string) string {
	lines := strings.Split(content_file, "\n")
	var requested_line string
	for i, line := range lines {
		if strings.HasPrefix(line, key+" =") {
			requested_line = lines[i]
		}
	}

	if utils.IsBlank(requested_line) {
		return utils.EMPTY
	}

	return strings.TrimSpace(strings.Split(requested_line, " = ")[1])
}

func GetConfigValue(key string, defaultValue string) string {
	trimedDefaultValue := strings.TrimSpace(defaultValue)
	if envValue := os.Getenv("CWCLOCK_" + strings.ToUpper(key)); utils.IsNotBlank(envValue) {
		return strings.TrimSpace(envValue)
	}

	dirname, err := os.UserHomeDir()
	if nil != err {
		return trimedDefaultValue
	}

	config_path := fmt.Sprintf("%s/.cwclock/config", dirname)
	content, err := os.ReadFile(config_path)
	if nil != err {
		return trimedDefaultValue
	}

	file_content := string(content)
	if value := GetValueFromFile(file_content, key); utils.IsNotBlank(value) {
		return value
	}

	return trimedDefaultValue
}

func GetApiKey() string {
	return GetConfigValue("api_key", utils.EMPTY)
}

func GetOpenAIBaseURL() string {
	return GetConfigValue("openai_base_url", "https://api.openai.com/v1")
}

func GetOpenAIAPIKey() string {
	return GetConfigValue("openai_api_key", utils.EMPTY)
}

func GetOpenRouterBaseURL() string {
	return GetConfigValue("openrouter_base_url", "https://openrouter.ai/api/v1")
}

func GetOpenRouterAPIKey() string {
	return GetConfigValue("openrouter_api_key", utils.EMPTY)
}

func GetDeepSeekBaseURL() string {
	return GetConfigValue("deepseek_base_url", "https://api.deepseek.com/v1")
}

func GetDeepSeekAPIKey() string {
	return GetConfigValue("deepseek_api_key", utils.EMPTY)
}

func GetAnthropicBaseURL() string {
	return GetConfigValue("anthropic_base_url", "https://api.anthropic.com/v1")
}

func GetAnthropicAPIKey() string {
	return GetConfigValue("anthropic_api_key", utils.EMPTY)
}

func GetGeminiBaseURL() string {
	return GetConfigValue("gemini_base_url", "https://generativelanguage.googleapis.com/v1beta")
}

func GetGeminiAPIKey() string {
	return GetConfigValue("gemini_api_key", utils.EMPTY)
}

func GetMistralBaseURL() string {
	return GetConfigValue("mistral_base_url", "https://api.mistral.ai/v1")
}

func GetMistralAPIKey() string {
	return GetConfigValue("mistral_api_key", utils.EMPTY)
}

func GetDefaultAiProvider() string {
	return GetConfigValue("default_ai_provider", "openai")
}

func GetDefaultAiModel() string {
	return GetConfigValue("default_ai_model", utils.EMPTY)
}

func GetAgentName() string {
	return GetConfigValue("web_agent_name", "cwclock")
}

func GetGitLabBaseURL() string {
	return GetConfigValue("gitlab_base_url", "https://gitlab.com")
}

func GetGitLabToken() string {
	return GetConfigValue("gitlab_token", utils.EMPTY)
}

func GetGitLabWebhookSecret() string {
	return GetConfigValue("gitlab_webhook_secret", utils.EMPTY)
}

func GetDefaultFormat(override string) string {
	if utils.IsNotBlank(override) {
		return strings.TrimSpace(override)
	}

	return GetConfigValue("format", "pretty")
}

func GetApiURL() string {
	return GetConfigValue("api_url", env.API_URL)
}

func GetOrgID() string {
	return GetConfigValue("org_id", utils.EMPTY)
}

func GetCorsEnabled() bool {
	return utils.IsTrue(GetConfigValue("cors_enabled", "false"))
}

func GetCorsAllowedOrigins() []string {
	var origins []string
	if err := json.Unmarshal([]byte(GetConfigValue("cors_allowed_origins", `["*"]`)), &origins); err != nil {
		return []string{"*"}
	}

	return origins
}

func GetRepoURL() string {
	return GetConfigValue("repo_url", env.REPO_URL)
}

func GetRepoBranch() string {
	return GetConfigValue("repo_branch", env.BRANCH)
}

func GetDefaultHelmDirectory() string {
	return GetConfigValue("default_helm_directory", env.DIRECTORY)
}

func SetValueToKeyInFile(file string, key string, value string) {
	dirname, err := os.UserHomeDir()
	utils.ExitIfError(err)

	file_path := fmt.Sprintf("%s/.cwclock/%s", dirname, file)
	file_output, err := os.ReadFile(file_path)
	utils.ExitIfError(err)

	file_content := string(file_output)
	lines := strings.Split(file_content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, key+" =") {
			lines[i] = fmt.Sprintf("%s = %s", key, value)
		}
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(file_path, []byte(output), fs.FileMode(0644))
	utils.ExitIfError(err)
}

func UpdateFileKeyValue(filename string, key string, value string) {
	dirname, err := os.UserHomeDir()
	utils.ExitIfError(err)

	cwclock_path := fmt.Sprintf("%s/.cwclock", dirname)
	file_path := fmt.Sprintf("%s/%s", cwclock_path, filename)
	config_path := fmt.Sprintf("%s/config", cwclock_path)

	if _, err := os.Stat(cwclock_path); os.IsNotExist(err) {
		err := os.Mkdir(cwclock_path, os.ModePerm)
		if nil != err {
			log.Fatal(err)
		}
		os.Create(file_path)
	} else {
		if _, err := os.Stat(file_path); os.IsNotExist(err) {
			os.Create(config_path)
		}
	}

	file_content, err := os.ReadFile(file_path)
	utils.ExitIfError(err)

	if utils.IsBlank(GetValueFromFile(string(file_content), key)) {
		config_file, err := os.OpenFile(file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		utils.ExitIfError(err)

		_, err = config_file.WriteString(fmt.Sprintf("%s = %s\n", key, value))
		utils.ExitIfError(err)
	} else {
		SetValueToKeyInFile(filename, key, value)
	}
}

func SetDefaultFormat(format string) {
	UpdateFileKeyValue("config", "format", format)
}

func SetApiKey(apiKey string) {
	UpdateFileKeyValue("config", "api_key", apiKey)
}

func SetApiURL(apiURL string) {
	UpdateFileKeyValue("config", "api_url", apiURL)
}

func SetOrgID(orgID string) {
	UpdateFileKeyValue("config", "org_id", orgID)
}

func SetClientID(clientID string) {
	UpdateFileKeyValue("config", "client_id", clientID)
}

func SetProjectID(projectID string) {
	UpdateFileKeyValue("config", "project_id", projectID)
}
