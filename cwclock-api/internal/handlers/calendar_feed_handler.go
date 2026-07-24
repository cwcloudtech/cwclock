package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

// CalendarFeedHandler lets a user subscribe to their own time entries from
// Outlook or Google Calendar (ai-instruct-85's "share the calendar"): an ICS
// feed URL, keyed by a per-user secret token instead of a session, since
// neither provider can send an Authorization header when polling a
// subscribed calendar.
type CalendarFeedHandler struct {
	users      *store.UserStore
	entries    *store.TimeEntryStore
	projects   *store.ProjectStore
	clients    *store.ClientStore
	apiBaseURL string
}

func NewCalendarFeedHandler(users *store.UserStore, entries *store.TimeEntryStore, projects *store.ProjectStore, clients *store.ClientStore, apiBaseURL string) *CalendarFeedHandler {
	return &CalendarFeedHandler{users: users, entries: entries, projects: projects, clients: clients, apiBaseURL: apiBaseURL}
}

type calendarFeedStatus struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
}

func (h *CalendarFeedHandler) statusResponse(user models.User) calendarFeedStatus {
	if !user.CalendarFeedEnabled || utils.IsBlank(user.CalendarFeedToken) {
		return calendarFeedStatus{Enabled: user.CalendarFeedEnabled}
	}
	return calendarFeedStatus{Enabled: true, URL: h.apiBaseURL + "/v1/calendar-feed/" + user.CalendarFeedToken}
}

func (h *CalendarFeedHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.statusResponse(user))
}

func (h *CalendarFeedHandler) Enable(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.users.SetCalendarFeedEnabled(r.Context(), userID, true)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.statusResponse(user))
}

func (h *CalendarFeedHandler) Disable(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.users.SetCalendarFeedEnabled(r.Context(), userID, false)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.statusResponse(user))
}

func (h *CalendarFeedHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.users.RegenerateCalendarFeedToken(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.statusResponse(user))
}

// Feed is the public, unauthenticated endpoint Outlook/Google Calendar poll
// once subscribed. There's no session to check - only the token embedded in
// the URL, same trust model as any other "share by link" URL.
func (h *CalendarFeedHandler) Feed(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	user, err := h.users.FindByCalendarFeedToken(r.Context(), token)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	entries, err := h.entries.ListAllForUser(r.Context(), user.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	projectCache := map[string]models.Project{}
	clientCache := map[string]models.Client{}
	summaryFor := func(entry models.TimeEntry) string {
		project, ok := projectCache[entry.ProjectID]
		if !ok {
			project, _ = h.projects.FindByID(r.Context(), entry.ProjectID)
			projectCache[entry.ProjectID] = project
		}
		client, ok := clientCache[entry.ClientID]
		if !ok {
			client, _ = h.clients.FindByID(r.Context(), entry.ClientID)
			clientCache[entry.ClientID] = client
		}

		if utils.IsBlank(client.Name) && utils.IsBlank(project.Name) {
			return entry.Text
		}

		return fmt.Sprintf("%s - %s: %s", client.Name, project.Name, entry.Text)
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `inline; filename="cwclock.ics"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(buildICS(entries, summaryFor)))
}

// icsEscape applies RFC 5545's TEXT escaping rules to a value used inside a
// content line (backslash, comma and semicolon are structural characters in
// the format, and a literal newline has to become the two-character "\n").
func icsEscape(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		"\n", `\n`,
		"\r", "",
		",", `\,`,
		";", `\;`,
	)
	return replacer.Replace(s)
}

// buildICS renders entries as a minimal RFC 5545 calendar. Entry times are
// emitted as floating local time (no TZID/UTC "Z" suffix): this app itself
// never stores a timezone alongside a day/start/end, so there isn't one to
// attach here either - Outlook/Google Calendar will interpret it in whatever
// timezone that calendar/device is set to, which for a personal feed is the
// same assumption the app already makes everywhere else.
func buildICS(entries []models.TimeEntry, summaryFor func(models.TimeEntry) string) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//CWClock//Calendar Feed//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("X-WR-CALNAME:CWClock\r\n")

	stamp := time.Now().UTC().Format("20060102T150405Z")
	for _, entry := range entries {
		day, err := time.Parse("2006-01-02", entry.Day)
		if err != nil {
			continue
		}

		b.WriteString("BEGIN:VEVENT\r\n")
		fmt.Fprintf(&b, "UID:%s@cwclock\r\n", entry.ID)
		fmt.Fprintf(&b, "DTSTAMP:%s\r\n", stamp)

		start, startErr := timeOfDay(day, entry.Start)
		end, endErr := timeOfDay(day, entry.End)
		if entry.AllDay || startErr != nil || endErr != nil {
			fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\r\n", day.Format("20060102"))
			fmt.Fprintf(&b, "DTEND;VALUE=DATE:%s\r\n", day.AddDate(0, 0, 1).Format("20060102"))
		} else {
			fmt.Fprintf(&b, "DTSTART:%s\r\n", start.Format("20060102T150405"))
			fmt.Fprintf(&b, "DTEND:%s\r\n", end.Format("20060102T150405"))
		}

		fmt.Fprintf(&b, "SUMMARY:%s\r\n", icsEscape(summaryFor(entry)))
		b.WriteString("END:VEVENT\r\n")
	}
	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func timeOfDay(day time.Time, hms *string) (time.Time, error) {
	if hms == nil || utils.IsBlank(*hms) {
		return time.Time{}, fmt.Errorf("no time of day")
	}
	return time.Parse("2006-01-02T15:04:05", day.Format("2006-01-02")+"T"+*hms)
}
