package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

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
