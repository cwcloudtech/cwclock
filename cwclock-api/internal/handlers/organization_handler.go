package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"cwclock-api/internal/externalconn"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type OrganizationHandler struct {
	orgs         *store.OrgStore
	users        *store.UserStore
	countries    *store.CountryStore
	currencies   *store.CurrencyStore
	maxImageSize int64
}

func NewOrganizationHandler(orgs *store.OrgStore, users *store.UserStore, countries *store.CountryStore, currencies *store.CurrencyStore, maxImageSize int64) *OrganizationHandler {
	return &OrganizationHandler{orgs: orgs, users: users, countries: countries, currencies: currencies, maxImageSize: maxImageSize}
}

type organizationPayload struct {
	Name                 string                      `json:"name"`
	AccountingEmail      string                      `json:"accountingEmail"`
	Address              string                      `json:"address"`
	PostalCode           string                      `json:"postalCode"`
	City                 string                      `json:"city"`
	Country              string                      `json:"country"`
	VATNumber            string                      `json:"vatNumber"`
	SIREN                string                      `json:"siren"`
	SIRET                string                      `json:"siret"`
	NAF                  string                      `json:"naf"`
	MF                   string                      `json:"mf"`
	IdentificationNumber string                      `json:"identificationNumber"`
	Picture              string                      `json:"picture"`
	PictureX             float64                     `json:"pictureX"`
	PictureY             float64                     `json:"pictureY"`
	Stamp                string                      `json:"stamp"`
	StampX               float64                     `json:"stampX"`
	StampY               float64                     `json:"stampY"`
	Currency             string                      `json:"currency"`
	ExternalConnections  []models.ExternalConnection `json:"externalConnections"`
}

// nameValid and Country's own blank check (see Create/Update) are kept
// separate rather than one combined "valid" bool, so a blank name and a
// blank country produce their own specific error message instead of both
// being reported as "Please fill in the Name field" (ai-instruct-37).
func (p organizationPayload) nameValid() bool {
	return utils.IsNotBlank(p.Name)
}

// validAccountingEmail lets the field stay blank (it's optional) but
// requires a plausible address when set.
func (p organizationPayload) validAccountingEmail() bool {
	return utils.IsBlank(p.AccountingEmail) || utils.IsValidEmail(p.AccountingEmail)
}

func (h *OrganizationHandler) validCountry(ctx context.Context, p organizationPayload) (bool, error) {
	return h.countries.Exists(ctx, p.Country)
}

func (h *OrganizationHandler) validCurrency(ctx context.Context, p organizationPayload) (bool, error) {
	if utils.IsBlank(p.Currency) {
		return true, nil
	}
	return h.currencies.Exists(ctx, p.Currency)
}

func (p organizationPayload) imageTooLarge(maxImageSize int64) bool {
	return utils.ImageSizeExceeds(p.Picture, maxImageSize) || utils.ImageSizeExceeds(p.Stamp, maxImageSize)
}

// validateExternalConnections enforces "every field is mandatory" per
// connection type (ai-instruct-39) and assigns a fresh ID to any connection
// the frontend submitted without one (newly added rows), so callers never
// have to invent ids client-side.
func validateExternalConnections(conns []models.ExternalConnection) error {
	for i, c := range conns {
		switch c.Type {
		case models.ExternalConnectionS3:
			if utils.IsBlank(c.Endpoint) || utils.IsBlank(c.BucketName) || utils.IsBlank(c.Region) ||
				utils.IsBlank(c.AccessKey) || utils.IsBlank(c.SecretKey) {
				return fmt.Errorf("s3 external connection requires endpoint, bucketName, region, accessKey and secretKey")
			}
		case models.ExternalConnectionGoogleDrive:
			if utils.IsBlank(c.ServiceAccountBase64) || utils.IsBlank(c.FolderID) {
				return fmt.Errorf("google_drive external connection requires serviceAccountBase64 and folderId")
			}
			if _, err := externalconn.DecodeServiceAccount(c.ServiceAccountBase64); err != nil {
				return fmt.Errorf("google_drive external connection has an invalid service account: %w", err)
			}
		default:
			return fmt.Errorf("unknown external connection type %q", c.Type)
		}
		if utils.IsBlank(c.ID) {
			conns[i].ID = uuid.NewString()
		}
	}
	return nil
}

func (p organizationPayload) toFields() store.OrganizationFields {
	return store.OrganizationFields{
		Name:                 p.Name,
		AccountingEmail:      p.AccountingEmail,
		Address:              p.Address,
		PostalCode:           p.PostalCode,
		City:                 p.City,
		Country:              p.Country,
		VATNumber:            p.VATNumber,
		SIREN:                p.SIREN,
		SIRET:                p.SIRET,
		NAF:                  p.NAF,
		MF:                   p.MF,
		IdentificationNumber: p.IdentificationNumber,
		Picture:              p.Picture,
		PictureX:             p.PictureX,
		PictureY:             p.PictureY,
		Stamp:                p.Stamp,
		StampX:               p.StampX,
		StampY:               p.StampY,
		Currency:             p.Currency,
		ExternalConnections:  p.ExternalConnections,
	}
}

// redactExternalConnections clears the credential-bearing fields of each
// connection (S3 access/secret key, the Google Drive service account) so
// they're never echoed back in an API response - see ai-instruct-41. This
// only affects what gets serialized to the client; the real values are
// still what's persisted and used server-side to talk to S3/Drive.
func redactExternalConnections(conns []models.ExternalConnection) []models.ExternalConnection {
	redacted := make([]models.ExternalConnection, len(conns))
	for i, c := range conns {
		c.AccessKey = ""
		c.SecretKey = ""
		c.ServiceAccountBase64 = ""
		redacted[i] = c
	}
	return redacted
}

func redactOrg(org models.Organization) models.Organization {
	org.ExternalConnections = redactExternalConnections(org.ExternalConnections)
	return org
}

func redactOrgs(orgs []models.Organization) []models.Organization {
	redacted := make([]models.Organization, len(orgs))
	for i, o := range orgs {
		redacted[i] = redactOrg(o)
	}
	return redacted
}

func redactOrgsWithOwner(orgs []models.OrganizationWithOwner) []models.OrganizationWithOwner {
	redacted := make([]models.OrganizationWithOwner, len(orgs))
	for i, o := range orgs {
		o.Organization = redactOrg(o.Organization)
		redacted[i] = o
	}
	return redacted
}

func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p organizationPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if utils.IsBlank(p.Country) {
		writeError(w, http.StatusBadRequest, "Please select a country", CodeCountryRequired)
		return
	}
	if !p.validAccountingEmail() {
		writeError(w, http.StatusBadRequest, "Please add a valid accounting email", CodeInvalidEmail)
		return
	}
	if ok, err := h.validCountry(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported country code", CodeInvalidCountry)
		return
	}
	if ok, err := h.validCurrency(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported currency code", CodeInvalidCurrency)
		return
	}
	if p.imageTooLarge(h.maxImageSize) {
		writeError(w, http.StatusBadRequest, "Image is too large", CodeImageTooLarge)
		return
	}
	if err := validateExternalConnections(p.ExternalConnections); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), CodeInvalidExternalConnection)
		return
	}

	org, err := h.orgs.Create(r.Context(), userID, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, redactOrg(org))
}

func (h *OrganizationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	orgs, err := h.orgs.ListForUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactOrgs(orgs))
}

// AdminList returns every organization, regardless of membership, for the
// superuser's organization-management screen.
func (h *OrganizationHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.orgs.ListAllWithOwner(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactOrgsWithOwner(orgs))
}

func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	org, err := h.orgs.FindByID(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactOrg(org))
}

// PublicLogo serves an organization's avatar as a real image over HTTP (see
// AssetsLogo, which this falls back to when the org has no picture set or
// it isn't a decodable, supported image - so a stale/invalid logo still
// shows something in an email rather than a broken image icon). No auth:
// it's a public route, since it needs to be fetchable by a mail client with
// no credentials - the same reasoning as AssetsLogo, and organization ids
// aren't guessable.
func (h *OrganizationHandler) PublicLogo(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	org, err := h.orgs.FindByID(r.Context(), orgID)
	if err == nil {
		if data, mimeType, ok := utils.DecodeImage(org.Picture); ok {
			w.Header().Set("Content-Type", mimeType)
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
	}
	AssetsLogo(w, r)
}

func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var p organizationPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if utils.IsBlank(p.Country) {
		writeError(w, http.StatusBadRequest, "Please select a country", CodeCountryRequired)
		return
	}
	if !p.validAccountingEmail() {
		writeError(w, http.StatusBadRequest, "Please add a valid accounting email", CodeInvalidEmail)
		return
	}
	if ok, err := h.validCountry(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported country code", CodeInvalidCountry)
		return
	}
	if ok, err := h.validCurrency(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported currency code", CodeInvalidCurrency)
		return
	}
	if p.imageTooLarge(h.maxImageSize) {
		writeError(w, http.StatusBadRequest, "Image is too large", CodeImageTooLarge)
		return
	}

	// External connections are managed exclusively through the dedicated
	// Add/RemoveExternalConnection endpoints now (ai-instruct-40/41): the
	// whole-org save ignores whatever the client sent for this field and
	// keeps the organization's current connections as-is. Accepting it here
	// would let a client round-trip the redacted list it was given (see
	// redactExternalConnections) back into storage, wiping out every
	// connection's real access/secret key or service account.
	existing, err := h.orgs.FindByID(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	fields := p.toFields()
	fields.ExternalConnections = existing.ExternalConnections

	org, err := h.orgs.Update(r.Context(), orgID, fields)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactOrg(org))
}

// isDuplicateExternalConnection reports whether conn points at the same
// destination as one of existing: same endpoint/bucket/region for S3 (the
// credentials may legitimately differ, but it's still the same bucket), or
// same folder for Google Drive.
func isDuplicateExternalConnection(existing []models.ExternalConnection, conn models.ExternalConnection) bool {
	for _, e := range existing {
		if e.Type != conn.Type {
			continue
		}
		switch conn.Type {
		case models.ExternalConnectionS3:
			if e.Endpoint == conn.Endpoint && e.BucketName == conn.BucketName && e.Region == conn.Region {
				return true
			}
		case models.ExternalConnectionGoogleDrive:
			if e.FolderID == conn.FolderID {
				return true
			}
		}
	}
	return false
}

// AddExternalConnection appends a single external connection to the
// organization and persists it immediately (ai-instruct-40: "add an
// external connection should automatically save the organization"), instead
// of requiring the whole organization form to be submitted. Adding one that
// already points at the same destination (same S3 bucket, or same Drive
// folder) is rejected with 409 rather than silently creating a duplicate.
func (h *OrganizationHandler) AddExternalConnection(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var conn models.ExternalConnection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		writeError(w, http.StatusBadRequest, "Please provide a valid external connection", CodeInvalidExternalConnection)
		return
	}
	conns := []models.ExternalConnection{conn}
	if err := validateExternalConnections(conns); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), CodeInvalidExternalConnection)
		return
	}

	org, err := h.orgs.FindByID(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if isDuplicateExternalConnection(org.ExternalConnections, conns[0]) {
		writeError(w, http.StatusConflict, "This external connection already exists", CodeDuplicateExternalConnection)
		return
	}

	updated, err := h.orgs.AddExternalConnection(r.Context(), orgID, conns[0])
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactOrg(updated))
}

// RemoveExternalConnection removes a single external connection from the
// organization and persists it immediately, mirroring AddExternalConnection.
func (h *OrganizationHandler) RemoveExternalConnection(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	connID := chi.URLParam(r, "connectionId")

	org, err := h.orgs.RemoveExternalConnection(r.Context(), orgID, connID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactOrg(org))
}

func (h *OrganizationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	if err := h.orgs.Delete(r.Context(), orgID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": orgID})
}

type memberPayload struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func validRole(role string) bool {
	switch models.Role(role) {
	case models.RoleAdmin, models.RoleMember, models.RoleReader:
		return true
	default:
		return false
	}
}

// redactRates hides the daily rate from the response unless the caller is
// an admin or the owner, per "the daily rate must appear only for the
// admins and owner of the organization".
func redactRates(r *http.Request, members []models.Member) []models.Member {
	role, _ := middleware.OrgRoleFromContext(r.Context())
	if role == models.RoleAdmin || role == models.RoleOwner {
		return members
	}
	redacted := make([]models.Member, len(members))
	for i, m := range members {
		m.DailyRate = nil
		redacted[i] = m
	}
	return redacted
}

func (h *OrganizationHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	members, err := h.orgs.ListMembers(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactRates(r, members))
}

func (h *OrganizationHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var p memberPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Email) || !validRole(p.Role) {
		writeError(w, http.StatusBadRequest, "Please add a valid email and role (admin, member or reader)", CodeInvalidMemberInvite)
		return
	}

	user, err := h.users.FindByEmail(r.Context(), p.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "No user with this email", CodeNoUserWithEmail)
			return
		}
		writeStoreError(w, err)
		return
	}

	member, err := h.orgs.AddMember(r.Context(), orgID, user.ID, models.Role(p.Role))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

func (h *OrganizationHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	var p memberPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !validRole(p.Role) {
		writeError(w, http.StatusBadRequest, "Please add a valid role (admin, member or reader)", CodeInvalidRole)
		return
	}

	member, err := h.orgs.UpdateMemberRole(r.Context(), orgID, userID, models.Role(p.Role))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, member)
}

type memberRatePayload struct {
	DailyRate float64 `json:"dailyRate"`
}

func (h *OrganizationHandler) SetMemberRate(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	var p memberRatePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.DailyRate <= 0 {
		writeError(w, http.StatusBadRequest, "Please add a valid dailyRate", CodeInvalidDailyRate)
		return
	}

	member, err := h.orgs.SetMemberRate(r.Context(), orgID, userID, p.DailyRate)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, member)
}

func (h *OrganizationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	if err := h.orgs.RemoveMember(r.Context(), orgID, userID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"userId": userID})
}

type transferOwnershipPayload struct {
	Email string `json:"email"`
}

func (h *OrganizationHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	previousOwnerID, _ := middleware.UserIDFromContext(r.Context())

	var p transferOwnershipPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Email) {
		writeError(w, http.StatusBadRequest, "Please add a valid email", CodeInvalidEmail)
		return
	}

	newOwner, err := h.users.FindByEmail(r.Context(), p.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "No user with this email", CodeNoUserWithEmail)
			return
		}
		writeStoreError(w, err)
		return
	}

	org, err := h.orgs.TransferOwnership(r.Context(), orgID, previousOwnerID, newOwner.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactOrg(org))
}
