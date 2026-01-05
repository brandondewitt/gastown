package handlers

import (
	"encoding/json"
	"net/http"

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
	// DispatchID is the unique ID for this dispatch
	DispatchID string `json:"dispatch_id"`
	// Message describes the result
	Message string `json:"message"`
	// IssuID echoes the dispatched issue ID
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

	// Create a dispatch ID for tracking
	// In a real implementation, this would create an actual dispatch record
	dispatchID := generateDispatchID()

	// Build response
	resp := SlingResponse{
		Success:    true,
		DispatchID: dispatchID,
		IssueID:    req.IssueID,
		Target:     req.Target,
		Message:    "Work dispatched successfully",
	}

	// In a real implementation, this would:
	// 1. Validate the issue exists
	// 2. Validate the target agent/rig is available
	// 3. Create a dispatch record in the beads system
	// 4. Send a message/hook to the target agent
	// 5. Return confirmation with dispatch tracking ID

	api.WriteJSON(w, resp)
}

// generateDispatchID generates a unique ID for dispatch tracking.
// In a real implementation, this would use the beads/Gas Town ID system.
func generateDispatchID() string {
	// Simple placeholder - real implementation would use proper ID generation
	return "dispatch-" + randString(12)
}

// randString generates a random string of specified length.
// Real implementation would use a proper random ID generator.
func randString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}
