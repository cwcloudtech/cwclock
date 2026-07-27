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

type ExportJobFields struct {
	Name             string
	Cron             string
	ReportTypes      string
	TimePeriod       string
	ClientIDs        []string
	ProjectIDs       []string
	IncludeFinancial bool
	Enabled          bool
}

// ExportTargetFields is the flag-facing shape of an export target, shared
// by "job create" (the job's initial target) and "job target create"
// (appending one to an existing job).
type ExportTargetFields struct {
	Type                    string
	To                      string
	CC                      string
	Endpoint                string
	BucketName              string
	Region                  string
	AccessKey               string
	SecretKey               string
	ServiceAccountBase64    string
	FolderID                string
	RepoURL                 string
	Username                string
	Password                string
	SSHPrivateKey           string
	SSHPrivateKeyPassphrase string
	Path                    string
}

func (f ExportTargetFields) hasAny() bool {
	return utils.IsNotBlank(f.Type) || utils.IsNotBlank(f.To) || utils.IsNotBlank(f.CC) ||
		utils.IsNotBlank(f.Endpoint) || utils.IsNotBlank(f.BucketName) || utils.IsNotBlank(f.Region) ||
		utils.IsNotBlank(f.AccessKey) || utils.IsNotBlank(f.SecretKey) ||
		utils.IsNotBlank(f.ServiceAccountBase64) || utils.IsNotBlank(f.FolderID) ||
		utils.IsNotBlank(f.RepoURL) || utils.IsNotBlank(f.Username) || utils.IsNotBlank(f.Password) ||
		utils.IsNotBlank(f.SSHPrivateKey) || utils.IsNotBlank(f.SSHPrivateKeyPassphrase)
}

func (f ExportTargetFields) toTarget() (client.ExportTarget, error) {
	targetType := strings.TrimSpace(f.Type)
	if utils.IsBlank(targetType) {
		targetType = "email"
	}

	switch targetType {
	case "email":
		if utils.IsBlank(f.To) {
			return client.ExportTarget{}, fmt.Errorf("--to is required for an email target")
		}
		return client.ExportTarget{Type: targetType, ToEmails: f.To, CCEmails: f.CC}, nil
	case "s3", "google_drive", "git":
		return client.ExportTarget{
			Type: targetType,
			Connection: &client.ExternalConnection{
				Type:                    targetType,
				Endpoint:                f.Endpoint,
				BucketName:              f.BucketName,
				Region:                  f.Region,
				AccessKey:               f.AccessKey,
				SecretKey:               f.SecretKey,
				ServiceAccountBase64:    f.ServiceAccountBase64,
				FolderID:                f.FolderID,
				RepoURL:                 f.RepoURL,
				Username:                f.Username,
				Password:                f.Password,
				SSHPrivateKey:           f.SSHPrivateKey,
				SSHPrivateKeyPassphrase: f.SSHPrivateKeyPassphrase,
				Path:                    f.Path,
			},
		}, nil
	default:
		return client.ExportTarget{}, fmt.Errorf("invalid target type %q: expected email, s3, google_drive or git", targetType)
	}
}

func (f ExportJobFields) toPayload(orgID string, targets []client.ExportTarget) (client.ExportJobPayload, error) {
	clientIDs, err := resolveClientIDs(orgID, f.ClientIDs)
	if err != nil {
		return client.ExportJobPayload{}, err
	}
	projects, err := resolveProjects(orgID, singleClientScope(clientIDs), f.ProjectIDs)
	if err != nil {
		return client.ExportJobPayload{}, err
	}
	if err := requireProjectsMatchClients(clientIDs, projects); err != nil {
		return client.ExportJobPayload{}, err
	}

	return client.ExportJobPayload{
		Name:             f.Name,
		CronExpression:   f.Cron,
		Targets:          targets,
		ReportTypes:      parseCommaList(f.ReportTypes),
		TimePeriod:       f.TimePeriod,
		ClientIDs:        clientIDs,
		ProjectIDs:       projectIDsOf(projects),
		IncludeFinancial: f.IncludeFinancial,
		Enabled:          f.Enabled,
	}, nil
}

func exportJobToPayload(job client.ExportJob) client.ExportJobPayload {
	return client.ExportJobPayload{
		Name:             job.Name,
		CronExpression:   job.CronExpression,
		Targets:          job.Targets,
		ReportTypes:      job.ReportTypes,
		TimePeriod:       job.TimePeriod,
		ClientIDs:        job.ClientIDs,
		ProjectIDs:       job.ProjectIDs,
		IncludeFinancial: job.IncludeFinancial,
		Enabled:          job.Enabled,
	}
}

func mergeExportJobFields(orgID string, current client.ExportJob, fields ExportJobFields, changed map[string]bool) (client.ExportJobPayload, error) {
	payload := exportJobToPayload(current)

	if changed["name"] {
		payload.Name = fields.Name
	}
	if changed["cron"] {
		payload.CronExpression = fields.Cron
	}
	if changed["report-types"] {
		payload.ReportTypes = parseCommaList(fields.ReportTypes)
	}
	if changed["time-period"] {
		payload.TimePeriod = fields.TimePeriod
	}
	// Resolve the (possibly updated) client scope first, so a project
	// resolution below can be scoped to it when unambiguous (see
	// ai-instruct-100).
	if changed["client"] {
		clientIDs, err := resolveClientIDs(orgID, fields.ClientIDs)
		if err != nil {
			return client.ExportJobPayload{}, err
		}
		payload.ClientIDs = clientIDs
	}

	// projects only needs to be populated (for the cross-check below) when
	// either side of the --client/--project pair could have changed -
	// otherwise there's nothing new to validate against the other.
	var projects []client.Project
	if changed["project"] {
		resolved, err := resolveProjects(orgID, singleClientScope(payload.ClientIDs), fields.ProjectIDs)
		if err != nil {
			return client.ExportJobPayload{}, err
		}
		projects = resolved
		payload.ProjectIDs = projectIDsOf(resolved)
	} else if changed["client"] && len(payload.ProjectIDs) > 0 {
		// Deliberately unscoped: these are the job's existing project ids,
		// which may well belong to a client other than the new one - that's
		// exactly the mismatch requireProjectsMatchClients needs to catch,
		// so don't narrow the search to the new client first.
		resolved, err := resolveProjects(orgID, utils.EMPTY, payload.ProjectIDs)
		if err != nil {
			return client.ExportJobPayload{}, err
		}
		projects = resolved
	}

	if err := requireProjectsMatchClients(payload.ClientIDs, projects); err != nil {
		return client.ExportJobPayload{}, err
	}

	if changed["include-financial"] {
		payload.IncludeFinancial = fields.IncludeFinancial
	}
	if changed["enabled"] {
		payload.Enabled = fields.Enabled
	}

	return payload, nil
}

func validateExportJobPayload(payload client.ExportJobPayload) error {
	if utils.IsBlank(payload.Name) {
		return fmt.Errorf("name is required: use --name")
	}
	if utils.IsBlank(payload.CronExpression) {
		return fmt.Errorf("cron expression is required: use --cron")
	}
	if len(payload.ReportTypes) == 0 {
		return fmt.Errorf("at least one report type is required: use --report-types")
	}
	if utils.IsBlank(payload.TimePeriod) {
		return fmt.Errorf("time period is required: use --time-period")
	}
	if len(payload.Targets) == 0 {
		return fmt.Errorf("at least one target is required: use --to for an email target, or --type s3/google_drive/git with its connection fields")
	}
	return nil
}

func HandleJobList(orgOverride string, formatOverride string) error {
	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	jobs, err := cli.ListExportJobs(orgID)
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(jobs)
		return nil
	}

	displayExportJobsAsTable(jobs)
	return nil
}

func HandleJobCreate(orgOverride string, fields ExportJobFields, targetFields ExportTargetFields, formatOverride string) error {
	if !targetFields.hasAny() {
		return fmt.Errorf("at least one target is required: use --to for an email target, or --type s3/google_drive/git with its connection fields")
	}

	target, err := targetFields.toTarget()
	if err != nil {
		return err
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	payload, err := fields.toPayload(orgID, []client.ExportTarget{target})
	if err != nil {
		return err
	}
	if err := validateExportJobPayload(payload); err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	job, err := cli.CreateExportJob(orgID, payload)
	if err != nil {
		return err
	}

	renderExportJob(job, formatOverride)
	return nil
}

func HandleJobUpdate(orgOverride string, id string, fields ExportJobFields, changed map[string]bool, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("job id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	jobs, err := cli.ListExportJobs(orgID)
	if err != nil {
		return err
	}

	current, found := findExportJob(jobs, trimmedID)
	if !found {
		return fmt.Errorf("job %q not found", trimmedID)
	}

	payload, err := mergeExportJobFields(orgID, current, fields, changed)
	if err != nil {
		return err
	}
	if err := validateExportJobPayload(payload); err != nil {
		return err
	}

	updated, err := cli.UpdateExportJob(orgID, trimmedID, payload)
	if err != nil {
		return err
	}

	renderExportJob(updated, formatOverride)
	return nil
}

func HandleJobDelete(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("job id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.DeleteExportJob(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

// HandleJobRun triggers an immediate, out-of-schedule run of an export job -
// see client.RunExportJob.
func HandleJobRun(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("job id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.RunExportJob(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func HandleJobTargetList(orgOverride string, jobID string, formatOverride string) error {
	trimmedID := strings.TrimSpace(jobID)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("job id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	jobs, err := cli.ListExportJobs(orgID)
	if err != nil {
		return err
	}

	current, found := findExportJob(jobs, trimmedID)
	if !found {
		return fmt.Errorf("job %q not found", trimmedID)
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(current.Targets)
		return nil
	}

	displayExportTargetsAsTable(current.Targets)
	return nil
}

func HandleJobTargetCreate(orgOverride string, jobID string, targetFields ExportTargetFields, formatOverride string) error {
	trimmedID := strings.TrimSpace(jobID)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("job id is required: use -i or --id")
	}
	if !targetFields.hasAny() {
		return fmt.Errorf("at least one target field is required: use --to for an email target, or --type s3/google_drive/git with its connection fields")
	}

	target, err := targetFields.toTarget()
	if err != nil {
		return err
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	jobs, err := cli.ListExportJobs(orgID)
	if err != nil {
		return err
	}

	current, found := findExportJob(jobs, trimmedID)
	if !found {
		return fmt.Errorf("job %q not found", trimmedID)
	}

	payload := exportJobToPayload(current)
	payload.Targets = append(payload.Targets, target)

	updated, err := cli.UpdateExportJob(orgID, trimmedID, payload)
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(updated.Targets)
		return nil
	}

	displayExportTargetsAsTable(updated.Targets)
	return nil
}

func HandleJobTargetDelete(orgOverride string, jobID string, offset int) error {
	trimmedID := strings.TrimSpace(jobID)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("job id is required: use -i or --id")
	}
	if offset < 0 {
		return fmt.Errorf("offset is required: use --offset")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	jobs, err := cli.ListExportJobs(orgID)
	if err != nil {
		return err
	}

	current, found := findExportJob(jobs, trimmedID)
	if !found {
		return fmt.Errorf("job %q not found", trimmedID)
	}

	if offset < 0 || offset >= len(current.Targets) {
		return fmt.Errorf("offset %d is out of range: job %q has %d target(s)", offset, trimmedID, len(current.Targets))
	}

	remaining := make([]client.ExportTarget, 0, len(current.Targets)-1)
	remaining = append(remaining, current.Targets[:offset]...)
	remaining = append(remaining, current.Targets[offset+1:]...)

	payload := exportJobToPayload(current)
	payload.Targets = remaining

	if _, err := cli.UpdateExportJob(orgID, trimmedID, payload); err != nil {
		return err
	}

	fmt.Printf("offset = %d\n", offset)
	return nil
}

func findExportJob(jobs []client.ExportJob, id string) (client.ExportJob, bool) {
	for _, job := range jobs {
		if job.ID == id {
			return job, true
		}
	}
	return client.ExportJob{}, false
}

func renderExportJob(job client.ExportJob, formatOverride string) {
	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(job)
		return
	}
	displayExportJobsAsTable([]client.ExportJob{job})
}

func displayExportJobsAsTable(jobs []client.ExportJob) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Cron", "Time Period", "Enabled", "Targets"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(jobs) {
		table.Append([]string{"No export jobs available", utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for _, job := range jobs {
		table.Append([]string{
			job.ID, job.Name, job.CronExpression, job.TimePeriod,
			strconv.FormatBool(job.Enabled), strconv.Itoa(len(job.Targets)),
		})
	}
	table.Render()
}

func displayExportTargetsAsTable(targets []client.ExportTarget) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Offset", "Type", "To", "CC", "Connection"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(targets) {
		table.Append([]string{utils.EMPTY, "No targets available", utils.EMPTY, utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for i, target := range targets {
		table.Append([]string{
			strconv.Itoa(i), target.Type, target.ToEmails, target.CCEmails,
			connectionSummary(target.Connection),
		})
	}
	table.Render()
}

func connectionSummary(conn *client.ExternalConnection) string {
	if conn == nil {
		return utils.EMPTY
	}
	switch conn.Type {
	case "s3":
		return fmt.Sprintf("s3://%s/%s", conn.BucketName, conn.Path)
	case "google_drive":
		return fmt.Sprintf("google_drive:%s", conn.FolderID)
	case "git":
		return fmt.Sprintf("git:%s", conn.RepoURL)
	default:
		return conn.Type
	}
}
