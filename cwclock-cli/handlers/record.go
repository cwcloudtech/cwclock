package handlers

import (
	"cwclock/client"
	"cwclock/config"
	"cwclock/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

const dayLayout = "2006-01-02"
const timeLayout = "15:04:05"

var relativeTimePattern = regexp.MustCompile(`^now\(\)((?:[+-]\d+[smhdw])*)$`)
var relativeTimeSegmentPattern = regexp.MustCompile(`([+-])(\d+)([smhdw])`)

var absoluteTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04",
	dayLayout,
}

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
		return "", err
	}
	return filepath.Join(dir, "timer.json"), nil
}

func resolveOrgID(override string) (string, error) {
	orgID := strings.TrimSpace(override)
	if utils.IsBlank(orgID) {
		orgID = config.GetOrgID()
	}
	if utils.IsBlank(orgID) {
		return "", fmt.Errorf("organization id is required: set it with 'cwclock configure set org_id <id>' or use --org")
	}
	return orgID, nil
}

func resolveClientID(override string) (string, error) {
	clientID := strings.TrimSpace(override)
	if utils.IsBlank(clientID) {
		return "", fmt.Errorf("client id is required: set it with 'cwclock configure set client_id <id>' or use --client")
	}
	return clientID, nil
}

func resolveProjectID(override string) (string, error) {
	projectID := strings.TrimSpace(override)
	if utils.IsBlank(projectID) {
		return "", fmt.Errorf("project id is required: set it with 'cwclock configure set project_id <id>' or use --project")
	}
	return projectID, nil
}

// ParseTimeExpr resolves a date/time expression to an absolute time. It
// supports plain ISO-8601 strings as well as Grafana-style relative
// expressions built from now() and a chain of +/-<amount><unit> segments
// (units: s, m, h, d, w), e.g. now(), now()-1h, now()-1d.
func ParseTimeExpr(expr string) (time.Time, error) {
	trimmed := strings.TrimSpace(expr)
	if utils.IsBlank(trimmed) {
		return time.Time{}, fmt.Errorf("date is required")
	}

	if matches := relativeTimePattern.FindStringSubmatch(trimmed); matches != nil {
		t := time.Now()
		for _, segment := range relativeTimeSegmentPattern.FindAllStringSubmatch(matches[1], -1) {
			amount, err := strconv.Atoi(segment[2])
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid date expression %q", expr)
			}
			if segment[1] == "-" {
				amount = -amount
			}
			switch segment[3] {
			case "s":
				t = t.Add(time.Duration(amount) * time.Second)
			case "m":
				t = t.Add(time.Duration(amount) * time.Minute)
			case "h":
				t = t.Add(time.Duration(amount) * time.Hour)
			case "d":
				t = t.AddDate(0, 0, amount)
			case "w":
				t = t.AddDate(0, 0, amount*7)
			}
		}
		return t, nil
	}

	for _, layout := range absoluteTimeLayouts {
		if t, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q: expected an ISO-8601 date/time or a now()/now()-1h/now()-1d style expression", expr)
}

func HandleRecordStart(text string, clientOverride string, projectOverride string) error {
	statePath, err := timerStatePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(statePath); err == nil {
		return fmt.Errorf("a timer is already running; stop it first with --stop")
	}

	clientID, err := resolveClientID(clientOverride)
	if err != nil {
		return err
	}

	projectID, err := resolveProjectID(projectOverride)
	if err != nil {
		return err
	}

	state := timerState{
		StartedAt: time.Now().Format(time.RFC3339),
		Text:      strings.TrimSpace(text),
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
		return fmt.Errorf("no timer is currently running; start one with --start")
	}

	var state timerState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to read the running timer state: %w", err)
	}

	startedAt, err := time.ParseInLocation(time.RFC3339, state.StartedAt, time.Local)
	if err != nil {
		return fmt.Errorf("failed to parse the running timer start time: %w", err)
	}

	text := state.Text
	if utils.IsNotBlank(textOverride) {
		text = strings.TrimSpace(textOverride)
	}
	if utils.IsBlank(text) {
		return fmt.Errorf("text is required: pass --text when starting or stopping the timer")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
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
	if utils.IsBlank(text) {
		return fmt.Errorf("text is required: use --text")
	}

	begin, err := ParseTimeExpr(beginExpr)
	if err != nil {
		return fmt.Errorf("invalid begin date: %w", err)
	}

	end, err := ParseTimeExpr(endExpr)
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

	clientID, err := resolveClientID(clientOverride)
	if err != nil {
		return err
	}

	projectID, err := resolveProjectID(projectOverride)
	if err != nil {
		return err
	}

	entry, err := createTimeEntryFromRange(orgID, clientID, projectID, strings.TrimSpace(text), begin, end)
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
		table.Append([]string{"No records available", "", "", "", "", ""})
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
		return ""
	}
	return *value
}
