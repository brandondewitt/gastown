package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/mail"
	"github.com/steveyegge/gastown/internal/townlog"
	"github.com/steveyegge/gastown/internal/web/api"
)

// AgentDetailResponse provides detailed information about an agent.
type AgentDetailResponse struct {
	Agent         AgentRuntime    `json:"agent"`
	Mail          []MailPreview   `json:"mail"`
	RecentEvents  []AgentEvent    `json:"recent_events"`
	TotalMail     int             `json:"total_mail"`
	TotalUnread   int             `json:"total_unread"`
}

// MailPreview is a summary of a mail message for display.
type MailPreview struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	Subject   string    `json:"subject"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
	Priority  string    `json:"priority,omitempty"`
}

// AgentEvent represents an activity event for the agent.
type AgentEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Context   string    `json:"context,omitempty"`
}

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

// GetDetails returns detailed information about an agent including mail and activity.
func (h *AgentsHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Parse query params for mail limit
	mailLimitStr := r.URL.Query().Get("mail_limit")
	mailLimit := 10
	if mailLimitStr != "" {
		if l, err := strconv.Atoi(mailLimitStr); err == nil && l > 0 && l <= 50 {
			mailLimit = l
		}
	}

	// Parse query params for events limit
	eventsLimitStr := r.URL.Query().Get("events_limit")
	eventsLimit := 20
	if eventsLimitStr != "" {
		if l, err := strconv.Atoi(eventsLimitStr); err == nil && l > 0 && l <= 100 {
			eventsLimit = l
		}
	}

	statusHandler := NewStatusHandler(h.townRoot)
	status, err := statusHandler.buildStatus(false)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}

	// Find the agent
	var agent *AgentRuntime
	for i := range status.Agents {
		if status.Agents[i].Address == address || status.Agents[i].Name == address {
			agent = &status.Agents[i]
			break
		}
	}
	if agent == nil {
		for _, rig := range status.Rigs {
			for i := range rig.Agents {
				if rig.Agents[i].Address == address {
					agent = &rig.Agents[i]
					break
				}
			}
			if agent != nil {
				break
			}
		}
	}

	if agent == nil {
		api.NotFound(w, "agent not found: "+address)
		return
	}

	// Build the response
	response := AgentDetailResponse{
		Agent: *agent,
	}

	// Get mail for this agent
	mailRouter := mail.NewRouter(h.townRoot)
	if mailbox, err := mailRouter.GetMailbox(agent.Address); err == nil {
		total, unread, _ := mailbox.Count()
		response.TotalMail = total
		response.TotalUnread = unread

		// Get recent mail
		if messages, err := mailbox.List(); err == nil {
			count := 0
			for _, msg := range messages {
				if count >= mailLimit {
					break
				}
				preview := MailPreview{
					ID:        msg.ID,
					From:      msg.From,
					Subject:   msg.Subject,
					Read:      msg.Read,
					Priority:  string(msg.Priority),
					Timestamp: msg.Timestamp,
				}
				response.Mail = append(response.Mail, preview)
				count++
			}
		}
	}

	// Get recent events for this agent
	if events, err := townlog.ReadEvents(h.townRoot); err == nil {
		// Filter events for this agent
		filtered := townlog.FilterEvents(events, townlog.Filter{
			Agent: agent.Address,
		})

		// Get the most recent events (up to limit)
		start := 0
		if len(filtered) > eventsLimit {
			start = len(filtered) - eventsLimit
		}
		for i := len(filtered) - 1; i >= start; i-- {
			e := filtered[i]
			response.RecentEvents = append(response.RecentEvents, AgentEvent{
				Timestamp: e.Timestamp,
				Type:      string(e.Type),
				Context:   e.Context,
			})
		}
	}

	api.WriteJSON(w, response)
}

// findAgent is a helper to find an agent by address.
func (h *AgentsHandler) findAgent(status *TownStatus, address string) *AgentRuntime {
	// Check global agents
	for i := range status.Agents {
		if status.Agents[i].Address == address || status.Agents[i].Name == address {
			return &status.Agents[i]
		}
	}

	// Check rig agents
	parts := strings.Split(address, "/")
	if len(parts) == 2 {
		for _, rig := range status.Rigs {
			if rig.Name == parts[0] {
				for i := range rig.Agents {
					if rig.Agents[i].Name == parts[1] || rig.Agents[i].Address == address {
						return &rig.Agents[i]
					}
				}
			}
		}
	}

	return nil
}
