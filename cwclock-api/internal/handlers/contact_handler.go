package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"cwclock-api/internal/contact"
	"cwclock-api/internal/utils"
)

type ContactHandler struct {
	contact *contact.Client
}

func NewContactHandler(contact *contact.Client) *ContactHandler {
	return &ContactHandler{contact: contact}
}

// contactPayload is the JSON body accepted by POST /v1/contact - name and
// firstname are optional (ai-instruct-54), everything else is required.
type contactPayload struct {
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Name      string `json:"name"`
	Firstname string `json:"firstname"`
}

// Create submits the contact form to CWCloud's contact-request API. Answers
// 405 (matching ai-instruct-54's spec exactly) when CWCLOUD_CONTACT_FORM_ID
// isn't configured, rather than the 500 an unconfigured downstream
// dependency would normally get - this is a deploy-time configuration gap,
// not a request-time server error.
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.contact.Configured() {
		writeError(w, http.StatusMethodNotAllowed, "The contact form is not configured", CodeContactFormNotConfigured)
		return
	}

	var p contactPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil ||
		utils.IsBlank(p.Email) || utils.IsBlank(p.Subject) || utils.IsBlank(p.Message) {
		writeError(w, http.StatusBadRequest, "Please add email, subject and message fields", CodeAllFieldsRequired)
		return
	}
	if !utils.IsValidEmail(p.Email) {
		writeError(w, http.StatusBadRequest, "Please add a valid email", CodeInvalidEmail)
		return
	}

	err := h.contact.Send(r.Context(), contact.Submission{
		Email: p.Email, Subject: p.Subject, Message: p.Message,
		Name: p.Name, Firstname: p.Firstname,
	})
	if err != nil {
		slog.Error("failed to submit contact form", "error", err)
		writeError(w, http.StatusBadGateway, "Failed to send your message, please try again later", CodeContactFormFailed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Your message has been sent."})
}
