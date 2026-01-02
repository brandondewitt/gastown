package handlers

import (
	"net/http"
	"strconv"

	"github.com/steveyegge/gastown/internal/web/api"
)

// EventsHandler handles event-related API requests.
type EventsHandler struct {
	townRoot string
}

// NewEventsHandler creates a new events handler.
func NewEventsHandler(townRoot string) *EventsHandler {
	return &EventsHandler{townRoot: townRoot}
}

// FeedEvent represents an event for API responses.
type FeedEvent struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor,omitempty"`
	Rig       string `json:"rig,omitempty"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
}

// List returns recent events with pagination.
func (h *EventsHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// TODO: Implement event listing from .events.jsonl
	// This will be fully implemented in Phase 2
	_ = limit
	_ = offset

	api.WritePaginated(w, []FeedEvent{}, 0, offset, limit)
}
