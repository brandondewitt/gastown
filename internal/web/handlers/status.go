package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/git"
	"github.com/steveyegge/gastown/internal/mail"
	"github.com/steveyegge/gastown/internal/rig"
	"github.com/steveyegge/gastown/internal/tmux"
	"github.com/steveyegge/gastown/internal/web/api"
)

// StatusHandler handles status-related API requests.
type StatusHandler struct {
	townRoot string
}

// NewStatusHandler creates a new status handler.
func NewStatusHandler(townRoot string) *StatusHandler {
	return &StatusHandler{townRoot: townRoot}
}

// TownStatus represents the overall status of the workspace.
// Mirrors the structure from internal/cmd/status.go for API responses.
type TownStatus struct {
	Name     string         `json:"name"`
	Location string         `json:"location"`
	Overseer *OverseerInfo  `json:"overseer,omitempty"`
	Agents   []AgentRuntime `json:"agents"`
	Rigs     []RigStatus    `json:"rigs"`
	Summary  StatusSummary  `json:"summary"`
}

// OverseerInfo represents the human operator.
type OverseerInfo struct {
	Name       string `json:"name"`
	Email      string `json:"email,omitempty"`
	Username   string `json:"username,omitempty"`
	Source     string `json:"source"`
	UnreadMail int    `json:"unread_mail"`
}

// AgentRuntime represents an agent's runtime state.
type AgentRuntime struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Session      string `json:"session"`
	Role         string `json:"role"`
	Running      bool   `json:"running"`
	HasWork      bool   `json:"has_work"`
	WorkTitle    string `json:"work_title,omitempty"`
	HookBead     string `json:"hook_bead,omitempty"`
	State        string `json:"state,omitempty"`
	UnreadMail   int    `json:"unread_mail"`
	FirstSubject string `json:"first_subject,omitempty"`
}

// RigStatus represents a rig's status.
type RigStatus struct {
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	Polecats     []string       `json:"polecats"`
	PolecatCount int            `json:"polecat_count"`
	Crews        []string       `json:"crews"`
	CrewCount    int            `json:"crew_count"`
	HasWitness   bool           `json:"has_witness"`
	HasRefinery  bool           `json:"has_refinery"`
	Agents       []AgentRuntime `json:"agents,omitempty"`
	MQ           *MQSummary     `json:"mq,omitempty"`
}

// MQSummary represents merge queue status.
type MQSummary struct {
	Pending  int    `json:"pending"`
	InFlight int    `json:"in_flight"`
	Blocked  int    `json:"blocked"`
	State    string `json:"state"`
	Health   string `json:"health"`
}

// StatusSummary provides summary counts.
type StatusSummary struct {
	RigCount      int `json:"rig_count"`
	PolecatCount  int `json:"polecat_count"`
	CrewCount     int `json:"crew_count"`
	WitnessCount  int `json:"witness_count"`
	RefineryCount int `json:"refinery_count"`
	ActiveHooks   int `json:"active_hooks"`
}

// GetStatus returns full town status.
func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.buildStatus(false)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}
	api.WriteJSON(w, status)
}

// GetSummary returns just the summary counts.
func (h *StatusHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	status, err := h.buildStatus(true)
	if err != nil {
		api.InternalError(w, err.Error())
		return
	}
	api.WriteJSON(w, status.Summary)
}

// buildStatus builds the full status, optionally in fast mode.
func (h *StatusHandler) buildStatus(fastMode bool) (*TownStatus, error) {
	// Load town config
	townConfigPath := constants.MayorTownPath(h.townRoot)
	townConfig, err := config.LoadTownConfig(townConfigPath)
	if err != nil {
		townConfig = &config.TownConfig{Name: filepath.Base(h.townRoot)}
	}

	// Load rigs config
	rigsConfigPath := constants.MayorRigsPath(h.townRoot)
	rigsConfig, err := config.LoadRigsConfig(rigsConfigPath)
	if err != nil {
		rigsConfig = &config.RigsConfig{Rigs: make(map[string]config.RigEntry)}
	}

	// Create rig manager
	g := git.NewGit(h.townRoot)
	mgr := rig.NewManager(h.townRoot, rigsConfig, g)

	// Create tmux instance
	t := tmux.NewTmux()

	// Pre-fetch all tmux sessions
	allSessions := make(map[string]bool)
	if sessions, err := t.ListSessions(); err == nil {
		for _, s := range sessions {
			allSessions[s] = true
		}
	}

	// Discover rigs
	rigs, err := mgr.DiscoverRigs()
	if err != nil {
		return nil, fmt.Errorf("discovering rigs: %w", err)
	}

	// Pre-fetch agent beads
	allAgentBeads := make(map[string]*beads.Issue)
	allHookBeads := make(map[string]*beads.Issue)
	for _, r := range rigs {
		rigBeadsPath := filepath.Join(r.Path, "mayor", "rig")
		rigBeads := beads.New(rigBeadsPath)
		rigAgentBeads, _ := rigBeads.ListAgentBeads()
		if rigAgentBeads == nil {
			continue
		}
		for id, issue := range rigAgentBeads {
			allAgentBeads[id] = issue
		}

		var hookIDs []string
		for _, issue := range rigAgentBeads {
			hookID := issue.HookBead
			if hookID == "" {
				fields := beads.ParseAgentFields(issue.Description)
				if fields != nil {
					hookID = fields.HookBead
				}
			}
			if hookID != "" {
				hookIDs = append(hookIDs, hookID)
			}
		}

		if len(hookIDs) > 0 {
			hookBeads, _ := rigBeads.ShowMultiple(hookIDs)
			for id, issue := range hookBeads {
				allHookBeads[id] = issue
			}
		}
	}

	// Create mail router
	mailRouter := mail.NewRouter(h.townRoot)

	// Load overseer config
	var overseerInfo *OverseerInfo
	if overseerConfig, err := config.LoadOrDetectOverseer(h.townRoot); err == nil && overseerConfig != nil {
		overseerInfo = &OverseerInfo{
			Name:     overseerConfig.Name,
			Email:    overseerConfig.Email,
			Username: overseerConfig.Username,
			Source:   overseerConfig.Source,
		}
		if !fastMode {
			if mailbox, err := mailRouter.GetMailbox("overseer"); err == nil {
				_, unread, _ := mailbox.Count()
				overseerInfo.UnreadMail = unread
			}
		}
	}

	// Build status
	status := &TownStatus{
		Name:     townConfig.Name,
		Location: h.townRoot,
		Overseer: overseerInfo,
		Rigs:     make([]RigStatus, len(rigs)),
	}

	var wg sync.WaitGroup

	// Fetch global agents in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		status.Agents = h.discoverGlobalAgents(allSessions, allAgentBeads, allHookBeads, mailRouter, fastMode)
	}()

	// Process rigs in parallel
	rigActiveHooks := make([]int, len(rigs))
	for i, r := range rigs {
		wg.Add(1)
		go func(idx int, r *rig.Rig) {
			defer wg.Done()
			rs, hooks := h.buildRigStatus(r, allSessions, allAgentBeads, allHookBeads, mailRouter, fastMode)
			status.Rigs[idx] = rs
			rigActiveHooks[idx] = hooks
		}(i, r)
	}

	wg.Wait()

	// Build summary
	for _, rs := range status.Rigs {
		status.Summary.RigCount++
		status.Summary.PolecatCount += rs.PolecatCount
		status.Summary.CrewCount += rs.CrewCount
		if rs.HasWitness {
			status.Summary.WitnessCount++
		}
		if rs.HasRefinery {
			status.Summary.RefineryCount++
		}
	}
	for _, hooks := range rigActiveHooks {
		status.Summary.ActiveHooks += hooks
	}

	return status, nil
}

// discoverGlobalAgents finds global agents (Mayor, Deacon).
func (h *StatusHandler) discoverGlobalAgents(
	allSessions map[string]bool,
	allAgentBeads map[string]*beads.Issue,
	allHookBeads map[string]*beads.Issue,
	mailRouter *mail.Router,
	fastMode bool,
) []AgentRuntime {
	agents := []AgentRuntime{}

	// Check for Mayor
	mayorSession := "gt-mayor"
	if allSessions[mayorSession] {
		agent := AgentRuntime{
			Name:    "mayor",
			Address: "mayor",
			Session: mayorSession,
			Role:    "mayor",
			Running: true,
		}
		if !fastMode {
			if mailbox, err := mailRouter.GetMailbox("mayor"); err == nil {
				_, unread, _ := mailbox.Count()
				agent.UnreadMail = unread
			}
		}
		agents = append(agents, agent)
	}

	// Check for Deacon
	deaconSession := "gt-deacon"
	if allSessions[deaconSession] {
		agent := AgentRuntime{
			Name:    "deacon",
			Address: "deacon",
			Session: deaconSession,
			Role:    "deacon",
			Running: true,
		}
		if !fastMode {
			if mailbox, err := mailRouter.GetMailbox("deacon"); err == nil {
				_, unread, _ := mailbox.Count()
				agent.UnreadMail = unread
			}
		}
		agents = append(agents, agent)
	}

	return agents
}

// buildRigStatus builds status for a single rig.
func (h *StatusHandler) buildRigStatus(
	r *rig.Rig,
	allSessions map[string]bool,
	allAgentBeads map[string]*beads.Issue,
	allHookBeads map[string]*beads.Issue,
	mailRouter *mail.Router,
	fastMode bool,
) (RigStatus, int) {
	rs := RigStatus{
		Name:         r.Name,
		Path:         r.Path,
		Polecats:     r.Polecats,
		PolecatCount: len(r.Polecats),
		Crews:        r.Crew,
		CrewCount:    len(r.Crew),
		HasWitness:   r.HasWitness,
		HasRefinery:  r.HasRefinery,
	}

	activeHooks := 0

	// Build agent list for this rig
	var agents []AgentRuntime

	// Add polecats
	for _, name := range r.Polecats {
		session := fmt.Sprintf("gt-%s-%s", r.Name, name)
		address := fmt.Sprintf("%s/%s", r.Name, name)
		agent := AgentRuntime{
			Name:    name,
			Address: address,
			Session: session,
			Role:    "polecat",
			Running: allSessions[session],
		}

		// Check for hook
		if bead, ok := allAgentBeads[address]; ok {
			hookID := bead.HookBead
			if hookID == "" {
				if fields := beads.ParseAgentFields(bead.Description); fields != nil {
					hookID = fields.HookBead
				}
			}
			if hookID != "" {
				agent.HasWork = true
				agent.HookBead = hookID
				if hookBead, ok := allHookBeads[hookID]; ok {
					agent.WorkTitle = hookBead.Title
				}
				activeHooks++
			}
			agent.State = bead.AgentState
		}

		if !fastMode {
			if mailbox, err := mailRouter.GetMailbox(address); err == nil {
				_, unread, _ := mailbox.Count()
				agent.UnreadMail = unread
			}
		}

		agents = append(agents, agent)
	}

	// Add crews
	for _, name := range r.Crew {
		session := fmt.Sprintf("gt-%s-%s", r.Name, name)
		address := fmt.Sprintf("%s/%s", r.Name, name)
		agent := AgentRuntime{
			Name:    name,
			Address: address,
			Session: session,
			Role:    "crew",
			Running: allSessions[session],
		}

		if bead, ok := allAgentBeads[address]; ok {
			hookID := bead.HookBead
			if hookID == "" {
				if fields := beads.ParseAgentFields(bead.Description); fields != nil {
					hookID = fields.HookBead
				}
			}
			if hookID != "" {
				agent.HasWork = true
				agent.HookBead = hookID
				if hookBead, ok := allHookBeads[hookID]; ok {
					agent.WorkTitle = hookBead.Title
				}
				activeHooks++
			}
			agent.State = bead.AgentState
		}

		if !fastMode {
			if mailbox, err := mailRouter.GetMailbox(address); err == nil {
				_, unread, _ := mailbox.Count()
				agent.UnreadMail = unread
			}
		}

		agents = append(agents, agent)
	}

	// Add witness if present
	if r.HasWitness {
		session := fmt.Sprintf("gt-%s-witness", r.Name)
		address := fmt.Sprintf("%s/witness", r.Name)
		agent := AgentRuntime{
			Name:    "witness",
			Address: address,
			Session: session,
			Role:    "witness",
			Running: allSessions[session],
		}
		agents = append(agents, agent)
	}

	// Add refinery if present
	if r.HasRefinery {
		session := fmt.Sprintf("gt-%s-refinery", r.Name)
		address := fmt.Sprintf("%s/refinery", r.Name)
		agent := AgentRuntime{
			Name:    "refinery",
			Address: address,
			Session: session,
			Role:    "refinery",
			Running: allSessions[session],
		}
		agents = append(agents, agent)
	}

	rs.Agents = agents
	return rs, activeHooks
}
