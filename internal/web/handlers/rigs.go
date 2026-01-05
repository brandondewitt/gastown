package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/web/api"
)

// RigsHandler handles rig-related API requests.
type RigsHandler struct {
	townRoot string
}

// NewRigsHandler creates a new rigs handler.
func NewRigsHandler(townRoot string) *RigsHandler {
	return &RigsHandler{townRoot: townRoot}
}

// RefinaryActionRequest represents a request to perform an action on refinery.
type RefinaryActionRequest struct {
	Action string `json:"action"`
}

// RefinaryActionResponse represents the response from a refinery action.
type RefinaryActionResponse struct {
	Action string `json:"action"`
	Status string `json:"status"`
}

// List returns all rigs.
func (h *RigsHandler) List(w http.ResponseWriter, r *http.Request) {
	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(true)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}
	api.WriteJSON(w, status.Rigs)
}

// Get returns a single rig by name.
func (h *RigsHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(false)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	for _, rig := range status.Rigs {
		if rig.Name == name {
			api.WriteJSON(w, rig)
			return
		}
	}

	api.NotFound(w, "rig not found: "+name)
}

// GetAgents returns agents for a specific rig.
func (h *RigsHandler) GetAgents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(false)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	for _, rig := range status.Rigs {
		if rig.Name == name {
			api.WriteJSON(w, rig.Agents)
			return
		}
	}

	api.NotFound(w, "rig not found: "+name)
}

// RefinaryAction performs an action (start, stop, restart) on a rig's refinery.
func (h *RigsHandler) RefinaryAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rigName := vars["rig"]
	action := vars["action"]

	// Validate rig exists
	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(false)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	rigFound := false
	for _, rig := range status.Rigs {
		if rig.Name == rigName {
			rigFound = true
			break
		}
	}

	if !rigFound {
		api.NotFound(w, "rig not found: "+rigName)
		return
	}

	// Validate action
	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
	}

	if !validActions[action] {
		api.BadRequest(w, "invalid action: "+action+" (must be start, stop, or restart)")
		return
	}

	// Decode request body (optional, for future payload expansion)
	var req RefinaryActionRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// TODO: Implement actual refinery control
	// This will be fully implemented in Phase 2/3
	// Actions would include sending signals to refinery process

	resp := RefinaryActionResponse{
		Action: action,
		Status: "accepted",
	}
	api.WriteJSON(w, resp)
}
