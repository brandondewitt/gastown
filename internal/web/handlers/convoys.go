package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

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
	Status      string   `json:"status"` // open, closed
	TrackedIDs  []string `json:"tracked_ids"`
	Progress    string   `json:"progress"` // e.g., "2/5"
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

// List returns all open convoys.
func (h *ConvoysHandler) List(w http.ResponseWriter, r *http.Request) {
	townBeads := filepath.Join(h.townRoot, ".beads")

	// List all open convoy-type issues
	listArgs := []string{"list", "--type=convoy", "--status=open", "--json"}
	listCmd := exec.Command("bd", listArgs...)
	listCmd.Dir = townBeads

	var stdout bytes.Buffer
	listCmd.Stdout = &stdout

	if err := listCmd.Run(); err != nil {
		api.InternalError(w, "failed to list convoys: "+err.Error())
		return
	}

	var convoys []struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &convoys); err != nil {
		api.InternalError(w, "failed to parse convoy list: "+err.Error())
		return
	}

	// Convert to API response format
	result := make([]ConvoyInfo, 0, len(convoys))
	for _, c := range convoys {
		info := ConvoyInfo{
			ID:        c.ID,
			Name:      c.Title,
			Status:    c.Status,
			CreatedAt: c.CreatedAt,
			Progress:  "0/0", // Will be calculated from tracked issues
			TrackedIDs: []string{},
		}
		result = append(result, info)
	}

	api.WriteJSON(w, result)
}

// Get returns a single convoy by ID.
func (h *ConvoysHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	townBeads := filepath.Join(h.townRoot, ".beads")

	// Get convoy details using bd show
	showCmd := exec.Command("bd", "show", id, "--json")
	showCmd.Dir = townBeads

	var stdout bytes.Buffer
	showCmd.Stdout = &stdout

	if err := showCmd.Run(); err != nil {
		api.NotFound(w, "convoy not found: "+id)
		return
	}

	var convoy struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &convoy); err != nil {
		api.InternalError(w, "failed to parse convoy: "+err.Error())
		return
	}

	// Return convoy info
	info := ConvoyInfo{
		ID:        convoy.ID,
		Name:      convoy.Title,
		Status:    convoy.Status,
		CreatedAt: convoy.CreatedAt,
		Progress:  "0/0",
		TrackedIDs: []string{},
	}

	api.WriteJSON(w, info)
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

	townBeads := filepath.Join(h.townRoot, ".beads")

	// Create convoy using bd create command with type=convoy
	// The name becomes the convoy title
	createArgs := []string{
		"create",
		"--type=convoy",
		fmt.Sprintf("--title=%s", req.Name),
	}
	createCmd := exec.Command("bd", createArgs...)
	createCmd.Dir = townBeads

	var stdout bytes.Buffer
	createCmd.Stdout = &stdout

	if err := createCmd.Run(); err != nil {
		api.InternalError(w, "failed to create convoy: "+err.Error())
		return
	}

	// Parse the created convoy ID from the output
	convoyID := strings.TrimSpace(stdout.String())
	if convoyID == "" {
		api.InternalError(w, "failed to get convoy ID from creation")
		return
	}

	// Add tracking dependencies for each issue
	for _, issueID := range req.IssueIDs {
		depCmd := exec.Command("bd", "dep", "add", convoyID, issueID, "--tracks")
		depCmd.Dir = townBeads
		if err := depCmd.Run(); err != nil {
			// Log error but continue - convoy was created successfully
		}
	}

	resp := CreateConvoyResponse{
		ID: convoyID,
	}
	api.WriteJSON(w, resp)
}
