package handlers

import (
	"cwclock/client"
	"cwclock/config"
	"cwclock/utils"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type ProjectFields struct {
	Name         string
	Color        string
	DailyRate    float64
	Subdivisions string
}

func parseCommaList(raw string) []string {
	if utils.IsBlank(raw) {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if utils.IsNotBlank(trimmed) {
			result = append(result, trimmed)
		}
	}
	return result
}

func HandleProjectList(orgOverride string, clientOverride string, formatOverride string) error {
	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	projects, err := cli.ListProjects(orgID, strings.TrimSpace(clientOverride))
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(projects)
		return nil
	}

	displayProjectsAsTable(projects)
	return nil
}

func HandleProjectCreate(orgOverride string, clientID string, fields ProjectFields, changed map[string]bool, formatOverride string) error {
	trimmedClientID := strings.TrimSpace(clientID)
	if utils.IsBlank(trimmedClientID) {
		return fmt.Errorf("client id is required: use --client")
	}
	if utils.IsBlank(fields.Name) {
		return fmt.Errorf("name is required: use --name")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	payload := client.ProjectPayload{
		Name:         fields.Name,
		Color:        fields.Color,
		Subdivisions: parseCommaList(fields.Subdivisions),
	}
	if changed["daily-rate"] {
		payload.DailyRate = &fields.DailyRate
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	project, err := cli.CreateProject(orgID, trimmedClientID, payload)
	if err != nil {
		return err
	}

	renderProject(project, formatOverride)
	return nil
}

func mergeProjectFields(current client.Project, clientOverride string, fields ProjectFields, changed map[string]bool) client.ProjectPayload {
	payload := client.ProjectPayload{
		ClientID:     current.ClientID,
		Name:         current.Name,
		Color:        current.Color,
		DailyRate:    current.DailyRate,
		Subdivisions: current.Subdivisions,
	}

	if trimmedClient := strings.TrimSpace(clientOverride); utils.IsNotBlank(trimmedClient) {
		payload.ClientID = trimmedClient
	}
	if changed["name"] {
		payload.Name = fields.Name
	}
	if changed["color"] {
		payload.Color = fields.Color
	}
	if changed["daily-rate"] {
		payload.DailyRate = &fields.DailyRate
	}
	if changed["subdivisions"] {
		payload.Subdivisions = parseCommaList(fields.Subdivisions)
	}

	return payload
}

func HandleProjectUpdate(orgOverride string, id string, clientOverride string, fields ProjectFields, changed map[string]bool, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("project id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	projects, err := cli.ListProjects(orgID, utils.EMPTY)
	if err != nil {
		return err
	}

	current, found := findProject(projects, trimmedID)
	if !found {
		return fmt.Errorf("project %q not found", trimmedID)
	}

	payload := mergeProjectFields(current, clientOverride, fields, changed)
	if utils.IsBlank(payload.Name) {
		return fmt.Errorf("name is required: use --name")
	}

	updated, err := cli.UpdateProject(orgID, trimmedID, payload)
	if err != nil {
		return err
	}

	renderProject(updated, formatOverride)
	return nil
}

func HandleProjectDelete(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("project id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.DeleteProject(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func findProject(projects []client.Project, id string) (client.Project, bool) {
	for _, p := range projects {
		if p.ID == id {
			return p, true
		}
	}
	return client.Project{}, false
}

func renderProject(project client.Project, formatOverride string) {
	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(project)
		return
	}
	displayProjectsAsTable([]client.Project{project})
}

func displayProjectsAsTable(projects []client.Project) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Client ID", "Color", "Daily Rate"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(projects) {
		table.Append([]string{"No projects available", utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for _, p := range projects {
		table.Append([]string{p.ID, p.Name, p.ClientID, p.Color, dailyRateOrBlank(p.DailyRate)})
	}
	table.Render()
}

func dailyRateOrBlank(value *float64) string {
	if value == nil {
		return utils.EMPTY
	}
	return strconv.FormatFloat(*value, 'f', -1, 64)
}
