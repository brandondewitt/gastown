package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/steveyegge/gastown/internal/web/api"
)

// BeadsHandler handles bead/issue creation API requests.
type BeadsHandler struct {
	townRoot string
}

// NewBeadsHandler creates a new beads handler.
func NewBeadsHandler(townRoot string) *BeadsHandler {
	return &BeadsHandler{townRoot: townRoot}
}

// BeadItem represents a bead/issue in list responses.
type BeadItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	PriorityStr string `json:"priority_str,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// CreateBeadRequest represents a request to create a new bead/issue.
type CreateBeadRequest struct {
	// Title is the issue title (required)
	Title string `json:"title"`
	// Type is the issue type: task, bug, feature, epic, chore (default: task)
	Type string `json:"type,omitempty"`
	// Description is the issue description
	Description string `json:"description,omitempty"`
	// Priority is the issue priority: 0-4 or P0-P4
	Priority string `json:"priority,omitempty"`
	// Rig is the optional rig to create the issue in
	Rig string `json:"rig,omitempty"`
}

// CreateBeadResponse represents the result of creating a bead.
type CreateBeadResponse struct {
	// Success indicates if the creation was successful
	Success bool `json:"success"`
	// ID is the new bead ID
	ID string `json:"id"`
	// Title echoes the bead title
	Title string `json:"title"`
	// Type echoes the bead type
	Type string `json:"type"`
	// Message describes the result
	Message string `json:"message"`
}

// Create handles POST requests to create new beads/issues.
func (h *BeadsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateBeadRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.Title == "" {
		api.BadRequest(w, "title is required")
		return
	}

	// Build bd create command
	args := []string{"create", req.Title, "--silent"}

	// Add optional type
	if req.Type != "" {
		args = append(args, "--type", req.Type)
	}

	// Add optional priority
	if req.Priority != "" {
		args = append(args, "--priority", req.Priority)
	}

	// Add optional rig
	if req.Rig != "" {
		args = append(args, "--rig", req.Rig)
	}

	// Execute 'bd create <title>' from town root
	ctx := r.Context()
	cmd := exec.CommandContext(ctx, "bd", args...)
	cmd.Dir = h.townRoot

	// Capture stderr and stdout
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Wait for command with timeout
	select {
	case err := <-done:
		if err != nil {
			// Command failed
			errMsg := strings.TrimSpace(stderr.String())
			if errMsg == "" {
				errMsg = err.Error()
			}
			api.BadRequest(w, "failed to create bead: "+errMsg)
			return
		}
	case <-time.After(30 * time.Second):
		api.InternalError(w, "bead creation timeout")
		return
	}

	// Parse the bead ID from output (--silent flag outputs just the ID)
	beadID := strings.TrimSpace(stdout.String())
	// Remove any warning lines (daemon warnings, etc.)
	lines := strings.Split(beadID, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// The bead ID should be the first non-warning line
		if line != "" && !strings.HasPrefix(line, "Warning:") && !strings.HasPrefix(line, "Hint:") {
			beadID = line
			break
		}
	}

	// Determine the type used
	beadType := req.Type
	if beadType == "" {
		beadType = "task"
	}

	// Build success response
	resp := CreateBeadResponse{
		Success: true,
		ID:      beadID,
		Title:   req.Title,
		Type:    beadType,
		Message: "Bead created successfully",
	}

	api.WriteJSON(w, resp)
}

// List handles GET requests to list beads/issues.
func (h *BeadsHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get query parameters for filtering
	status := r.URL.Query().Get("status")
	beadType := r.URL.Query().Get("type")

	// Build bd list command with --json flag
	args := []string{"list", "--json"}

	// Add optional status filter
	if status != "" && status != "all" {
		args = append(args, "--status", status)
	}

	// Add optional type filter
	if beadType != "" {
		args = append(args, "--type", beadType)
	}

	// Execute 'bd list --json' from town root
	ctx := r.Context()
	cmd := exec.CommandContext(ctx, "bd", args...)
	cmd.Dir = h.townRoot

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Wait for command with timeout
	select {
	case err := <-done:
		if err != nil {
			// Command failed
			errMsg := strings.TrimSpace(stderr.String())
			if errMsg == "" {
				errMsg = err.Error()
			}
			api.BadRequest(w, "failed to list beads: "+errMsg)
			return
		}
	case <-time.After(30 * time.Second):
		api.InternalError(w, "bead list timeout")
		return
	}

	// Parse the JSON output from bd list
	output := strings.TrimSpace(stdout.String())

	// Handle empty output
	if output == "" || output == "[]" {
		api.WriteJSON(w, []BeadItem{})
		return
	}

	// Parse the JSON output - bd list --json outputs a JSON array
	var beads []BeadItem

	// First, try to find the JSON content (skip any warning lines)
	jsonStart := strings.Index(output, "[")
	if jsonStart == -1 {
		// No array found - try object with data field
		jsonStart = strings.Index(output, "{")
	}

	if jsonStart == -1 {
		api.WriteJSON(w, []BeadItem{})
		return
	}

	jsonContent := output[jsonStart:]

	// Check if it's an array or wrapped object
	if strings.HasPrefix(jsonContent, "[") {
		// Direct array format
		if err := json.Unmarshal([]byte(jsonContent), &beads); err != nil {
			api.InternalError(w, "failed to parse beads list: "+err.Error())
			return
		}
	} else {
		// Wrapped format: {"data": [...]}
		var wrappedResp struct {
			Data []BeadItem `json:"data"`
		}
		if err := json.Unmarshal([]byte(jsonContent), &wrappedResp); err != nil {
			api.InternalError(w, "failed to parse beads response: "+err.Error())
			return
		}
		beads = wrappedResp.Data
	}

	if beads == nil {
		beads = []BeadItem{}
	}

	api.WriteJSON(w, beads)
}
