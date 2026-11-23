package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/holiday"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/notify"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

const (
	dayLayout        = "2006-01-02"
	maxFillRangeDays = 366
)

// WorkingDayHandler computes the working days in a date range for a client
// (weekends and the resolved country's public holidays are non-working), and
// fills them with all-day time entries. Country resolution is the client's own
// country first, then the organization's (so a French client billed by a
// Belgian org still gets the French holidays).
type WorkingDayHandler struct {
	auditor
	entries  *store.TimeEntryStore
	clients  *store.ClientStore
	orgs     *store.OrgStore
	projects *store.ProjectStore
	holidays *store.HolidayStore
	disc     *notify.Discord
}

func NewWorkingDayHandler(entries *store.TimeEntryStore, clients *store.ClientStore, orgs *store.OrgStore, projects *store.ProjectStore, holidays *store.HolidayStore, audit *store.AuditStore) *WorkingDayHandler {
	return &WorkingDayHandler{auditor: newAuditor(audit), entries: entries, clients: clients, orgs: orgs, projects: projects, holidays: holidays}
}

func (h *WorkingDayHandler) WithNotifier(disc *notify.Discord) *WorkingDayHandler {
	h.disc = disc
	return h
}

// dayClassification is the split of a date range into working days and the
// days skipped (with why), plus the resolved country and its holidays in range.
type dayClassification struct {
	country     string
	workingDays []string
	weekends    []string
	holidays    []map[string]string // {day, name}
}

// classify walks [from, to] inclusive and sorts each day into working / weekend
// / holiday using the resolved country's decision table. The weekend is
// whatever weekday rules the country configured; a country with no weekday
// rules falls back to Saturday/Sunday, so an unconfigured or holiday-only
// country still doesn't fill its weekends. A holiday that lands on a weekend is
// reported as the holiday (it is non-working either way).
func classify(country string, rules []models.PublicHoliday, from, to time.Time) dayClassification {
	cal := holiday.NewCalendar(rules)
	holidays := cal.Holidays(from, to)
	weekend := cal.NonWorkingWeekdays()
	useDefaultWeekend := len(weekend) == 0
	out := dayClassification{country: country}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := holiday.DayKey(d)
		switch {
		case holidays[key] != "":
			out.holidays = append(out.holidays, map[string]string{"day": key, "name": holidays[key]})
		case weekend[d.Weekday()] || (useDefaultWeekend && holiday.IsWeekend(d)):
			out.weekends = append(out.weekends, key)
		default:
			out.workingDays = append(out.workingDays, key)
		}
	}
	return out
}

// parseRange reads and validates from/to "YYYY-MM-DD" days, writing the error
// and returning ok=false when they're missing, malformed, inverted, or span
// more than a year.
func parseRange(w http.ResponseWriter, fromStr, toStr string) (time.Time, time.Time, bool) {
	if utils.IsBlank(fromStr) || utils.IsBlank(toStr) {
		writeError(w, http.StatusBadRequest, "Please provide from and to days (YYYY-MM-DD)", CodeInvalidDateRange)
		return time.Time{}, time.Time{}, false
	}
	from, err1 := time.ParseInLocation(dayLayout, strings.TrimSpace(fromStr), time.UTC)
	to, err2 := time.ParseInLocation(dayLayout, strings.TrimSpace(toStr), time.UTC)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "Days must be formatted YYYY-MM-DD", CodeInvalidDateRange)
		return time.Time{}, time.Time{}, false
	}
	if to.Before(from) {
		writeError(w, http.StatusBadRequest, "The end day must not be before the start day", CodeInvalidDateRange)
		return time.Time{}, time.Time{}, false
	}
	if to.Sub(from) > time.Duration(maxFillRangeDays)*24*time.Hour {
		writeError(w, http.StatusBadRequest, "The range must not exceed one year", CodeInvalidDateRange)
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}

// countryFor loads the client (checking it belongs to the org) and resolves the
// country to apply (client's own, else the org's). Writes the error and returns
// ok=false on failure.
func (h *WorkingDayHandler) countryFor(w http.ResponseWriter, r *http.Request, orgID, clientID string) (models.Client, string, bool) {
	client, err := h.clients.FindByID(r.Context(), clientID)
	if err != nil {
		writeStoreError(w, err)
		return models.Client{}, "", false
	}
	if client.OrganizationID != orgID {
		writeError(w, http.StatusBadRequest, "This client does not belong to the organization", CodeInvalidClientForOrg)
		return models.Client{}, "", false
	}
	country := strings.ToUpper(strings.TrimSpace(client.Country))
	if country == "" {
		if org, err := h.orgs.FindByID(r.Context(), orgID); err == nil {
			country = strings.ToUpper(strings.TrimSpace(org.Country))
		}
	}
	return client, country, true
}

// WorkingDays is the read-only preview: GET .../clients/{clientId}/working-days
// ?from=&to=&projectId=. It returns the working days, and the weekends/holidays
// skipped, for the resolved country. With an optional projectId, it also flags
// the working days that already have an entry for the caller on that project,
// so the UI and the assistant can show exactly what a fill would create.
func (h *WorkingDayHandler) WorkingDays(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())
	clientID := chi.URLParam(r, "clientId")

	from, to, ok := parseRange(w, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if !ok {
		return
	}
	_, country, ok := h.countryFor(w, r, orgID, clientID)
	if !ok {
		return
	}
	rules, err := h.holidays.ListByCountry(r.Context(), country)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	c := classify(country, rules, from, to)

	resp := map[string]any{
		"from":        holiday.DayKey(from),
		"to":          holiday.DayKey(to),
		"country":     country,
		"workingDays": c.workingDays,
		"weekends":    c.weekends,
		"holidays":    c.holidays,
	}
	if projectID := strings.TrimSpace(r.URL.Query().Get("projectId")); projectID != "" {
		logged, err := h.loggedDays(r, orgID, userID, projectID, holiday.DayKey(from), holiday.DayKey(to))
		if err == nil {
			resp["alreadyLogged"] = filterLogged(c.workingDays, logged)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

type fillPayload struct {
	ClientID  string `json:"clientId"`
	ProjectID string `json:"projectId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Text      string `json:"text"`
	Half      bool   `json:"half"`
}

// Fill is the guarded write: POST .../time-entries/fill creates one all-day
// (or half-day) entry per working day in [from, to] for the caller, skipping
// weekends, the resolved country's holidays, and any day already logged on this
// project. The assistant only calls this after the user has confirmed the days
// previewed by WorkingDays.
func (h *WorkingDayHandler) Fill(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p fillPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	if utils.IsBlank(p.ClientID) || utils.IsBlank(p.ProjectID) {
		writeError(w, http.StatusBadRequest, "Please add clientId and projectId fields", CodeTimeEntryFields)
		return
	}
	from, to, ok := parseRange(w, p.From, p.To)
	if !ok {
		return
	}
	client, country, ok := h.countryFor(w, r, orgID, p.ClientID)
	if !ok {
		return
	}

	// The project must belong to the same client, and its name is the default
	// entry text when the caller gave none.
	project, err := h.projects.FindByID(r.Context(), p.ProjectID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if project.ClientID != p.ClientID {
		writeError(w, http.StatusBadRequest, "This project does not belong to the client", CodeInvalidClientForOrg)
		return
	}
	text := strings.TrimSpace(p.Text)
	if text == "" {
		text = project.Name
	}

	rules, err := h.holidays.ListByCountry(r.Context(), country)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	c := classify(country, rules, from, to)
	logged, err := h.loggedDays(r, orgID, userID, p.ProjectID, holiday.DayKey(from), holiday.DayKey(to))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	created := []map[string]string{}
	skipped := []map[string]string{}
	for _, day := range c.weekends {
		skipped = append(skipped, map[string]string{"day": day, "reason": "weekend"})
	}
	for _, hol := range c.holidays {
		skipped = append(skipped, map[string]string{"day": hol["day"], "reason": "holiday", "name": hol["name"]})
	}
	for _, day := range c.workingDays {
		if logged[day] {
			skipped = append(skipped, map[string]string{"day": day, "reason": "exists"})
			continue
		}
		entry, err := h.entries.Create(r.Context(), orgID, p.ClientID, p.ProjectID, userID,
			store.TimeEntryFields{Text: text, Day: day, AllDay: !p.Half, Half: p.Half})
		if err != nil {
			skipped = append(skipped, map[string]string{"day": day, "reason": "error"})
			continue
		}
		created = append(created, map[string]string{"id": entry.ID, "day": day})
	}

	// One audit entry with the counts and range (like a CSV import), not one per
	// day, and one Discord notice summarising the fill.
	h.record(r, models.AuditRecordFilled, models.AuditSubjectRecord, p.ProjectID, map[string]any{
		"client":  client.Name,
		"project": project.Name,
		"from":    holiday.DayKey(from),
		"to":      holiday.DayKey(to),
		"created": len(created),
		"skipped": len(skipped),
	})
	notifyDiscord(h.disc, fmt.Sprintf("[RECORD] :calendar: filled %d working day(s) for %s / %s (%s..%s) by %s",
		len(created), client.Name, project.Name, holiday.DayKey(from), holiday.DayKey(to), actingEmail(r)))

	writeJSON(w, http.StatusCreated, map[string]any{
		"country": country,
		"created": created,
		"skipped": skipped,
	})
}

// loggedDays returns the set of days in [from, to] the caller already has an
// entry for on projectID, so a fill (or its preview) never double-books a day.
func (h *WorkingDayHandler) loggedDays(r *http.Request, orgID, userID, projectID, from, to string) (map[string]bool, error) {
	entries, err := h.entries.ListByRange(r.Context(), orgID, userID, from, to)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, e := range entries {
		if e.ProjectID == projectID {
			out[e.Day] = true
		}
	}
	return out, nil
}

// filterLogged returns the days of workingDays that are already logged.
func filterLogged(workingDays []string, logged map[string]bool) []string {
	out := []string{}
	for _, day := range workingDays {
		if logged[day] {
			out = append(out, day)
		}
	}
	return out
}
