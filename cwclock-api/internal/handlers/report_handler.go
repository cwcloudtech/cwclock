package handlers

import (
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

func splitIDs(v string) []string {
	if utils.IsBlank(v) {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// reportContext parses the shared filter query params, loads the matching
// entries and enriches them, gating billable amounts on the caller's role
// (only admins/owner ever get an Amount; readers never reach this handler at
// all thanks to the router's RequireRole). On failure it writes the error
// response itself and returns ok=false so callers can just return.
func (h *ReportHandler) reportContext(w http.ResponseWriter, r *http.Request) (org models.Organization, filter store.ReportFilter, entries []models.ReportEntry, canSeeAmount bool, ok bool) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	role, _ := middleware.OrgRoleFromContext(r.Context())
	canSeeAmount = role == models.RoleAdmin || role == models.RoleOwner

	q := r.URL.Query()
	start, end := q.Get("start"), q.Get("end")
	if utils.IsBlank(start) || utils.IsBlank(end) {
		writeError(w, http.StatusBadRequest, "Please add start and end fields", CodeInvalidRequestBody)
		return models.Organization{}, store.ReportFilter{}, nil, false, false
	}

	filter = store.ReportFilter{
		Start:      start,
		End:        end,
		ClientIDs:  splitIDs(q.Get("clientIds")),
		ProjectIDs: splitIDs(q.Get("projectIds")),
		UserIDs:    splitIDs(q.Get("userIds")),
	}

	org, err := h.orgs.FindByID(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return models.Organization{}, store.ReportFilter{}, nil, false, false
	}

	rawEntries, err := h.entries.ListForReport(r.Context(), orgID, filter)
	if err != nil {
		writeStoreError(w, err)
		return models.Organization{}, store.ReportFilter{}, nil, false, false
	}
	clientsList, err := h.clients.List(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return models.Organization{}, store.ReportFilter{}, nil, false, false
	}
	projectsList, err := h.projects.List(r.Context(), orgID, "")
	if err != nil {
		writeStoreError(w, err)
		return models.Organization{}, store.ReportFilter{}, nil, false, false
	}
	members, err := h.orgs.ListMembers(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return models.Organization{}, store.ReportFilter{}, nil, false, false
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

	entries = report.Enrich(rawEntries, lk, canSeeAmount)
	return org, filter, entries, canSeeAmount, true
}

// Get returns the summary or detailed report as JSON, for the reports page.
func (h *ReportHandler) Get(w http.ResponseWriter, r *http.Request) {
	org, filter, entries, canSeeAmount, ok := h.reportContext(w, r)
	if !ok {
		return
	}

	if r.URL.Query().Get("type") == "detailed" {
		writeJSON(w, http.StatusOK, models.DetailedReport{
			Totals:  report.Totals(entries, canSeeAmount, org.Currency),
			Entries: entries,
		})
		return
	}

	start, _ := time.Parse(report.DayLayout, filter.Start)
	end, _ := time.Parse(report.DayLayout, filter.End)
	writeJSON(w, http.StatusOK, models.SummaryReport{
		Totals: report.Totals(entries, canSeeAmount, org.Currency),
		Daily:  report.DailyBuckets(entries, start, end),
		Rows:   report.SummaryRows(entries, canSeeAmount),
	})
}

func exportFilenameDate(day string) string {
	d, err := time.Parse(report.DayLayout, day)
	if err != nil {
		return day
	}
	return d.Format(report.FilenameDateLayout)
}

// Export streams the report as a downloadable CSV or PDF file, named per
// CWClock_Time_Report_{type}_{start}-{end}.{extension}.
func (h *ReportHandler) Export(w http.ResponseWriter, r *http.Request) {
	org, filter, entries, canSeeAmount, ok := h.reportContext(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	reportType := q.Get("type")
	format := q.Get("format")

	typeLabel := "Summary"
	if reportType == "detailed" {
		typeLabel = "Detailed"
	}
	filename := fmt.Sprintf("CWClock_Time_Report_%s_%s-%s.%s",
		typeLabel, exportFilenameDate(filter.Start), exportFilenameDate(filter.End), format)

	var data []byte
	var err error
	var contentType string

	switch {
	case reportType == "detailed" && format == "csv":
		contentType = "text/csv"
		data, err = report.DetailedCSV(entries, canSeeAmount, org.Currency)
	case reportType == "detailed" && format == "pdf":
		contentType = "application/pdf"
		logoData, logoType := report.ResolveLogo(org.Picture)
		data, err = report.DetailedPDF(org.Name, filter.Start, filter.End, models.DetailedReport{
			Totals:  report.Totals(entries, canSeeAmount, org.Currency),
			Entries: entries,
		}, logoData, logoType)
	case reportType == "summary" && format == "csv":
		contentType = "text/csv"
		data, err = report.SummaryCSV(report.SummaryRows(entries, canSeeAmount), canSeeAmount, org.Currency)
	case reportType == "summary" && format == "pdf":
		contentType = "application/pdf"
		logoData, logoType := report.ResolveLogo(org.Picture)
		start, _ := time.Parse(report.DayLayout, filter.Start)
		end, _ := time.Parse(report.DayLayout, filter.End)
		data, err = report.SummaryPDF(org.Name, filter.Start, filter.End, models.SummaryReport{
			Totals: report.Totals(entries, canSeeAmount, org.Currency),
			Daily:  report.DailyBuckets(entries, start, end),
			Rows:   report.SummaryRows(entries, canSeeAmount),
		}, logoData, logoType)
	default:
		writeError(w, http.StatusBadRequest, "Please specify a valid report type and format", CodeInvalidRequestBody)
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
