package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/web/api"
)

// ConvoysHandler handles convoy-related API requests.
type ConvoysHandler struct {
	townRoot string
}

// NewConvoysHandler creates a new convoys handler.
func NewConvoysHandler(townRoot string) *ConvoysHandler {
	return &ConvoysHandler{townRoot: townRoot}
}

// ConvoyInfo represents a convoy for API responses.
type ConvoyInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"` // in_progress, landed
	TrackedIDs  []string `json:"tracked_ids"`
	Progress    float64  `json:"progress"` // 0.0 - 1.0
	Completed   int      `json:"completed"`
	Total       int      `json:"total"`
	CreatedAt   string   `json:"created_at,omitempty"`
	CompletedAt string   `json:"completed_at,omitempty"`
}

// List returns all convoys.
func (h *ConvoysHandler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement convoy listing
	// This will be fully implemented in Phase 3
	api.WriteJSON(w, []ConvoyInfo{})
}

// Get returns a single convoy by ID.
func (h *ConvoysHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Implement convoy lookup
	// This will be fully implemented in Phase 3
	api.NotFound(w, "convoy not found: "+id)
}
