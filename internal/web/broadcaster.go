// Package web provides the HTTP server for the Gas Town web dashboard.
package web

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/git"
	"github.com/steveyegge/gastown/internal/rig"
	"github.com/steveyegge/gastown/internal/tmux"
	"github.com/steveyegge/gastown/internal/web/api"
	"github.com/steveyegge/gastown/internal/web/handlers"
)

const (
	// Default poll interval for status changes
	defaultPollInterval = 3 * time.Second

	// Minimum poll interval allowed
	minPollInterval = 1 * time.Second
)

// Broadcaster watches for status changes and broadcasts them via WebSocket.
type Broadcaster struct {
	townRoot     string
	hub          wsHub
	pollInterval time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup

	// Previous state for change detection
	mu             sync.Mutex
	prevAgents     map[string]agentSnapshot
	prevConvoys    map[string]string // id -> status
	prevMQCounts   map[string]int    // rig -> pending count
}

// wsHub interface for broadcasting - matches ws.Hub
type wsHub interface {
	Broadcast(msg *api.WSMessage)
}

// agentSnapshot captures agent state for comparison
type agentSnapshot struct {
	Running   bool
	HasWork   bool
	HookBead  string
	State     string
	WorkTitle string
}

// NewBroadcaster creates a new event broadcaster.
func NewBroadcaster(townRoot string, hub wsHub) *Broadcaster {
	return &Broadcaster{
		townRoot:     townRoot,
		hub:          hub,
		pollInterval: defaultPollInterval,
		stopCh:       make(chan struct{}),
		prevAgents:   make(map[string]agentSnapshot),
		prevConvoys:  make(map[string]string),
		prevMQCounts: make(map[string]int),
	}
}

// Start begins the broadcaster polling loop.
func (b *Broadcaster) Start() {
	b.wg.Add(1)
	go b.pollLoop()
	log.Printf("Event broadcaster started (poll interval: %v)", b.pollInterval)
}

// Stop gracefully stops the broadcaster.
func (b *Broadcaster) Stop() {
	close(b.stopCh)
	b.wg.Wait()
	log.Println("Event broadcaster stopped")
}

// pollLoop periodically checks for status changes.
func (b *Broadcaster) pollLoop() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.pollInterval)
	defer ticker.Stop()

	// Initial poll
	b.checkForChanges()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			b.checkForChanges()
		}
	}
}

// checkForChanges polls current status and broadcasts any changes.
func (b *Broadcaster) checkForChanges() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get current agent states
	currentAgents := b.getAgentSnapshots()

	// Compare and broadcast agent changes
	for addr, curr := range currentAgents {
		prev, existed := b.prevAgents[addr]
		if !existed {
			// New agent appeared
			b.broadcastAgentUpdate(addr, "connected", curr)
		} else if curr != prev {
			// State changed
			changeType := b.detectChangeType(prev, curr)
			b.broadcastAgentUpdate(addr, changeType, curr)
		}
	}

	// Check for agents that disappeared
	for addr, prev := range b.prevAgents {
		if _, exists := currentAgents[addr]; !exists {
			b.broadcastAgentUpdate(addr, "disconnected", prev)
		}
	}

	// Update previous state
	b.prevAgents = currentAgents
}

// getAgentSnapshots fetches current state of all agents.
func (b *Broadcaster) getAgentSnapshots() map[string]agentSnapshot {
	snapshots := make(map[string]agentSnapshot)

	// Load configs
	rigsConfigPath := constants.MayorRigsPath(b.townRoot)
	rigsConfig, err := config.LoadRigsConfig(rigsConfigPath)
	if err != nil {
		return snapshots
	}

	g := git.NewGit(b.townRoot)
	mgr := rig.NewManager(b.townRoot, rigsConfig, g)
	t := tmux.NewTmux()

	// Get running sessions
	sessions := make(map[string]bool)
	if sessionList, err := t.ListSessions(); err == nil {
		for _, s := range sessionList {
			sessions[s] = true
		}
	}

	// Discover rigs
	rigs, err := mgr.DiscoverRigs()
	if err != nil {
		return snapshots
	}

	// Pre-fetch agent and hook beads
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

	// Check global agents (Mayor, Deacon)
	for _, name := range []string{"mayor", "deacon"} {
		session := "gt-" + name
		if sessions[session] {
			snapshots[name] = agentSnapshot{
				Running: true,
			}
		}
	}

	// Process rig agents
	for _, r := range rigs {
		// Polecats
		for _, name := range r.Polecats {
			session := "gt-" + r.Name + "-" + name
			address := r.Name + "/" + name
			snap := agentSnapshot{
				Running: sessions[session],
			}
			if bead, ok := allAgentBeads[address]; ok {
				hookID := bead.HookBead
				if hookID == "" {
					if fields := beads.ParseAgentFields(bead.Description); fields != nil {
						hookID = fields.HookBead
					}
				}
				if hookID != "" {
					snap.HasWork = true
					snap.HookBead = hookID
					if hookBead, ok := allHookBeads[hookID]; ok {
						snap.WorkTitle = hookBead.Title
					}
				}
				snap.State = bead.AgentState
			}
			snapshots[address] = snap
		}

		// Crew
		for _, name := range r.Crew {
			session := "gt-" + r.Name + "-" + name
			address := r.Name + "/" + name
			snap := agentSnapshot{
				Running: sessions[session],
			}
			if bead, ok := allAgentBeads[address]; ok {
				hookID := bead.HookBead
				if hookID == "" {
					if fields := beads.ParseAgentFields(bead.Description); fields != nil {
						hookID = fields.HookBead
					}
				}
				if hookID != "" {
					snap.HasWork = true
					snap.HookBead = hookID
					if hookBead, ok := allHookBeads[hookID]; ok {
						snap.WorkTitle = hookBead.Title
					}
				}
				snap.State = bead.AgentState
			}
			snapshots[address] = snap
		}

		// Witness
		if r.HasWitness {
			session := "gt-" + r.Name + "-witness"
			address := r.Name + "/witness"
			snapshots[address] = agentSnapshot{
				Running: sessions[session],
			}
		}

		// Refinery
		if r.HasRefinery {
			session := "gt-" + r.Name + "-refinery"
			address := r.Name + "/refinery"
			snapshots[address] = agentSnapshot{
				Running: sessions[session],
			}
		}
	}

	return snapshots
}

// detectChangeType determines what kind of change occurred.
func (b *Broadcaster) detectChangeType(prev, curr agentSnapshot) string {
	if prev.Running != curr.Running {
		if curr.Running {
			return "started"
		}
		return "stopped"
	}
	if prev.HookBead != curr.HookBead {
		if curr.HookBead != "" {
			return "work_assigned"
		}
		return "work_completed"
	}
	if prev.State != curr.State {
		return "state_changed"
	}
	return "updated"
}

// broadcastAgentUpdate sends an agent update message.
func (b *Broadcaster) broadcastAgentUpdate(address, changeType string, snap agentSnapshot) {
	msg := &api.WSMessage{
		Type:      api.WSTypeAgentUpdate,
		Timestamp: time.Now(),
		Payload: handlers.AgentRuntime{
			Address:   address,
			Running:   snap.Running,
			HasWork:   snap.HasWork,
			HookBead:  snap.HookBead,
			WorkTitle: snap.WorkTitle,
			State:     snap.State,
		},
	}

	// Add change info to payload
	type agentUpdatePayload struct {
		handlers.AgentRuntime
		ChangeType string `json:"change_type"`
	}

	msg.Payload = agentUpdatePayload{
		AgentRuntime: handlers.AgentRuntime{
			Address:   address,
			Running:   snap.Running,
			HasWork:   snap.HasWork,
			HookBead:  snap.HookBead,
			WorkTitle: snap.WorkTitle,
			State:     snap.State,
		},
		ChangeType: changeType,
	}

	b.hub.Broadcast(msg)
	log.Printf("Broadcast: agent %s %s", address, changeType)
}
