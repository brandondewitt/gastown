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

// SlingHandler handles work dispatching (sling) API requests.
type SlingHandler struct {
	townRoot string
}

// NewSlingHandler creates a new sling handler.
func NewSlingHandler(townRoot string) *SlingHandler {
	return &SlingHandler{townRoot: townRoot}
}

// SlingRequest represents a request to dispatch work to an agent.
type SlingRequest struct {
	// IssueID is the beads issue/task ID to dispatch
	IssueID string `json:"issue_id"`
	// Target is the agent address or rig name to dispatch to
	Target string `json:"target"`
	// Message is optional context/instructions for the dispatch
	Message string `json:"message,omitempty"`
}

// SlingResponse represents the result of a sling operation.
type SlingResponse struct {
	// Success indicates if the dispatch was successful
	Success bool `json:"success"`
	// Message describes the result
	Message string `json:"message"`
	// Output contains the command output
	Output string `json:"output,omitempty"`
	// IssueID echoes the dispatched issue ID
	IssueID string `json:"issue_id"`
	// Target echoes the target agent
	Target string `json:"target"`
}

// Dispatch handles POST requests to dispatch work to agents.
func (h *SlingHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	var req SlingRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.IssueID == "" {
		api.BadRequest(w, "issue_id is required")
		return
	}
	if req.Target == "" {
		api.BadRequest(w, "target agent/rig is required")
		return
	}

	// Build gt sling command
	args := []string{"sling", req.IssueID, req.Target}
	if req.Message != "" {
		args = append(args, "--args", req.Message)
	}

	// Execute 'gt sling <issue_id> <target>' from town root
	ctx := r.Context()
	cmd := exec.CommandContext(ctx, "gt", args...)
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
			api.BadRequest(w, "sling failed: "+errMsg)
			return
		}
	case <-time.After(30 * time.Second):
		api.InternalError(w, "sling timeout")
		return
	}

	// Build success response
	output := strings.TrimSpace(stdout.String())
	resp := SlingResponse{
		Success: true,
		Message: "Work dispatched successfully",
		Output:  output,
		IssueID: req.IssueID,
		Target:  req.Target,
	}

	api.WriteJSON(w, resp)
}
