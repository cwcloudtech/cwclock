package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// A user's threads with the assistant: listing them, reading one with its exchanges,
// and clearing a selection or the whole history. Stored at all because an answer is
// worth keeping, and because somebody lending a key out is entitled to see what it
// was spent on.

// ListConversations is the caller's threads, newest activity first, without contents.
func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	conversations, err := h.connections.ListConversations(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, conversations)
}

// GetConversation is one thread with its exchanges, for its owner. Scoped in the
// query, so a thread that is not the caller's is indistinguishable from one that does
// not exist (404).
func (h *AIHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	conversation, err := h.connections.FindConversation(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, conversation)
}

type deleteConversationsPayload struct {
	IDs []string `json:"ids"`
	// All empties the history. Asked for explicitly rather than inferred from an
	// empty id list, so "delete these" with the list accidentally empty does not mean
	// "delete everything".
	All bool `json:"all"`
}

// DeleteConversations clears a selection, or the whole history.
func (h *AIHandler) DeleteConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := aiUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", CodeInvalidCredentials)
		return
	}
	var p deleteConversationsPayload
	if !decodeAIBody(w, r, &p) {
		return
	}
	if !p.All && len(p.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "Pick what to delete", CodeInvalidRequestBody)
		return
	}

	removed, err := h.connections.DeleteConversations(r.Context(), userID, p.IDs, p.All)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": removed})
}
