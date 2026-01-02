package handlers

import (
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
