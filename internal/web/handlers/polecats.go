package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
	Success bool         `json:"success"`
	Data    PolecatData  `json:"data"`
}

// RemovePolecatResponse represents the response to a remove polecat request.
type RemovePolecatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
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
	rig := vars["rig"]

	// TODO: Implement actual polecat creation logic
	// For now, return a mock response with the address format "rig/polecats/name"
	data := PolecatData{
		Name:    req.Name,
		Address: rig + "/polecats/" + req.Name,
	}

	response := AddPolecatResponse{
		Success: true,
		Data:    data,
	}

	api.WriteJSON(w, response)
}

// RemovePolecat handles DELETE requests to remove a polecat.
func (h *PolecatsHandler) RemovePolecat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rig := vars["rig"]
	name := vars["name"]

	if name == "" {
		api.BadRequest(w, "polecat name is required")
		return
	}

	// TODO: Implement actual polecat removal logic

	response := RemovePolecatResponse{
		Success: true,
		Message: "polecat " + rig + "/polecats/" + name + " removed",
	}

	api.WriteJSON(w, response)
}
