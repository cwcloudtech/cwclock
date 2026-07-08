package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/report"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type ReportHandler struct {
	orgs     *store.OrgStore
	clients  *store.ClientStore
	projects *store.ProjectStore
	entries  *store.TimeEntryStore
}

func NewReportHandler(orgs *store.OrgStore, clients *store.ClientStore, projects *store.ProjectStore, entries *store.TimeEntryStore) *ReportHandler {
	return &ReportHandler{orgs: orgs, clients: clients, projects: projects, entries: entries}
}

// idFilter mirrors the {ids, contains, status} shape of a Clockify-style
// report filter. Only ids is honored - contains/status are accepted (so a
// request using them doesn't fail) but not applied.
type idFilter struct {
	IDs []string `json:"ids"`
}

// detailedFilter mirrors Clockify's detailedFilter object. sortColumn and
// sortOrder are intentionally not bound here, which both ignores them and
// (since Go's json.Decoder ignores unknown fields by default) means any
// other unrecognized field never causes a decode failure either.
type detailedFilter struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// exportRequest is the JSON body accepted by POST .../reports/detailed and
// .../reports/summary, shaped after the payload cwclock's own export
// scripts already send to a Clockify-compatible reports API.
type exportRequest struct {
	ExportType     string          `json:"exportType"`
	DateRangeStart string          `json:"dateRangeStart"`
	DateRangeEnd   string          `json:"dateRangeEnd"`
	Clients        *idFilter       `json:"clients"`
	Projects       *idFilter       `json:"projects"`
	Users          *idFilter       `json:"users"`
	DetailedFilter *detailedFilter `json:"detailedFilter"`
}

// dayPart extracts the leading "YYYY-MM-DD" from a full timestamp like
// "2021-06-26T06:00:00.000Z", matching the bare-date strings ReportFilter
// and time_entries.data->>'day' are compared against everywhere else.
func dayPart(v string) string {
	if len(v) >= len(report.DayLayout) {
		return v[:len(report.DayLayout)]
	}
	return v
}

func (req exportRequest) filter() store.ReportFilter {
	f := store.ReportFilter{Start: dayPart(req.DateRangeStart), End: dayPart(req.DateRangeEnd)}
	if req.Clients != nil {
		f.ClientIDs = req.Clients.IDs
	}
	if req.Projects != nil {
		f.ProjectIDs = req.Projects.IDs
	}
	if req.Users != nil {
		f.UserIDs = req.Users.IDs
	}
	return f
}

// decodeExportRequest reads and validates the shared request body. On
// failure it writes the error response itself and returns ok=false so
// callers can just return.
func decodeExportRequest(w http.ResponseWriter, r *http.Request) (exportRequest, bool) {
	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return exportRequest{}, false
	}
	if utils.IsBlank(req.DateRangeStart) || utils.IsBlank(req.DateRangeEnd) {
		writeError(w, http.StatusBadRequest, "Please add dateRangeStart and dateRangeEnd fields", CodeInvalidRequestBody)
		return exportRequest{}, false
	}
	return req, true
}

// loadEnrichedEntries loads every time entry matching filter and enriches
// it with the display data a report needs (client/project/member names)
// plus its computed duration and, when the caller is allowed to see it, its
// billable amount.
func (h *ReportHandler) loadEnrichedEntries(ctx context.Context, orgID string, filter store.ReportFilter, canSeeAmount bool) (models.Organization, []models.ReportEntry, error) {
	org, err := h.orgs.FindByID(ctx, orgID)
	if err != nil {
		return models.Organization{}, nil, err
	}
	rawEntries, err := h.entries.ListForReport(ctx, orgID, filter)
	if err != nil {
		return models.Organization{}, nil, err
	}
	clientsList, err := h.clients.List(ctx, orgID)
	if err != nil {
		return models.Organization{}, nil, err
	}
	projectsList, err := h.projects.List(ctx, orgID, "")
	if err != nil {
		return models.Organization{}, nil, err
	}
	members, err := h.orgs.ListMembers(ctx, orgID)
	if err != nil {
		return models.Organization{}, nil, err
	}

	lk := report.Lookups{
		Clients:  make(map[string]models.Client, len(clientsList)),
		Projects: make(map[string]models.Project, len(projectsList)),
		Members:  make(map[string]models.Member, len(members)),
	}
	for _, c := range clientsList {
		lk.Clients[c.ID] = c
	}
	for _, p := range projectsList {
		lk.Projects[p.ID] = p
	}
	for _, m := range members {
		lk.Members[m.UserID] = m
	}

	return org, report.Enrich(rawEntries, lk, canSeeAmount), nil
}

func exportFilenameDate(day string) string {
	d, err := time.Parse(report.DayLayout, day)
	if err != nil {
		return day
	}
	return d.Format(report.FilenameDateLayout)
}

func writeExportFile(w http.ResponseWriter, contentType, filename string, data []byte, err error) {
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// Detailed serves the flat, per-entry report - as JSON for the reports page
// (exportType absent/"JSON"), or streamed as a PDF/CSV file for scripted
// exports (exportType "PDF"/"CSV"), matching the payload shape cwclock's
// own export scripts already send. detailedFilter.page/pageSize, when set,
// only pages the JSON entry list; totals and file exports always cover the
// full requested range.
func (h *ReportHandler) Detailed(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	role, _ := middleware.OrgRoleFromContext(r.Context())
	canSeeAmount := role == models.RoleAdmin || role == models.RoleOwner

	req, ok := decodeExportRequest(w, r)
	if !ok {
		return
	}
	filter := req.filter()

	org, entries, err := h.loadEnrichedEntries(r.Context(), orgID, filter, canSeeAmount)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	totals := report.Totals(entries, canSeeAmount, org.Currency)

	switch strings.ToUpper(req.ExportType) {
	case "PDF":
		logoData, logoType := report.ResolveLogo(org.Picture)
		data, err := report.DetailedPDF(org.Name, filter.Start, filter.End, models.DetailedReport{Totals: totals, Entries: entries}, logoData, logoType)
		filename := fmt.Sprintf("CWClock_Time_Report_Detailed_%s-%s.pdf", exportFilenameDate(filter.Start), exportFilenameDate(filter.End))
		writeExportFile(w, "application/pdf", filename, data, err)
	case "CSV":
		data, err := report.DetailedCSV(entries, canSeeAmount, org.Currency)
		filename := fmt.Sprintf("CWClock_Time_Report_Detailed_%s-%s.csv", exportFilenameDate(filter.Start), exportFilenameDate(filter.End))
		writeExportFile(w, "text/csv", filename, data, err)
	default:
		pageEntries := entries
		if req.DetailedFilter != nil && req.DetailedFilter.Page > 0 && req.DetailedFilter.PageSize > 0 {
			start := min((req.DetailedFilter.Page-1)*req.DetailedFilter.PageSize, len(entries))
			end := min(start+req.DetailedFilter.PageSize, len(entries))
			pageEntries = entries[start:end]
		}
		writeJSON(w, http.StatusOK, models.DetailedReport{Totals: totals, Entries: pageEntries})
	}
}

// Summary serves the aggregated report (rows grouped by project+description
// +user, plus a per-day duration chart) - as JSON for the reports page, or
// streamed as a PDF/CSV file for scripted exports. Same request contract as
// Detailed.
func (h *ReportHandler) Summary(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	role, _ := middleware.OrgRoleFromContext(r.Context())
	canSeeAmount := role == models.RoleAdmin || role == models.RoleOwner

	req, ok := decodeExportRequest(w, r)
	if !ok {
		return
	}
	filter := req.filter()

	org, entries, err := h.loadEnrichedEntries(r.Context(), orgID, filter, canSeeAmount)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	start, _ := time.Parse(report.DayLayout, filter.Start)
	end, _ := time.Parse(report.DayLayout, filter.End)
	summary := models.SummaryReport{
		Totals: report.Totals(entries, canSeeAmount, org.Currency),
		Daily:  report.DailyBuckets(entries, start, end),
		Rows:   report.SummaryRows(entries, canSeeAmount),
	}

	switch strings.ToUpper(req.ExportType) {
	case "PDF":
		logoData, logoType := report.ResolveLogo(org.Picture)
		data, err := report.SummaryPDF(org.Name, filter.Start, filter.End, summary, logoData, logoType)
		filename := fmt.Sprintf("CWClock_Time_Report_Summary_%s-%s.pdf", exportFilenameDate(filter.Start), exportFilenameDate(filter.End))
		writeExportFile(w, "application/pdf", filename, data, err)
	case "CSV":
		data, err := report.SummaryCSV(summary.Rows, canSeeAmount, org.Currency)
		filename := fmt.Sprintf("CWClock_Time_Report_Summary_%s-%s.csv", exportFilenameDate(filter.Start), exportFilenameDate(filter.End))
		writeExportFile(w, "text/csv", filename, data, err)
	default:
		writeJSON(w, http.StatusOK, summary)
	}
}
