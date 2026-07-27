package handlers

import (
	"cwclock/client"
	"cwclock/config"
	"cwclock/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

const dayLayout = "2006-01-02"
const timeLayout = "15:04:05"

type timerState struct {
	StartedAt string `json:"startedAt"`
	Text      string `json:"text"`
	ClientID  string `json:"clientId"`
	ProjectID string `json:"projectId"`
}

func timerStatePath() (string, error) {
	homeDir := getHomeDir()
	dir := filepath.Join(homeDir, ".cwclock")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return utils.EMPTY, err
	}
	return filepath.Join(dir, "timer.json"), nil
}

func resolveOrgID(override string) (string, error) {
	value := strings.TrimSpace(override)
	if utils.IsBlank(value) {
		value = config.GetOrgID()
	}
	if utils.IsBlank(value) {
		return utils.EMPTY, fmt.Errorf("organization id is required: set it with 'cwclock configure set org_id <id>' or use --org")
	}

	if looksLikeUUID(value) {
		return value, nil
	}

	org, err := resolveOrganization(value)
	if err != nil {
		return utils.EMPTY, fmt.Errorf("organization %q not found by id or name: set it with 'cwclock configure set org_id <id>' or use --org", value)
	}
	return org.ID, nil
}

// resolveTimeEntryDefaults resolves the mandatory project (accepting its id
// or name, see resolveProject) and, with the same project lookup reused
// when needed, fills in the client (also accepting id or name; inferred
// from the project when blank, since a project belongs to exactly one
// client - see ai-instruct-96) and the entry text (defaults to the
// project's own name, matching the web app's TaskInput.jsx "name ||
// project.name" behavior, see ai-instruct-97) whenever their overrides are
// left blank.
func resolveTimeEntryDefaults(orgID string, clientOverride string, projectOverride string, textOverride string) (clientID string, projectID string, text string, err error) {
	if utils.IsBlank(projectOverride) {
		return utils.EMPTY, utils.EMPTY, utils.EMPTY, fmt.Errorf("project id is required: set it with --project")
	}

	project, err := resolveProject(orgID, projectOverride)
	if err != nil {
		return utils.EMPTY, utils.EMPTY, utils.EMPTY, err
	}
	projectID = project.ID

	clientID = strings.TrimSpace(clientOverride)
	if utils.IsNotBlank(clientID) {
		resolvedClientID, err := resolveClientID(orgID, clientID)
		if err != nil {
			return utils.EMPTY, utils.EMPTY, utils.EMPTY, err
		}
		if err := requireProjectsMatchClients([]string{resolvedClientID}, []client.Project{project}); err != nil {
			return utils.EMPTY, utils.EMPTY, utils.EMPTY, err
		}
		clientID = resolvedClientID
	} else {
		clientID = project.ClientID
	}

	text = strings.TrimSpace(textOverride)
	if utils.IsBlank(text) {
		text = project.Name
	}

	return clientID, projectID, text, nil
}

func HandleRecordStart(orgOverride string, text string, clientOverride string, projectOverride string) error {
	statePath, err := timerStatePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(statePath); err == nil {
		return fmt.Errorf("a timer is already running; stop it first with 'cwclock record stop'")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	clientID, projectID, resolvedText, err := resolveTimeEntryDefaults(orgID, clientOverride, projectOverride, text)
	if err != nil {
		return err
	}

	state := timerState{
		StartedAt: time.Now().Format(time.RFC3339),
		Text:      resolvedText,
		ClientID:  clientID,
		ProjectID: projectID,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Timer started at %s\n", state.StartedAt)
	return nil
}

func HandleRecordStop(orgOverride string, textOverride string, formatOverride string) error {
	statePath, err := timerStatePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("no timer is currently running; start one with 'cwclock record start'")
	}

	var state timerState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to read the running timer state: %w", err)
	}

	startedAt, err := time.ParseInLocation(time.RFC3339, state.StartedAt, time.Local)
	if err != nil {
		return fmt.Errorf("failed to parse the running timer start time: %w", err)
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	text := strings.TrimSpace(textOverride)
	if utils.IsBlank(text) {
		text = strings.TrimSpace(state.Text)
	}
	if utils.IsBlank(text) {
		// Not expected once a record started with HandleRecordStart, which
		// already defaults text to the project's name - this only covers a
		// pre-existing timer.json written before that defaulting existed.
		project, err := resolveProject(orgID, state.ProjectID)
		if err != nil {
			return err
		}
		text = project.Name
	}

	entry, err := createTimeEntryFromRange(orgID, state.ClientID, state.ProjectID, text, startedAt, time.Now())
	if err != nil {
		return err
	}

	if err := os.Remove(statePath); err != nil {
		return err
	}

	renderTimeEntry(entry, formatOverride)
	return nil
}

func HandleRecordCreateRange(orgOverride string, clientOverride string, projectOverride string, text string, beginExpr string, endExpr string, formatOverride string) error {
	begin, err := utils.ParseTimeExpr(beginExpr)
	if err != nil {
		return fmt.Errorf("invalid begin date: %w", err)
	}

	end, err := utils.ParseTimeExpr(endExpr)
	if err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}

	if begin.Format(dayLayout) != end.Format(dayLayout) {
		return fmt.Errorf("begin and end must fall on the same day")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	clientID, projectID, resolvedText, err := resolveTimeEntryDefaults(orgID, clientOverride, projectOverride, text)
	if err != nil {
		return err
	}

	entry, err := createTimeEntryFromRange(orgID, clientID, projectID, resolvedText, begin, end)
	if err != nil {
		return err
	}

	renderTimeEntry(entry, formatOverride)
	return nil
}

func createTimeEntryFromRange(orgID string, clientID string, projectID string, text string, begin time.Time, end time.Time) (client.TimeEntry, error) {
	cli, err := client.NewClient()
	if err != nil {
		return client.TimeEntry{}, err
	}

	start := begin.Format(timeLayout)
	stop := end.Format(timeLayout)

	payload := client.CreateTimeEntryPayload{
		ClientID:  clientID,
		ProjectID: projectID,
		Text:      text,
		Day:       begin.Format(dayLayout),
		Start:     &start,
		End:       &stop,
		AllDay:    false,
	}

	return cli.CreateTimeEntry(orgID, payload)
}

func HandleRecordList(orgOverride string, max int, formatOverride string) error {
	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	if max <= 0 {
		max = 10
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	list, err := cli.ListTimeEntries(orgID, max)
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(list.Items)
		return nil
	}

	displayTimeEntriesAsTable(list.Items)
	return nil
}

func HandleRecordDelete(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("record id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.DeleteTimeEntry(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func renderTimeEntry(entry client.TimeEntry, formatOverride string) {
	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(entry)
		return
	}

	displayTimeEntriesAsTable([]client.TimeEntry{entry})
}

func displayTimeEntriesAsTable(entries []client.TimeEntry) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Day", "Start", "End", "All Day", "Text"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(entries) {
		table.Append([]string{"No records available", utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for _, entry := range entries {
		table.Append([]string{
			entry.ID,
			entry.Day,
			timeValueOrBlank(entry.Start),
			timeValueOrBlank(entry.End),
			strconv.FormatBool(entry.AllDay),
			entry.Text,
		})
	}
	table.Render()
}

func timeValueOrBlank(value *string) string {
	if value == nil {
		return utils.EMPTY
	}
	return *value
}
