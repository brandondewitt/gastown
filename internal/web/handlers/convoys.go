package handlers

import (
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/web/api"
)

// ConvoysHandler handles convoy-related API requests.
type ConvoysHandler struct {
	townRoot string
}

// NewConvoysHandler creates a new convoys handler.
func NewConvoysHandler(townRoot string) *ConvoysHandler {
	return &ConvoysHandler{townRoot: townRoot}
}

// ConvoyInfo represents a convoy for API responses.
type ConvoyInfo struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Status      string              `json:"status"` // open, closed
	TrackedIDs  []string            `json:"tracked_ids"`
	Progress    float64             `json:"progress"` // 0.0 - 1.0
	Completed   int                 `json:"completed"`
	Total       int                 `json:"total"`
	CreatedAt   string              `json:"created_at,omitempty"`
	CompletedAt string              `json:"completed_at,omitempty"`
	Members     []ConvoyMemberInfo  `json:"members,omitempty"`
	Blockers    []ConvoyBlockerInfo `json:"blockers,omitempty"`
}

// ConvoyMemberInfo represents an issue tracked by a convoy.
type ConvoyMemberInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	Assignee  string `json:"assignee,omitempty"`
	AgentName string `json:"agent_name,omitempty"` // Short name for display
}

// ConvoyBlockerInfo represents an issue blocking convoy progress.
type ConvoyBlockerInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	BlockedBy string `json:"blocked_by"`
	Reason    string `json:"reason,omitempty"`
}

// List returns all convoys.
func (h *ConvoysHandler) List(w http.ResponseWriter, r *http.Request) {
	// Query parameters
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "open" // Default to open convoys
	}

	// Get convoys from town-level beads
	townBeadsPath := filepath.Join(h.townRoot, ".beads")
	bd := beads.New(townBeadsPath)

	opts := beads.ListOptions{
		Type:     "convoy",
		Priority: -1, // No priority filter
	}
	if status != "all" {
		opts.Status = status
	}

	convoys, err := bd.List(opts)
	if err != nil {
		// If beads not initialized, return empty list
		if strings.Contains(err.Error(), "not a beads") {
			api.WriteJSON(w, []ConvoyInfo{})
			return
		}
		api.InternalError(w, "listing convoys: "+err.Error())
		return
	}

	// Build convoy info with progress
	result := make([]ConvoyInfo, 0, len(convoys))

	// Process convoys in parallel for better performance
	type convoyResult struct {
		idx  int
		info ConvoyInfo
	}
	results := make(chan convoyResult, len(convoys))
	var wg sync.WaitGroup

	for i, convoy := range convoys {
		wg.Add(1)
		go func(idx int, c *beads.Issue) {
			defer wg.Done()
			info := h.buildConvoyInfo(c, false) // Brief mode for list
			results <- convoyResult{idx: idx, info: info}
		}(i, convoy)
	}

	// Wait and close
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results maintaining order
	orderedResults := make([]ConvoyInfo, len(convoys))
	for r := range results {
		orderedResults[r.idx] = r.info
	}

	// Filter out empty results and build final list
	for _, info := range orderedResults {
		if info.ID != "" {
			result = append(result, info)
		}
	}

	api.WriteJSON(w, result)
}

// Get returns a single convoy by ID.
func (h *ConvoysHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get convoy from town-level beads
	townBeadsPath := filepath.Join(h.townRoot, ".beads")
	bd := beads.New(townBeadsPath)

	convoy, err := bd.Show(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.NotFound(w, "convoy not found: "+id)
			return
		}
		api.InternalError(w, "getting convoy: "+err.Error())
		return
	}

	// Verify it's a convoy type
	if convoy.Type != "convoy" {
		api.NotFound(w, "not a convoy: "+id)
		return
	}

	// Build detailed convoy info
	info := h.buildConvoyInfo(convoy, true) // Full mode for detail

	api.WriteJSON(w, info)
}

// buildConvoyInfo builds a ConvoyInfo from a beads issue.
// If detailed is true, includes member details and blockers.
func (h *ConvoysHandler) buildConvoyInfo(convoy *beads.Issue, detailed bool) ConvoyInfo {
	info := ConvoyInfo{
		ID:        convoy.ID,
		Name:      convoy.Title,
		Status:    convoy.Status,
		CreatedAt: convoy.CreatedAt,
	}

	if convoy.Status == "closed" && convoy.ClosedAt != "" {
		info.CompletedAt = convoy.ClosedAt
	}

	// Get tracked issues from dependencies
	// Convoys use 'tracks' dependency type stored in Dependencies
	var trackedIDs []string
	for _, dep := range convoy.Dependencies {
		if dep.DependencyType == "tracks" {
			// Handle external references: external:rig:issue-id
			issueID := dep.ID
			if strings.HasPrefix(issueID, "external:") {
				parts := strings.SplitN(issueID, ":", 3)
				if len(parts) == 3 {
					issueID = parts[2]
				}
			}
			trackedIDs = append(trackedIDs, issueID)
		}
	}

	// Also check DependsOn list (dependency IDs without full info)
	if len(trackedIDs) == 0 {
		trackedIDs = convoy.DependsOn
	}

	info.TrackedIDs = trackedIDs
	info.Total = len(trackedIDs)

	if info.Total == 0 {
		info.Progress = 1.0 // Empty convoy is complete
		return info
	}

	// Get status of tracked issues
	members, blockers, completed := h.getTrackedIssueDetails(trackedIDs, detailed)

	info.Completed = completed
	if info.Total > 0 {
		info.Progress = float64(completed) / float64(info.Total)
	}

	if detailed {
		info.Members = members
		info.Blockers = blockers
	}

	return info
}

// getTrackedIssueDetails fetches details about tracked issues.
// Returns members list, blockers list, and count of completed issues.
func (h *ConvoysHandler) getTrackedIssueDetails(issueIDs []string, detailed bool) ([]ConvoyMemberInfo, []ConvoyBlockerInfo, int) {
	if len(issueIDs) == 0 {
		return nil, nil, 0
	}

	// Group issues by prefix to route to correct beads DB
	prefixGroups := make(map[string][]string)
	for _, id := range issueIDs {
		prefix := extractPrefix(id)
		prefixGroups[prefix] = append(prefixGroups[prefix], id)
	}

	var members []ConvoyMemberInfo
	var blockers []ConvoyBlockerInfo
	completed := 0

	// For simplicity, use bd show which handles routing
	// This works because bd has prefix-based routing built in
	townBeadsPath := filepath.Join(h.townRoot, ".beads")
	bd := beads.New(townBeadsPath)

	// Try batch show for all IDs
	issueMap, _ := bd.ShowMultiple(issueIDs)

	for _, id := range issueIDs {
		issue, ok := issueMap[id]
		if !ok {
			// Issue not found in batch - might be in a different rig
			// Try individual lookup (bd routing will find it)
			issue, _ = bd.Show(id)
		}

		if issue == nil {
			// Still not found - add placeholder
			if detailed {
				members = append(members, ConvoyMemberInfo{
					ID:     id,
					Title:  "(not found)",
					Status: "unknown",
				})
			}
			continue
		}

		if issue.Status == "closed" {
			completed++
		}

		if detailed {
			member := ConvoyMemberInfo{
				ID:       issue.ID,
				Title:    issue.Title,
				Status:   issue.Status,
				Type:     issue.Type,
				Assignee: issue.Assignee,
			}

			// Extract short agent name from assignee path
			if issue.Assignee != "" {
				parts := strings.Split(issue.Assignee, "/")
				member.AgentName = parts[len(parts)-1]
			}

			members = append(members, member)

			// Check for blockers (issues that are blocked by dependencies)
			if len(issue.BlockedBy) > 0 && issue.Status != "closed" {
				for _, blocker := range issue.BlockedBy {
					blockers = append(blockers, ConvoyBlockerInfo{
						ID:        issue.ID,
						Title:     issue.Title,
						BlockedBy: blocker,
					})
				}
			}
		}
	}

	return members, blockers, completed
}

// extractPrefix extracts the prefix from an issue ID (e.g., "gt" from "gt-abc123").
func extractPrefix(id string) string {
	idx := strings.Index(id, "-")
	if idx > 0 && idx <= 3 {
		return id[:idx]
	}
	return ""
}
