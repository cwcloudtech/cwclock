package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/holiday"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

// HolidayHandler serves the superuser-only public-holiday decision table
// (deployment-wide, keyed by ISO country code). It's what the admin "Jours
// fériés" screen drives; the working-day resolver reads the same table.
type HolidayHandler struct {
	auditor
	holidays *store.HolidayStore
}

func NewHolidayHandler(holidays *store.HolidayStore, audit *store.AuditStore) *HolidayHandler {
	return &HolidayHandler{auditor: newAuditor(audit), holidays: holidays}
}

type holidayPayload struct {
	Country      string             `json:"country"`
	Name         string             `json:"name"`
	Type         models.HolidayType `json:"type"`
	Month        int                `json:"month"`
	Day          int                `json:"day"`
	EasterOffset int                `json:"easterOffset"`
	Weekday      string             `json:"weekday"`
	Nth          int                `json:"nth"`
}

// validate checks a holiday rule and returns the normalised country code, or
// writes the error and returns ok=false. A fixed rule needs a real month/day;
// an easter rule needs only its offset (any day of the year, including the
// Easter-Sunday anchor itself).
func (p holidayPayload) validate(w http.ResponseWriter) (string, store.HolidayFields, bool) {
	country := strings.ToUpper(strings.TrimSpace(p.Country))
	name := strings.TrimSpace(p.Name)
	if len(country) != 2 {
		writeError(w, http.StatusBadRequest, "Please provide a 2-letter ISO country code", CodeInvalidCountry)
		return "", store.HolidayFields{}, false
	}
	if name == "" {
		writeError(w, http.StatusBadRequest, "Please provide a holiday name", CodeInvalidHoliday)
		return "", store.HolidayFields{}, false
	}
	f := store.HolidayFields{Name: name, Type: p.Type, Month: p.Month, Day: p.Day, EasterOffset: p.EasterOffset, Weekday: strings.ToLower(strings.TrimSpace(p.Weekday)), Nth: p.Nth}
	switch p.Type {
	case models.HolidayFixed:
		if p.Month < 1 || p.Month > 12 || p.Day < 1 || p.Day > 31 {
			writeError(w, http.StatusBadRequest, "A fixed holiday needs a month (1-12) and day (1-31)", CodeInvalidHoliday)
			return "", store.HolidayFields{}, false
		}
		f.EasterOffset, f.Weekday, f.Nth = 0, "", 0
	case models.HolidayEaster:
		if p.EasterOffset < -60 || p.EasterOffset > 60 {
			writeError(w, http.StatusBadRequest, "An easter-relative holiday needs an offset between -60 and 60 days", CodeInvalidHoliday)
			return "", store.HolidayFields{}, false
		}
		f.Month, f.Day, f.Weekday, f.Nth = 0, 0, "", 0
	case models.HolidayWeekend:
		if _, ok := holiday.ParseWeekday(f.Weekday); !ok {
			writeError(w, http.StatusBadRequest, "A weekend rule needs a weekday (monday..sunday)", CodeInvalidHoliday)
			return "", store.HolidayFields{}, false
		}
		f.Month, f.Day, f.EasterOffset, f.Nth = 0, 0, 0, 0
	case models.HolidayNthWeekday:
		if _, ok := holiday.ParseWeekday(f.Weekday); !ok || p.Month < 1 || p.Month > 12 {
			writeError(w, http.StatusBadRequest, "An nth-weekday rule needs a month (1-12) and a weekday (monday..sunday)", CodeInvalidHoliday)
			return "", store.HolidayFields{}, false
		}
		if p.Nth != -1 && (p.Nth < 1 || p.Nth > 5) {
			writeError(w, http.StatusBadRequest, "The occurrence must be 1-5, or -1 for the last", CodeInvalidHoliday)
			return "", store.HolidayFields{}, false
		}
		f.Day, f.EasterOffset = 0, 0
	default:
		writeError(w, http.StatusBadRequest, "Holiday type must be 'fixed', 'easter', 'weekend' or 'nthWeekday'", CodeInvalidHoliday)
		return "", store.HolidayFields{}, false
	}
	return country, f, true
}

// List returns the holiday rules, optionally filtered by ?country=. Superuser.
func (h *HolidayHandler) List(w http.ResponseWriter, r *http.Request) {
	holidays, err := h.holidays.List(r.Context(), r.URL.Query().Get("country"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, holidays)
}

func (h *HolidayHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p holidayPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	country, f, ok := p.validate(w)
	if !ok {
		return
	}
	holiday, err := h.holidays.Create(r.Context(), country, f)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditHolidayCreated, models.AuditSubjectHoliday, holiday.ID,
		map[string]any{"country": holiday.Country, "name": holiday.Name})
	writeJSON(w, http.StatusCreated, holiday)
}

func (h *HolidayHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p holidayPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	country, f, ok := p.validate(w)
	if !ok {
		return
	}
	holiday, err := h.holidays.Update(r.Context(), id, country, f)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditHolidayUpdated, models.AuditSubjectHoliday, holiday.ID,
		map[string]any{"country": holiday.Country, "name": holiday.Name})
	writeJSON(w, http.StatusOK, holiday)
}

func (h *HolidayHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Named before the delete so the trail still says what went.
	var country, name string
	if before, err := h.holidays.FindByID(r.Context(), id); err == nil {
		country, name = before.Country, before.Name
	}
	if err := h.holidays.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditHolidayDeleted, models.AuditSubjectHoliday, id,
		map[string]any{"country": country, "name": name})
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
