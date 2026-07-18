package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"cwclock-api/internal/store"
)

type errorBody struct {
	Message  string `json:"message"`
	I18nCode string `json:"i18n_code,omitempty"`
}

// I18n codes for API errors. The frontend looks these up in its own
// translation dictionaries, falling back to the English Message when a code
// is absent or unrecognized (e.g. an older client or a code it doesn't know).
const (
	CodeInternal                    = "errors.internal"
	CodeNotFound                    = "errors.notFound"
	CodeDuplicateEmail              = "errors.duplicateEmail"
	CodeUserNotFound                = "errors.userNotFound"
	CodeInvalidCredentials          = "errors.invalidCredentials"
	CodeNameRequired                = "errors.nameRequired"
	CodeCountryRequired             = "errors.countryRequired"
	CodeAllFieldsRequired           = "errors.allFieldsRequired"
	CodeInvalidUserData             = "errors.invalidUserData"
	CodeInvalidRequestBody          = "errors.invalidRequestBody"
	CodeInvalidMemberInvite         = "errors.invalidMemberInvite"
	CodeNoUserWithEmail             = "errors.noUserWithEmail"
	CodeInvalidRole                 = "errors.invalidRole"
	CodeInvalidDailyRate            = "errors.invalidDailyRate"
	CodeInvalidEmail                = "errors.invalidEmail"
	CodeInvalidCurrency             = "errors.invalidCurrency"
	CodeCantDeleteOwnAccount        = "errors.cantDeleteOwnAccount"
	CodeInvalidAdminUserEdit        = "errors.invalidAdminUserEdit"
	CodeTimeEntryFields             = "errors.timeEntryFields"
	CodeTimeEntryStartEnd           = "errors.timeEntryStartEnd"
	CodeOwnTimeEntriesOnly          = "errors.ownTimeEntriesOnly"
	CodeReassignForbidden           = "errors.reassignForbidden"
	CodeNameAndSurnameRequired      = "errors.nameAndSurnameRequired"
	CodePasswordsMismatch           = "errors.passwordsMismatch"
	CodeApiKeyDescription           = "errors.apiKeyDescription"
	CodeInvalidExpiration           = "errors.invalidExpiration"
	CodeImageTooLarge               = "errors.imageTooLarge"
	CodeInvalidInvoiceStatus        = "errors.invalidInvoiceStatus"
	CodeInvalidInvoiceRequest       = "errors.invalidInvoiceRequest"
	CodeExportLimitExceeded         = "errors.exportLimitExceeded"
	CodeCantRemoveOwner             = "errors.cantRemoveOwner"
	CodeMustOwnTargetOrg            = "errors.mustOwnTargetOrg"
	CodeInvalidClientForOrg         = "errors.invalidClientForOrg"
	CodeInvalidCountry              = "errors.invalidCountry"
	CodeInvalidExternalConnection   = "errors.invalidExternalConnection"
	CodeDuplicateExternalConnection = "errors.duplicateExternalConnection"
	CodeInvoiceNumberExists         = "errors.invoiceNumberExists"
	CodeInvalidToken                = "errors.invalidToken"
	CodeNoInvoiceRecipient          = "errors.noInvoiceRecipient"
	CodePasswordTooShort            = "errors.passwordTooShort"
	CodePasswordNoUpper             = "errors.passwordNoUpper"
	CodePasswordNoLower             = "errors.passwordNoLower"
	CodePasswordNoSymbol            = "errors.passwordNoSymbol"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func writeError(w http.ResponseWriter, status int, message string, i18nCode string) {
	writeJSON(w, status, errorBody{Message: message, I18nCode: i18nCode})
}

// passwordPolicyMessages pairs each i18n code utils.IsPasswordValid can
// return with its English fallback message.
var passwordPolicyMessages = map[string]string{
	CodePasswordTooShort: "Password must be at least 8 characters long",
	CodePasswordNoUpper:  "Password must contain an uppercase letter",
	CodePasswordNoLower:  "Password must contain a lowercase letter",
	CodePasswordNoSymbol: "Password must contain a special character",
}

// writeInvalidPassword rejects a request whose new password failed
// utils.IsPasswordValid, given the i18n code it returned.
func writeInvalidPassword(w http.ResponseWriter, code string) {
	writeError(w, http.StatusBadRequest, passwordPolicyMessages[code], code)
}

// writeStoreError maps a store error to its HTTP status: 404 when the
// resource doesn't exist, 400 for a duplicate email, 500 for anything else.
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Resource not found", CodeNotFound)
		return
	}
	if errors.Is(err, store.ErrDuplicateEmail) {
		writeError(w, http.StatusBadRequest, "Email already in use", CodeDuplicateEmail)
		return
	}
	if errors.Is(err, store.ErrExportLimitExceeded) {
		writeError(w, http.StatusBadRequest, "Too many entries for a single export", CodeExportLimitExceeded)
		return
	}
	if errors.Is(err, store.ErrCannotRemoveOwner) {
		writeError(w, http.StatusBadRequest, "The organization owner cannot be removed", CodeCantRemoveOwner)
		return
	}
	if errors.Is(err, store.ErrInvoiceNumberExists) {
		writeError(w, http.StatusConflict, "An invoice with this number already exists", CodeInvoiceNumberExists)
		return
	}
	slog.Error("unhandled store error", "error", err)
	writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
}
