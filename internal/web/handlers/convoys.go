package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

// CreateConvoyRequest represents a request to create a new convoy.
type CreateConvoyRequest struct {
	Name     string   `json:"name"`
	IssueIDs []string `json:"issue_ids"`
}

// CreateConvoyResponse represents the response after creating a convoy.
type CreateConvoyResponse struct {
	ID string `json:"id"`
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

// Create creates a new convoy.
func (h *ConvoysHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateConvoyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body: "+err.Error())
		return
	}

	// Validate input
	if req.Name == "" {
		api.BadRequest(w, "name is required")
		return
	}

	if len(req.IssueIDs) == 0 {
		api.BadRequest(w, "at least one issue ID is required")
		return
	}

	// Generate convoy ID
	convoyID := generateConvoyID()

	// TODO: Persist convoy to backend
	// This will be fully implemented in Phase 2/3

	resp := CreateConvoyResponse{
		ID: convoyID,
	}
	api.WriteJSON(w, resp)
}

// generateConvoyID generates a unique convoy ID.
func generateConvoyID() string {
	// Generate 6 random bytes and convert to hex for a 12-character ID
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
