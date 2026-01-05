package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/web/api"
)

// ServicesHandler handles service control API requests.
type ServicesHandler struct {
	townRoot string
}

// NewServicesHandler creates a new services handler.
func NewServicesHandler(townRoot string) *ServicesHandler {
	return &ServicesHandler{townRoot: townRoot}
}

// ServiceActionResponse represents the response from a service action.
type ServiceActionResponse struct {
	Success bool   `json:"success"`
	Running bool   `json:"running"`
	Error   string `json:"error,omitempty"`
}

// HandleWitness handles start/stop actions for witness service.
func (h *ServicesHandler) HandleWitness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_ = vars["rig"] // TODO: Use rig name when implementing service control
	action := vars["action"]

	// Validate action
	if action != "start" && action != "stop" {
		api.BadRequest(w, "invalid action: must be 'start' or 'stop'")
		return
	}

	// TODO: Implement witness service control logic
	// For now, return a placeholder response
	response := ServiceActionResponse{
		Success: true,
		Running: action == "start",
	}

	api.WriteJSON(w, response)
}

// HandleRefinery handles start/stop actions for refinery service.
func (h *ServicesHandler) HandleRefinery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_ = vars["rig"] // TODO: Use rig name when implementing service control
	action := vars["action"]

	// Validate action
	if action != "start" && action != "stop" {
		api.BadRequest(w, "invalid action: must be 'start' or 'stop'")
		return
	}

	// TODO: Implement refinery service control logic
	// For now, return a placeholder response
	response := ServiceActionResponse{
		Success: true,
		Running: action == "start",
	}

	api.WriteJSON(w, response)
}
