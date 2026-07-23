package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/scheduler"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type ExportJobHandler struct {
	jobs      *store.ExportJobStore
	scheduler *scheduler.ExportJobScheduler
}

func NewExportJobHandler(jobs *store.ExportJobStore, sched *scheduler.ExportJobScheduler) *ExportJobHandler {
	return &ExportJobHandler{jobs: jobs, scheduler: sched}
}

type exportJobPayload struct {
	Name             string                   `json:"name"`
	CronExpression   string                   `json:"cronExpression"`
	Targets          []exportJobTargetPayload `json:"targets"`
	ReportTypes      []string                 `json:"reportTypes"`
	TimePeriod       string                   `json:"timePeriod"`
	ClientIDs        []string                 `json:"clientIds"`
	ProjectIDs       []string                 `json:"projectIds"`
	IncludeFinancial bool                     `json:"includeFinancial"`
	Enabled          bool                     `json:"enabled"`
}

type exportJobTargetPayload struct {
	Type       string                     `json:"type"`
	ToEmails   string                     `json:"toEmails,omitempty"`
	CCEmails   string                     `json:"ccEmails,omitempty"`
	Connection *models.ExternalConnection `json:"connection,omitempty"`
}

func (p exportJobPayload) nameValid() bool {
	return utils.IsNotBlank(p.Name)
}

func (p exportJobPayload) cronExpressionValid() bool {
	return utils.IsNotBlank(p.CronExpression) && scheduler.ValidCronExpression(p.CronExpression)
}

// targetsValid checks every target's shape, and - for a non-"email" target -
// validates and normalizes its embedded connection through the same
// per-type rules an organization's own external connections use (see
// validateExternalConnections), since it's the exact same struct captured
// through the exact same form fields, just stored independently in the
// job's own data instead of the organization's connections list. Unlike an
// organization's own connections, an export job target's connection is
// always forced flat (ai-instruct-78): the UI doesn't even offer the
// checkbox for a target, and a job's reports are named/timestamped by the
// export itself, so nesting them under YYYY/MM would just be redundant.
func (p exportJobPayload) targetsValid() bool {
	if len(p.Targets) == 0 {
		return false
	}
	allowedTypes := []string{"s3", "google_drive", "git", "email"}
	for i := range p.Targets {
		t := &p.Targets[i]
		if !slices.Contains(allowedTypes, t.Type) {
			return false
		}

		if t.Type == "email" {
			if len(utils.SplitList(t.ToEmails)) == 0 {
				return false
			}
			continue
		}

		if t.Connection == nil {
			return false
		}
		conns := []models.ExternalConnection{*t.Connection}
		if err := validateExternalConnections(conns); err != nil {
			return false
		}
		conns[0].FlatDirectory = true
		*t.Connection = conns[0]
	}
	return true
}

func (p exportJobPayload) reportTypesValid() bool {
	if len(p.ReportTypes) == 0 {
		return false
	}
	validTypes := map[string]bool{
		"summary-pdf":  true,
		"summary-csv":  true,
		"detailed-pdf": true,
		"detailed-csv": true,
		"invoices-pdf": true,
	}
	for _, rt := range p.ReportTypes {
		if !validTypes[rt] {
			return false
		}
	}
	return true
}

func (p exportJobPayload) timePeriodValid() bool {
	return utils.IsNotBlank(p.TimePeriod)
}

func (p exportJobPayload) toFields() store.ExportJobFields {
	return store.ExportJobFields{
		Name:             p.Name,
		CronExpression:   p.CronExpression,
		Targets:          convertTargets(p.Targets),
		ReportTypes:      p.ReportTypes,
		TimePeriod:       p.TimePeriod,
		ClientIDs:        p.ClientIDs,
		ProjectIDs:       p.ProjectIDs,
		IncludeFinancial: p.IncludeFinancial,
		Enabled:          p.Enabled,
	}
}

// exportJobResponse adds the job's next scheduled run to the JSON response
// - computed on the fly from its cron expression (see scheduler.NextRunAt)
// rather than stored, so it's always accurate as of the response and never
// needs to be kept in sync with edits. Only set for an enabled job: a
// disabled one has no next run to report. NextRunInSeconds is the exact
// remaining duration as computed by this same server clock (ai-instruct-79)
// - the frontend counts down from it locally instead of diffing its own
// clock against NextRunAt, which would drift if the browser's clock isn't
// in sync with the server's.
type exportJobResponse struct {
	models.ExportJob
	NextRunAt        *time.Time `json:"nextRunAt,omitempty"`
	NextRunInSeconds *int64     `json:"nextRunInSeconds,omitempty"`
}

func toExportJobResponse(job models.ExportJob) exportJobResponse {
	resp := exportJobResponse{ExportJob: job}
	if !job.Enabled {
		return resp
	}
	next := scheduler.NextRunAt(job.CronExpression)
	if next == nil {
		return resp
	}
	resp.NextRunAt = next
	seconds := max(int64(time.Until(*next).Round(time.Second).Seconds()), 0)
	resp.NextRunInSeconds = &seconds
	return resp
}

func convertTargets(payloadTargets []exportJobTargetPayload) []models.ExportTarget {
	targets := make([]models.ExportTarget, len(payloadTargets))
	for i, t := range payloadTargets {
		targets[i] = models.ExportTarget{
			Type:       t.Type,
			ToEmails:   t.ToEmails,
			CCEmails:   t.CCEmails,
			Connection: t.Connection,
		}
	}
	return targets
}

func (h *ExportJobHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	jobs, err := h.jobs.List(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	resp := make([]exportJobResponse, len(jobs))
	for i, job := range jobs {
		resp[i] = toExportJobResponse(job)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ExportJobHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var p exportJobPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}

	if !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if !p.cronExpressionValid() {
		writeError(w, http.StatusBadRequest, "Please provide a valid cron expression", CodeInvalidRequestBody)
		return
	}
	if !p.targetsValid() {
		writeError(w, http.StatusBadRequest, "Please add at least one valid target", CodeInvalidRequestBody)
		return
	}
	if !p.reportTypesValid() {
		writeError(w, http.StatusBadRequest, "Please select at least one report type", CodeInvalidRequestBody)
		return
	}
	if !p.timePeriodValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Time Period field", CodeInvalidRequestBody)
		return
	}

	job, err := h.jobs.Create(r.Context(), orgID, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if err := h.scheduler.ScheduleJob(job); err != nil {
		slog.Error("failed to schedule export job", "error", err, "jobId", job.ID)
	}
	writeJSON(w, http.StatusCreated, toExportJobResponse(job))
}

func (h *ExportJobHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "jobId")

	var p exportJobPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}

	if !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if !p.cronExpressionValid() {
		writeError(w, http.StatusBadRequest, "Please provide a valid cron expression", CodeInvalidRequestBody)
		return
	}
	if !p.targetsValid() {
		writeError(w, http.StatusBadRequest, "Please add at least one valid target", CodeInvalidRequestBody)
		return
	}
	if !p.reportTypesValid() {
		writeError(w, http.StatusBadRequest, "Please select at least one report type", CodeInvalidRequestBody)
		return
	}
	if !p.timePeriodValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Time Period field", CodeInvalidRequestBody)
		return
	}

	job, err := h.jobs.Update(r.Context(), id, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if err := h.scheduler.ScheduleJob(job); err != nil {
		slog.Error("failed to reschedule export job", "error", err, "jobId", job.ID)
	}
	writeJSON(w, http.StatusOK, toExportJobResponse(job))
}

func (h *ExportJobHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "jobId")

	if err := h.jobs.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	h.scheduler.UnscheduleJob(id)
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
