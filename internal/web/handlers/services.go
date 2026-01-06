package handlers

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/git"
	"github.com/steveyegge/gastown/internal/rig"
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

// HandleWitness handles start/stop/restart actions for witness service.
func (h *ServicesHandler) HandleWitness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rigName := vars["rig"]
	action := vars["action"]

	// Validate action
	if action != "start" && action != "stop" && action != "restart" {
		api.BadRequest(w, "invalid action: must be 'start', 'stop', or 'restart'")
		return
	}

	// Get rig information
	g := git.NewGit(h.townRoot)
	rigsConfigPath := constants.MayorRigsPath(h.townRoot)
	rigsConfig, err := config.LoadRigsConfig(rigsConfigPath)
	if err != nil {
		response := ServiceActionResponse{
			Success: false,
			Error:   "failed to load rigs config: " + err.Error(),
		}
		api.WriteJSON(w, response)
		return
	}

	mgr := rig.NewManager(h.townRoot, rigsConfig, g)
	targetRig, err := mgr.GetRig(rigName)
	if err != nil {
		response := ServiceActionResponse{
			Success: false,
			Error:   "rig not found: " + rigName,
		}
		api.WriteJSON(w, response)
		return
	}

	// Execute the witness command
	cmd := exec.Command("gt", "witness", action)
	cmd.Dir = targetRig.Path

	if err := cmd.Run(); err != nil {
		response := ServiceActionResponse{
			Success: false,
			Running: false,
			Error:   fmt.Sprintf("failed to %s witness: %v", action, err),
		}
		api.WriteJSON(w, response)
		return
	}

	// Determine running state based on action
	running := action != "stop"

	response := ServiceActionResponse{
		Success: true,
		Running: running,
	}

	api.WriteJSON(w, response)
}

// HandleRefinery handles start/stop/restart actions for refinery service.
func (h *ServicesHandler) HandleRefinery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rigName := vars["rig"]
	action := vars["action"]

	// Validate action
	if action != "start" && action != "stop" && action != "restart" {
		api.BadRequest(w, "invalid action: must be 'start', 'stop', or 'restart'")
		return
	}

	// Get rig information
	g := git.NewGit(h.townRoot)
	rigsConfigPath := constants.MayorRigsPath(h.townRoot)
	rigsConfig, err := config.LoadRigsConfig(rigsConfigPath)
	if err != nil {
		response := ServiceActionResponse{
			Success: false,
			Error:   "failed to load rigs config: " + err.Error(),
		}
		api.WriteJSON(w, response)
		return
	}

	mgr := rig.NewManager(h.townRoot, rigsConfig, g)
	targetRig, err := mgr.GetRig(rigName)
	if err != nil {
		response := ServiceActionResponse{
			Success: false,
			Error:   "rig not found: " + rigName,
		}
		api.WriteJSON(w, response)
		return
	}

	// Execute the refinery command
	cmd := exec.Command("gt", "refinery", action)
	cmd.Dir = targetRig.Path

	if err := cmd.Run(); err != nil {
		response := ServiceActionResponse{
			Success: false,
			Running: false,
			Error:   fmt.Sprintf("failed to %s refinery: %v", action, err),
		}
		api.WriteJSON(w, response)
		return
	}

	// Determine running state based on action
	running := action != "stop"

	response := ServiceActionResponse{
		Success: true,
		Running: running,
	}

	api.WriteJSON(w, response)
}
