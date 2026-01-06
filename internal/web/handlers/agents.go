package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/tmux"
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

// GetOutput returns recent terminal output for an agent's tmux session.
func (h *AgentsHandler) GetOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Convert address to session name: "gastown/toast" -> "gt-gastown-toast"
	sessionName := "gt-" + strings.ReplaceAll(address, "/", "-")

	t := tmux.NewTmux()
	output, err := t.CapturePane(sessionName, 100)
	if err != nil {
		api.WriteJSON(w, map[string]interface{}{
			"output":  "",
			"error":   err.Error(),
			"session": sessionName,
		})
		return
	}

	api.WriteJSON(w, map[string]interface{}{
		"output":  output,
		"session": sessionName,
	})
}

// SendMessageRequest is the request body for sending a message to an agent.
type SendMessageRequest struct {
	Message string `json:"message"`
}

// SendMessage sends a message to an agent's tmux session.
func (h *AgentsHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request: "+err.Error())
		return
	}

	if req.Message == "" {
		api.BadRequest(w, "message is required")
		return
	}

	// Convert address to session name: "gastown/toast" -> "gt-gastown-toast"
	sessionName := "gt-" + strings.ReplaceAll(address, "/", "-")

	t := tmux.NewTmux()
	if err := t.NudgeSession(sessionName, req.Message); err != nil {
		api.InternalError(w, "failed to send message: "+err.Error())
		return
	}

	api.WriteJSON(w, map[string]bool{"success": true})
}
