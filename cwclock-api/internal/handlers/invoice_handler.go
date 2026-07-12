package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/externalconn"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/report"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type InvoiceHandler struct {
	invoices *store.InvoiceStore
	orgs     *store.OrgStore
	clients  *store.ClientStore
	projects *store.ProjectStore
	entries  *store.TimeEntryStore
	users    *store.UserStore
	maxSize  int
}

func NewInvoiceHandler(
	invoices *store.InvoiceStore,
	orgs *store.OrgStore,
	clients *store.ClientStore,
	projects *store.ProjectStore,
	entries *store.TimeEntryStore,
	users *store.UserStore,
	maxSize int,
) *InvoiceHandler {
	return &InvoiceHandler{invoices: invoices, orgs: orgs, clients: clients, projects: projects, entries: entries, users: users, maxSize: maxSize}
}

// invoiceRequest is the JSON body accepted by the preview/generate
// endpoints: one client (required) and a date range (required), matching
// exportRequest's date shape so the frontend can reuse the same
// dateRangeStart/dateRangeEnd payload convention as reports.
type invoiceRequest struct {
	ClientID       string   `json:"clientId"`
	DateRangeStart string   `json:"dateRangeStart"`
	DateRangeEnd   string   `json:"dateRangeEnd"`
	ProjectIDs     []string `json:"projectIds"`
}

func decodeInvoiceRequest(w http.ResponseWriter, r *http.Request) (invoiceRequest, bool) {
	var req invoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		utils.IsBlank(req.ClientID) || utils.IsBlank(req.DateRangeStart) || utils.IsBlank(req.DateRangeEnd) {
		writeError(w, http.StatusBadRequest, "Please add a clientId, dateRangeStart and dateRangeEnd", CodeInvalidInvoiceRequest)
		return invoiceRequest{}, false
	}
	return req, true
}

// invoiceContext is everything computed from an invoiceRequest that both
// Preview and Generate need before they can render the PDF.
type invoiceContext struct {
	org      models.Organization
	client   models.Client
	owner    models.User
	items    []report.InvoiceLineItem
	totalHT  float64
	totalVAT float64
	totalTTC float64
	startDay string
	endDay   string
}

func (h *InvoiceHandler) load(ctx context.Context, orgID string, req invoiceRequest) (invoiceContext, error) {
	org, err := h.orgs.FindByID(ctx, orgID)
	if err != nil {
		return invoiceContext{}, err
	}

	client, err := h.clients.FindByID(ctx, req.ClientID)
	if err != nil {
		return invoiceContext{}, err
	}
	if client.OrganizationID != orgID {
		return invoiceContext{}, store.ErrNotFound
	}

	owner, err := h.users.FindByID(ctx, org.OwnerID)
	if err != nil {
		return invoiceContext{}, err
	}

	projectsList, err := h.projects.List(ctx, orgID, req.ClientID)
	if err != nil {
		return invoiceContext{}, err
	}
	projectsByID := make(map[string]models.Project, len(projectsList))
	for _, p := range projectsList {
		projectsByID[p.ID] = p
	}

	start, end := dayPart(req.DateRangeStart), dayPart(req.DateRangeEnd)
	filter := store.ReportFilter{
		Start: start, End: end,
		ClientIDs:  []string{req.ClientID},
		ProjectIDs: req.ProjectIDs,
	}

	count, err := h.entries.CountForReport(ctx, orgID, filter)
	if err != nil {
		return invoiceContext{}, err
	}
	if count > h.maxSize {
		return invoiceContext{}, store.ErrExportLimitExceeded
	}

	entries, err := h.entries.ListForReport(ctx, orgID, filter)
	if err != nil {
		return invoiceContext{}, err
	}

	items := report.BuildInvoiceLineItems(entries, projectsByID, client)
	totalHT, totalVAT, totalTTC := report.InvoiceVATTotals(items, client)

	return invoiceContext{
		org: org, client: client, owner: owner,
		items: items, totalHT: totalHT, totalVAT: totalVAT, totalTTC: totalTTC,
		startDay: start, endDay: end,
	}, nil
}

// Preview renders the invoice PDF without saving anything, so the caller
// can check it before committing to a real invoice number. Its displayed
// number is only a best-effort peek (see InvoiceStore.PeekNextNumber) -
// generating for real may land on a different one if another invoice for
// this client is created in between.
func (h *InvoiceHandler) Preview(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	req, ok := decodeInvoiceRequest(w, r)
	if !ok {
		return
	}

	ic, err := h.load(r.Context(), orgID, req)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	number, err := h.invoices.PeekNextNumber(r.Context(), orgID, ic.client.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	pdf, err := report.RenderInvoicePDF(ic.org, ic.client, ic.owner, number, ic.items, ic.totalHT, ic.totalVAT, ic.totalTTC, ic.startDay, ic.endDay)
	writeExportFile(w, "application/pdf", number+".pdf", pdf, err)
}

// Generate renders the invoice PDF, saves it (with its authoritatively
// allocated invoice number - see InvoiceStore.Create) in the invoices
// table, and streams the same PDF back as a download.
func (h *InvoiceHandler) Generate(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	req, ok := decodeInvoiceRequest(w, r)
	if !ok {
		return
	}

	ic, err := h.load(r.Context(), orgID, req)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	var pdf []byte
	var renderErr error
	inv, err := h.invoices.Create(r.Context(), orgID, req.ClientID, ic.client.Name, func(number string) (store.InvoiceFields, error) {
		pdf, renderErr = report.RenderInvoicePDF(ic.org, ic.client, ic.owner, number, ic.items, ic.totalHT, ic.totalVAT, ic.totalTTC, ic.startDay, ic.endDay)
		if renderErr != nil {
			return store.InvoiceFields{}, renderErr
		}
		return store.InvoiceFields{
			Status:            string(models.InvoiceStatusUnpaid),
			TotalHT:           ic.totalHT,
			TotalVAT:          ic.totalVAT,
			TotalTTC:          ic.totalTTC,
			PDF:               pdf,
			SelectedBeginDate: ic.startDay,
			SelectedEndDate:   ic.endDay,
		}, nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	externalconn.SyncUpload(r.Context(), ic.org.ExternalConnections, externalconn.YearFolder(inv.CreatedAt), externalconn.MonthCandidates(inv.CreatedAt), inv.Number+".pdf", pdf)

	writeExportFile(w, "application/pdf", inv.Number+".pdf", pdf, nil)
}

// List returns an organization's invoices for one client within a date
// range (all three query params required), most recent first.
func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	clientID := r.URL.Query().Get("clientId")
	start := dayPart(r.URL.Query().Get("start"))
	end := dayPart(r.URL.Query().Get("end"))
	if utils.IsBlank(clientID) || utils.IsBlank(start) || utils.IsBlank(end) {
		writeError(w, http.StatusBadRequest, "Please add clientId, start and end query params", CodeInvalidInvoiceRequest)
		return
	}

	invoices, err := h.invoices.List(r.Context(), orgID, clientID, start, end)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, invoices)
}

// DownloadPDF streams a previously generated invoice's stored PDF.
func (h *InvoiceHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	invoiceID := chi.URLParam(r, "invoiceId")

	inv, err := h.invoices.FindByID(r.Context(), invoiceID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if inv.OrganizationID != orgID {
		writeError(w, http.StatusNotFound, "Resource not found", CodeNotFound)
		return
	}

	pdf, number, err := h.invoices.GetPDF(r.Context(), invoiceID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeExportFile(w, "application/pdf", number+".pdf", pdf, nil)
}

// Reupload pushes an already-generated invoice's stored PDF to every one of
// its organization's external connections again (e.g. after fixing a
// connection's credentials), replacing the file previously written there.
// It always responds 200 regardless of per-connection outcomes - those are
// logged server-side (see externalconn.SyncUpload) - since the invoice
// itself is unchanged either way.
func (h *InvoiceHandler) Reupload(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	invoiceID := chi.URLParam(r, "invoiceId")

	inv, err := h.invoices.FindByID(r.Context(), invoiceID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if inv.OrganizationID != orgID {
		writeError(w, http.StatusNotFound, "Resource not found", CodeNotFound)
		return
	}

	org, err := h.orgs.FindByID(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	pdf, number, err := h.invoices.GetPDF(r.Context(), invoiceID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	externalconn.SyncUpload(r.Context(), org.ExternalConnections, externalconn.YearFolder(inv.CreatedAt), externalconn.MonthCandidates(inv.CreatedAt), number+".pdf", pdf)
	writeJSON(w, http.StatusOK, map[string]string{"id": invoiceID})
}

type updateInvoiceStatusPayload struct {
	Status string `json:"status"`
}

// UpdateStatus lets an admin/owner move an invoice through its payment
// lifecycle (unpaid/paid/canceled/refunded).
func (h *InvoiceHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	invoiceID := chi.URLParam(r, "invoiceId")

	var p updateInvoiceStatusPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !models.IsValidInvoiceStatus(p.Status) {
		writeError(w, http.StatusBadRequest, "Please add a valid status", CodeInvalidInvoiceStatus)
		return
	}

	inv, err := h.invoices.FindByID(r.Context(), invoiceID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if inv.OrganizationID != orgID {
		writeError(w, http.StatusNotFound, "Resource not found", CodeNotFound)
		return
	}

	updated, err := h.invoices.UpdateStatus(r.Context(), invoiceID, p.Status)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete removes an invoice (and its stored PDF) entirely.
func (h *InvoiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	invoiceID := chi.URLParam(r, "invoiceId")

	inv, err := h.invoices.FindByID(r.Context(), invoiceID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if inv.OrganizationID != orgID {
		writeError(w, http.StatusNotFound, "Resource not found", CodeNotFound)
		return
	}

	if err := h.invoices.Delete(r.Context(), invoiceID); err != nil {
		writeStoreError(w, err)
		return
	}

	if org, err := h.orgs.FindByID(r.Context(), orgID); err == nil {
		externalconn.SyncDelete(r.Context(), org.ExternalConnections, externalconn.YearFolder(inv.CreatedAt), externalconn.MonthCandidates(inv.CreatedAt), inv.Number+".pdf")
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": invoiceID})
}
