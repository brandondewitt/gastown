package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

	// Read events from .events.jsonl file
	eventsPath := filepath.Join(h.townRoot, ".events.jsonl")
	events, err := h.readEventsFile(eventsPath)
	if err != nil {
		// If file doesn't exist or can't be read, return empty list
		api.WritePaginated(w, []FeedEvent{}, 0, offset, limit)
		return
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	// Apply offset and limit
	total := len(events)
	if offset >= total {
		api.WritePaginated(w, []FeedEvent{}, total, offset, limit)
		return
	}

	end := offset + limit
	if end > total {
		end = total
	}

	api.WritePaginated(w, events[offset:end], total, offset, limit)
}

// rawEvent represents the raw event structure from the JSONL file
type rawEvent struct {
	Timestamp  string                 `json:"ts"`
	Source     string                 `json:"source"`
	Type       string                 `json:"type"`
	Actor      string                 `json:"actor"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
	Visibility string                 `json:"visibility"`
}

// readEventsFile reads and parses the events JSONL file
func (h *EventsHandler) readEventsFile(filePath string) ([]FeedEvent, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []FeedEvent{}, nil // Return empty list if file doesn't exist
	}
	defer file.Close()

	var events []FeedEvent
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw rawEvent
		if err := json.Unmarshal(line, &raw); err != nil {
			continue // Skip malformed lines
		}

		// Build message from event
		message := h.buildEventMessage(raw)

		// Extract rig from actor if present
		rig := h.extractRig(raw.Actor)

		events = append(events, FeedEvent{
			ID:        raw.Type + "-" + raw.Timestamp, // Generate ID from type and timestamp
			Type:      raw.Type,
			Timestamp: raw.Timestamp,
			Actor:     raw.Actor,
			Rig:       rig,
			Message:   message,
			Details:   raw.Payload,
		})
	}

	return events, nil
}

// buildEventMessage creates a human-readable message from an event
func (h *EventsHandler) buildEventMessage(event rawEvent) string {
	switch event.Type {
	case "sling":
		if event.Payload != nil {
			if target, ok := event.Payload["target"]; ok {
				return "Slung work to " + toString(target)
			}
		}
		return "Work slung"
	case "hook":
		return "Work hooked"
	case "unhook":
		return "Work unhooked"
	case "handoff":
		return "Handoff created"
	case "done":
		if event.Payload != nil {
			if branch, ok := event.Payload["branch"]; ok {
				return "Work completed on " + toString(branch)
			}
		}
		return "Work completed"
	case "mail":
		if event.Payload != nil {
			if subject, ok := event.Payload["subject"]; ok {
				return "Mail sent: " + toString(subject)
			}
		}
		return "Mail sent"
	case "spawn":
		if event.Payload != nil {
			if polecat, ok := event.Payload["polecat"]; ok {
				return "Spawned " + toString(polecat)
			}
		}
		return "Polecat spawned"
	case "kill":
		return "Service killed"
	case "nudge":
		return "Nudge sent"
	case "boot":
		return "Rig booted"
	case "halt":
		return "Services halted"
	case "patrol_started":
		return "Witness patrol started"
	case "patrol_complete":
		return "Witness patrol completed"
	case "polecat_checked":
		return "Polecat checked"
	case "merge_started":
		return "Merge started"
	case "merged":
		return "Merge completed"
	case "merge_failed":
		return "Merge failed"
	default:
		return event.Type
	}
}

// extractRig extracts the rig name from an actor address
func (h *EventsHandler) extractRig(actor string) string {
	// Actor format: "rig/agent" or "rig"
	parts := splitActor(actor)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// splitActor splits an actor address into parts
func splitActor(actor string) []string {
	var parts []string
	var current string
	for _, ch := range actor {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// toString converts a value to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case nil:
		return ""
	default:
		if b, err := json.Marshal(val); err == nil {
			return string(b)
		}
		return ""
	}
}
