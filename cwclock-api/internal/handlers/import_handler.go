package handlers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/store"
)

// ImportHandler handles bulk data imports from third-party tools.
type ImportHandler struct {
	users       *store.UserStore
	clients     *store.ClientStore
	projects    *store.ProjectStore
	timeEntries *store.TimeEntryStore
}

func NewImportHandler(
	users *store.UserStore,
	clients *store.ClientStore,
	projects *store.ProjectStore,
	timeEntries *store.TimeEntryStore,
) *ImportHandler {
	return &ImportHandler{
		users:       users,
		clients:     clients,
		projects:    projects,
		timeEntries: timeEntries,
	}
}

type importResult struct {
	Created int `json:"created"`
	Skipped int `json:"skipped"`
}

const (
	csvDateLayout     = "01/02/2006"  // MM/DD/YYYY
	csvTimeLayout     = "03:04:05 PM" // 12-hour with AM/PM
	importDayLayout   = "2006-01-02"  // YYYY-MM-DD stored in time_entries
	maxImportBodySize = 4 << 20       // 4 MB
)

// parseCSVTime converts a 12-hour "HH:MM:SS AM|PM" string to 24-hour
// "HH:MM:SS".
func parseCSVTime(s string) (string, error) {
	t, err := time.Parse(csvTimeLayout, strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("invalid time %q: %w", s, err)
	}
	return t.Format("15:04:05"), nil
}

// parseCSVDate converts "MM/DD/YYYY" to "YYYY-MM-DD".
func parseCSVDate(s string) (string, error) {
	t, err := time.Parse(csvDateLayout, strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("invalid date %q: %w", s, err)
	}
	return t.Format(importDayLayout), nil
}

// randomColor returns a random 6-digit hex color string like "#a3b4c5".
func randomColor() string {
	return fmt.Sprintf("#%06x", rand.Intn(0xFFFFFF+1)) //nolint:gosec
}

// parseCSVUserName splits a "Firstname Lastname" display name into (name,
// surname). If there is only one token it is used as the name.
func parseCSVUserName(full string) (name, surname string) {
	full = strings.TrimSpace(full)
	idx := strings.IndexByte(full, ' ')
	if idx < 0 {
		return full, ""
	}
	return full[:idx], strings.TrimSpace(full[idx+1:])
}

// ImportCSV accepts a "detailed report" CSV (sent as the raw request body)
// and creates matching time entries in the organization. This column format
// originated with Clockify but is also this app's own detailed report export
// format (see report.DetailedCSV), so the same endpoint doubles as the way
// to migrate data from another cwclock instance.
// For each row the handler will:
//   - find or create the client (by name, with a random colour)
//   - find or create the project under that client (by name, with a random colour)
//   - find or create the user (by email, created as disabled)
//   - skip the row if a matching entry (same user/project/day/start/end) already exists
//   - otherwise create the time entry
//
// The response body is {"created": N, "skipped": M}.
func (h *ImportHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, maxImportBodySize)
	reader := csv.NewReader(r.Body)
	reader.FieldsPerRecord = -1 // tolerate variable field counts
	reader.LazyQuotes = true

	// Read and validate the header row.
	header, err := reader.Read()
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read CSV header", CodeInvalidRequestBody)
		return
	}
	colIdx := make(map[string]int, len(header))
	for i, col := range header {
		colIdx[col] = i
	}
	required := []string{"Project", "Client", "Description", "User", "Email",
		"Start Date", "Start Time", "End Date", "End Time"}
	for _, col := range required {
		if _, ok := colIdx[col]; !ok {
			writeError(w, http.StatusBadRequest,
				fmt.Sprintf("Missing required CSV column: %s", col), CodeInvalidRequestBody)
			return
		}
	}

	created, skipped := 0, 0

	for {
		rec, err := reader.Read()
		if err != nil {
			break // io.EOF or unrecoverable parse error
		}
		minLen := colIdx["End Time"] + 1
		if len(rec) < minLen {
			skipped++
			continue
		}

		projectName := strings.TrimSpace(rec[colIdx["Project"]])
		clientName := strings.TrimSpace(rec[colIdx["Client"]])
		description := strings.TrimSpace(rec[colIdx["Description"]])
		userFullName := strings.TrimSpace(rec[colIdx["User"]])
		email := strings.TrimSpace(rec[colIdx["Email"]])

		if projectName == "" || clientName == "" || email == "" {
			skipped++
			continue
		}

		day, err := parseCSVDate(rec[colIdx["Start Date"]])
		if err != nil {
			skipped++
			continue
		}
		startStr, err := parseCSVTime(rec[colIdx["Start Time"]])
		if err != nil {
			skipped++
			continue
		}
		endStr, err := parseCSVTime(rec[colIdx["End Time"]])
		if err != nil {
			skipped++
			continue
		}

		// Get or create client.
		client, err := h.clients.FindByName(ctx, orgID, clientName)
		if errors.Is(err, store.ErrNotFound) {
			client, err = h.clients.Create(ctx, orgID, store.ClientFields{Name: clientName})
		}
		if err != nil {
			skipped++
			continue
		}

		// Get or create project under the client.
		project, err := h.projects.FindByName(ctx, orgID, client.ID, projectName)
		if errors.Is(err, store.ErrNotFound) {
			project, err = h.projects.Create(ctx, orgID, client.ID, projectName, randomColor())
		}
		if err != nil {
			skipped++
			continue
		}

		// Get or create user (created as disabled, no password).
		user, err := h.users.FindByEmail(ctx, email)
		if errors.Is(err, store.ErrNotFound) {
			name, surname := parseCSVUserName(userFullName)
			user, err = h.users.CreateDisabled(ctx, email, name, surname)
		}
		if err != nil {
			skipped++
			continue
		}

		// Skip if a matching entry already exists.
		exists, err := h.timeEntries.ExistsByKey(ctx, orgID, user.ID, project.ID, day, startStr, endStr)
		if err != nil || exists {
			skipped++
			continue
		}

		text := description
		if text == "" {
			text = projectName
		}

		_, err = h.timeEntries.Create(ctx, orgID, client.ID, project.ID, user.ID, store.TimeEntryFields{
			Text:  text,
			Day:   day,
			Start: &startStr,
			End:   &endStr,
		})
		if err != nil {
			skipped++
			continue
		}
		created++
	}

	writeJSON(w, http.StatusOK, importResult{Created: created, Skipped: skipped})
}
