// Package handlers provides HTTP request handlers for the Gas Town web dashboard.
package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/mail"
	"github.com/steveyegge/gastown/internal/web/api"
)

// MailHandler handles mail-related HTTP requests.
type MailHandler struct {
	townRoot string
}

// NewMailHandler creates a new mail handler.
func NewMailHandler(townRoot string) *MailHandler {
	return &MailHandler{
		townRoot: townRoot,
	}
}

// MailMessage represents a mail message for the API.
type MailMessage struct {
	ID        string   `json:"id"`
	From      string   `json:"from"`
	To        string   `json:"to"`
	Subject   string   `json:"subject"`
	Body      string   `json:"body"`
	Timestamp string   `json:"timestamp"`
	Read      bool     `json:"read"`
	Priority  string   `json:"priority"`
	Type      string   `json:"type"`
	ThreadID  string   `json:"thread_id,omitempty"`
	ReplyTo   string   `json:"reply_to,omitempty"`
	Pinned    bool     `json:"pinned,omitempty"`
	CC        []string `json:"cc,omitempty"`
}

// convertMessage converts a mail.Message to MailMessage for API response.
func convertMessage(msg *mail.Message) MailMessage {
	return MailMessage{
		ID:        msg.ID,
		From:      msg.From,
		To:        msg.To,
		Subject:   msg.Subject,
		Body:      msg.Body,
		Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Read:      msg.Read,
		Priority:  string(msg.Priority),
		Type:      string(msg.Type),
		ThreadID:  msg.ThreadID,
		ReplyTo:   msg.ReplyTo,
		Pinned:    msg.Pinned,
		CC:        msg.CC,
	}
}

// ListInbox lists messages for the mayor's inbox.
func (h *MailHandler) ListInbox(w http.ResponseWriter, r *http.Request) {
	// Mayor mailbox uses town .beads directory
	beadsDir := filepath.Join(h.townRoot, ".beads")
	mailbox := mail.NewMailboxWithBeadsDir("mayor/", h.townRoot, beadsDir)

	messages, err := mailbox.List()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "MAIL_ERROR", "Failed to list mail: "+err.Error())
		return
	}

	var result []MailMessage
	for _, msg := range messages {
		result = append(result, convertMessage(msg))
	}

	api.WriteJSON(w, result)
}

// ListAgentInbox lists messages for a specific agent.
func (h *MailHandler) ListAgentInbox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" {
		api.WriteError(w, http.StatusBadRequest, "INVALID_ADDRESS", "Agent address is required")
		return
	}

	// Resolve beads directory based on address
	var beadsDir string
	var identity string

	// Town-level agents
	if address == "mayor" || address == "mayor/" {
		beadsDir = filepath.Join(h.townRoot, ".beads")
		identity = "mayor/"
	} else if address == "deacon" || address == "deacon/" {
		beadsDir = filepath.Join(h.townRoot, ".beads")
		identity = "deacon/"
	} else {
		// Rig-level agents - parse rig name from address
		// Format: rig/agent or rig/polecats/agent or rig/crew/agent
		api.WriteError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Agent-specific inbox not yet implemented")
		return
	}

	mailbox := mail.NewMailboxWithBeadsDir(identity, h.townRoot, beadsDir)

	messages, err := mailbox.List()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "MAIL_ERROR", "Failed to list mail: "+err.Error())
		return
	}

	var result []MailMessage
	for _, msg := range messages {
		result = append(result, convertMessage(msg))
	}

	api.WriteJSON(w, result)
}

// GetMessage gets a single message by ID.
func (h *MailHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Message ID is required")
		return
	}

	// Use mayor's mailbox for now
	beadsDir := filepath.Join(h.townRoot, ".beads")
	mailbox := mail.NewMailboxWithBeadsDir("mayor/", h.townRoot, beadsDir)

	msg, err := mailbox.Get(id)
	if err != nil {
		if err == mail.ErrMessageNotFound {
			api.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Message not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "MAIL_ERROR", "Failed to get message: "+err.Error())
		return
	}

	api.WriteJSON(w, convertMessage(msg))
}

// MarkRead marks a message as read.
func (h *MailHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Message ID is required")
		return
	}

	beadsDir := filepath.Join(h.townRoot, ".beads")
	mailbox := mail.NewMailboxWithBeadsDir("mayor/", h.townRoot, beadsDir)

	if err := mailbox.MarkRead(id); err != nil {
		if err == mail.ErrMessageNotFound {
			api.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Message not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "MAIL_ERROR", "Failed to mark as read: "+err.Error())
		return
	}

	api.WriteJSON(w, map[string]bool{"success": true})
}

// MailCount returns the count of messages.
type MailCount struct {
	Total  int `json:"total"`
	Unread int `json:"unread"`
}

// GetCount returns message counts.
func (h *MailHandler) GetCount(w http.ResponseWriter, r *http.Request) {
	beadsDir := filepath.Join(h.townRoot, ".beads")
	mailbox := mail.NewMailboxWithBeadsDir("mayor/", h.townRoot, beadsDir)

	total, unread, err := mailbox.Count()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "MAIL_ERROR", "Failed to count mail: "+err.Error())
		return
	}

	api.WriteJSON(w, MailCount{Total: total, Unread: unread})
}

// SearchRequest represents a mail search request.
type SearchRequest struct {
	Query       string `json:"query"`
	FromFilter  string `json:"from,omitempty"`
	SubjectOnly bool   `json:"subject_only,omitempty"`
	BodyOnly    bool   `json:"body_only,omitempty"`
}

// Search searches for messages matching criteria.
func (h *MailHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid search request")
		return
	}

	if req.Query == "" {
		api.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", "Search query is required")
		return
	}

	beadsDir := filepath.Join(h.townRoot, ".beads")
	mailbox := mail.NewMailboxWithBeadsDir("mayor/", h.townRoot, beadsDir)

	messages, err := mailbox.Search(mail.SearchOptions{
		Query:       req.Query,
		FromFilter:  req.FromFilter,
		SubjectOnly: req.SubjectOnly,
		BodyOnly:    req.BodyOnly,
	})
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "SEARCH_ERROR", "Search failed: "+err.Error())
		return
	}

	var result []MailMessage
	for _, msg := range messages {
		result = append(result, convertMessage(msg))
	}

	api.WriteJSON(w, result)
}
