package handlers

import (
	"cwclock/config"
	"cwclock/utils"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

func renderCommandOutput(responseBody []byte, formatOverride string, jqQuery string, action string, indexID string) {
	format := config.GetDefaultFormat(formatOverride)
	responseText := strings.TrimSpace(string(responseBody))

	payload := buildCommandPayload(action, indexID, responseText)
	payload = applyJQOnPayload(payload, jqQuery)

	if format == "json" {
		utils.PrintJson(payload)
		return
	}

	utils.PrintPrettyArray("Result", toPrettyLines(payload))
}

func buildCommandPayload(action string, indexID string, responseText string) interface{} {
	if utils.IsBlank(responseText) {
		if utils.IsBlank(action) && utils.IsBlank(indexID) {
			return map[string]string{"status": "success"}
		}

		payload := map[string]string{"status": "success"}
		if utils.IsNotBlank(action) {
			payload["action"] = action
		}
		if utils.IsNotBlank(indexID) {
			payload["index_id"] = indexID
		}
		return payload
	}

	var payload interface{}
	if err := json.Unmarshal([]byte(responseText), &payload); err == nil {
		return payload
	}

	return map[string]string{"output": responseText}
}

func toPrettyLines(payload interface{}) []string {
	if list, ok := payload.([]interface{}); ok {
		lines := make([]string, 0, len(list))
		for _, item := range list {
			encodedItem, err := json.Marshal(item)
			if err != nil {
				lines = append(lines, strings.TrimSpace(fmt.Sprintf("%v", item)))
				continue
			}
			lines = append(lines, utils.JsonInlinePrint(string(encodedItem)))
		}

		if len(lines) > 0 {
			return lines
		}
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return []string{strings.TrimSpace(fmt.Sprintf("%v", payload))}
	}

	return []string{utils.JsonInlinePrint(string(encodedPayload))}
}

func applyJQOnPayload(payload interface{}, jqQuery string) interface{} {
	if utils.IsBlank(jqQuery) {
		return payload
	}

	parsedQuery, err := gojq.Parse(jqQuery)
	if err != nil {
		utils.ExitIfErrorWithMsg("Invalid jq query", err)
	}

	results := make([]interface{}, 0)
	iterator := parsedQuery.Run(payload)
	for {
		output, hasNext := iterator.Next()
		if !hasNext {
			break
		}

		if outputErr, ok := output.(error); ok {
			utils.ExitIfErrorWithMsg("jq execution error", outputErr)
		}

		results = append(results, output)
	}

	return results
}
