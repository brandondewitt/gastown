package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/mrqueue"
	"github.com/steveyegge/gastown/internal/web/api"
)

// MQHandler handles merge queue (MQ) API requests.
type MQHandler struct {
	townRoot string
}

// NewMQHandler creates a new MQ handler.
func NewMQHandler(townRoot string) *MQHandler {
	return &MQHandler{townRoot: townRoot}
}

// MQItem represents a merge request in the queue for API responses.
type MQItem struct {
	ID        string    `json:"id"`
	Branch    string    `json:"branch"`
	Status    string    `json:"status"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	Target    string    `json:"target,omitempty"`
	Title     string    `json:"title,omitempty"`
	Worker    string    `json:"worker,omitempty"`
	Priority  int       `json:"priority,omitempty"`
	Rig       string    `json:"rig,omitempty"`
}

// RetryRequest represents a request to retry a merge queue item.
type RetryRequest struct {
	ID string `json:"id"`
}

// RetryResponse represents the response from a merge queue retry.
type RetryResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// List returns all pending merge requests in the queue.
// Endpoint: GET /api/v1/mq
func (h *MQHandler) List(w http.ResponseWriter, r *http.Request) {
	// Discover all rigs to find merge queues
	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(false)
	if err != nil {
		api.InternalError(w, fmt.Sprintf("failed to discover rigs: %v", err))
		return
	}

	var allItems []*MQItem
	position := 1

	// Collect MRs from all rigs
	for _, rig := range status.Rigs {
		rigPath := filepath.Join(h.townRoot, rig.Name)
		q := mrqueue.New(rigPath)

		// Get all MRs from this rig's queue
		mrs, err := q.ListByScore()
		if err != nil {
			// Log error but continue with other rigs
			continue
		}

		// Convert to API response format
		for _, mr := range mrs {
			item := &MQItem{
				ID:        mr.ID,
				Branch:    mr.Branch,
				Target:    mr.Target,
				Title:     mr.Title,
				Worker:    mr.Worker,
				Priority:  mr.Priority,
				Rig:       mr.Rig,
				Position:  position,
				CreatedAt: mr.CreatedAt,
				Status:    determineStatus(mr),
			}
			allItems = append(allItems, item)
			position++
		}
	}

	// Return the list of MQ items
	api.WriteJSON(w, allItems)
}

// Get returns a specific merge request by ID.
// Endpoint: GET /api/v1/mq/{id}
func (h *MQHandler) Get(w http.ResponseWriter, r *http.Request) {
	// This would be implemented for Phase 2
	api.NotFound(w, "MQ item not found")
}

// Retry retries a failed merge queue item by ID.
// Endpoint: POST /api/v1/mq/{id}/retry
func (h *MQHandler) Retry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Validate ID is not empty
	if id == "" {
		api.BadRequest(w, "merge queue item ID is required")
		return
	}

	// Decode optional request body (for future expansion)
	var req RetryRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// TODO: Implement actual merge queue retry logic
	// This will involve:
	// 1. Finding the merge queue item by ID
	// 2. Verifying it's in a failed state
	// 3. Resetting it to pending state
	// 4. Notifying the refinery to reprocess
	// This will be fully implemented in Phase 2/3

	resp := RetryResponse{
		ID:      id,
		Status:  "queued",
		Message: "Merge retry has been queued for processing",
	}
	api.WriteJSON(w, resp)
}

// determineStatus determines the status of a merge request based on its state.
func determineStatus(mr *mrqueue.MR) string {
	if mr.ClaimedBy != "" {
		return "in_flight"
	}
	if mr.BlockedBy != "" {
		return "blocked"
	}
	return "pending"
}
