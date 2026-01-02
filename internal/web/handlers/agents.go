package handlers

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/web/api"
)

// AgentsHandler handles agent-related API requests.
type AgentsHandler struct {
	townRoot string
}

// NewAgentsHandler creates a new agents handler.
func NewAgentsHandler(townRoot string) *AgentsHandler {
	return &AgentsHandler{townRoot: townRoot}
}

// List returns all agents across all rigs.
func (h *AgentsHandler) List(w http.ResponseWriter, r *http.Request) {
	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(false)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	var agents []AgentRuntime

	// Add global agents
	agents = append(agents, status.Agents...)

	// Add rig agents
	for _, rig := range status.Rigs {
		agents = append(agents, rig.Agents...)
	}

	api.WriteJSON(w, agents)
}

// Get returns a single agent by address (e.g., "rigname/agentname").
func (h *AgentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(false)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	// Check global agents
	for _, agent := range status.Agents {
		if agent.Address == address {
			api.WriteJSON(w, agent)
			return
		}
	}

	// Check rig agents
	for _, rig := range status.Rigs {
		for _, agent := range rig.Agents {
			if agent.Address == address {
				api.WriteJSON(w, agent)
				return
			}
		}
	}

	// Also try matching just the name part for global agents
	parts := strings.Split(address, "/")
	if len(parts) == 1 {
		for _, agent := range status.Agents {
			if agent.Name == address {
				api.WriteJSON(w, agent)
				return
			}
		}
	}

	api.NotFound(w, "agent not found: "+address)
}
