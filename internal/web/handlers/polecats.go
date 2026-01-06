package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/git"
	"github.com/steveyegge/gastown/internal/rig"
	"github.com/steveyegge/gastown/internal/web/api"
)

// PolecatsHandler handles polecat-related API requests.
type PolecatsHandler struct {
	townRoot string
}

// AddPolecatRequest represents a request to add a polecat.
type AddPolecatRequest struct {
	Name string `json:"name"`
}

// PolecatData represents polecat information in responses.
type PolecatData struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// AddPolecatResponse represents the response to an add polecat request.
type AddPolecatResponse struct {
	Success bool        `json:"success"`
	Data    PolecatData `json:"data"`
}

// RemovalResponse represents the result of a polecat removal operation.
type RemovalResponse struct {
	Success bool   `json:"success"`
	Rig     string `json:"rig"`
	Polecat string `json:"polecat"`
	Message string `json:"message"`
	Removed bool   `json:"removed"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewPolecatsHandler creates a new polecats handler.
func NewPolecatsHandler(townRoot string) *PolecatsHandler {
	return &PolecatsHandler{townRoot: townRoot}
}

// HandlePolecats routes POST requests to add and DELETE requests to remove polecats.
func (h *PolecatsHandler) HandlePolecats(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.AddPolecat(w, r)
	case http.MethodDelete:
		h.RemovePolecat(w, r)
	default:
		api.BadRequest(w, "method not allowed")
	}
}

// AddPolecat handles POST requests to add a polecat.
func (h *PolecatsHandler) AddPolecat(w http.ResponseWriter, r *http.Request) {
	var req AddPolecatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		api.BadRequest(w, "name is required")
		return
	}

	vars := mux.Vars(r)
	rigName := vars["rig"]

	// Construct the rig directory path
	rigPath := filepath.Join(h.townRoot, rigName)

	// Verify the rig directory exists
	if _, err := os.Stat(rigPath); err != nil {
		api.BadRequest(w, "rig not found: "+rigName)
		return
	}

	// Execute 'gt polecat add <name>' in the rig directory
	ctx := r.Context()
	cmd := exec.CommandContext(ctx, "gt", "polecat", "add", req.Name)
	cmd.Dir = rigPath

	// Capture stderr and stdout
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Wait for command with timeout
	select {
	case err := <-done:
		if err != nil {
			// Command failed
			errMsg := strings.TrimSpace(stderr.String())
			if errMsg == "" {
				errMsg = err.Error()
			}
			api.BadRequest(w, "failed to create polecat: "+errMsg)
			return
		}
	case <-time.After(30 * time.Second):
		api.InternalError(w, "polecat creation timeout")
		return
	}

	// Parse the output to extract the polecat address
	// The output format is expected to contain the polecat path
	output := strings.TrimSpace(stdout.String())
	polecatAddress := rigName + "/polecats/" + req.Name

	// Try to extract a more specific address from the output if available
	// For now, use the standard format
	lines := strings.Split(output, "\n")
	if len(lines) > 0 && lines[0] != "" {
		// Check if output contains the address format
		for _, line := range lines {
			if strings.Contains(line, "/polecats/") {
				polecatAddress = strings.TrimSpace(line)
				break
			}
		}
	}

	// Return success with the created polecat address
	data := PolecatData{
		Name:    req.Name,
		Address: polecatAddress,
	}

	response := AddPolecatResponse{
		Success: true,
		Data:    data,
	}

	api.WriteJSON(w, response)
}

// RemovePolecat handles DELETE requests to remove a specific polecat.
// Endpoint: DELETE /api/v1/rigs/{rig}/polecats/{name}
func (h *PolecatsHandler) RemovePolecat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rigName := vars["rig"]
	polecatName := vars["name"]

	// Validate parameters
	if rigName == "" {
		api.BadRequest(w, "rig name is required")
		return
	}
	if polecatName == "" {
		api.BadRequest(w, "polecat name is required")
		return
	}

	// Build the path to the polecat directory
	polecatPath := filepath.Join(h.townRoot, rigName, "polecats", polecatName)

	// Check if polecat exists
	info, err := os.Stat(polecatPath)
	if err != nil {
		if os.IsNotExist(err) {
			api.NotFound(w, fmt.Sprintf("polecat not found: %s/%s", rigName, polecatName))
			return
		}
		api.InternalError(w, fmt.Sprintf("failed to stat polecat: %v", err))
		return
	}

	// Ensure it's a directory
	if !info.IsDir() {
		api.BadRequest(w, fmt.Sprintf("polecat path is not a directory: %s", polecatPath))
		return
	}

	// Attempt to remove the polecat directory
	if err := os.RemoveAll(polecatPath); err != nil {
		api.InternalError(w, fmt.Sprintf("failed to remove polecat: %v", err))
		return
	}

	// Return success response
	resp := RemovalResponse{
		Success: true,
		Rig:     rigName,
		Polecat: polecatName,
		Message: fmt.Sprintf("Polecat %s removed successfully from rig %s", polecatName, rigName),
		Removed: true,
		Path:    polecatPath,
	}
	api.WriteJSON(w, resp)
}

// ListPolecats handles GET requests to list all polecats in a rig.
// Endpoint: GET /api/v1/rigs/{rig}/polecats
func (h *PolecatsHandler) ListPolecats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rigName := vars["rig"]

	if rigName == "" {
		api.BadRequest(w, "rig name is required")
		return
	}

	// Create a rig manager to discover polecats
	g := git.NewGit(h.townRoot)
	mgr := rig.NewManager(h.townRoot, nil, g)

	// Discover the specific rig
	rigs, err := mgr.DiscoverRigs()
	if err != nil {
		api.InternalError(w, fmt.Sprintf("failed to discover rigs: %v", err))
		return
	}

	// Find the requested rig
	for _, r := range rigs {
		if r.Name == rigName {
			resp := map[string]interface{}{
				"rig":      rigName,
				"polecats": r.Polecats,
				"count":    len(r.Polecats),
			}
			api.WriteJSON(w, resp)
			return
		}
	}

	api.NotFound(w, fmt.Sprintf("rig not found: %s", rigName))
}
