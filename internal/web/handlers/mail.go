// Package handlers provides HTTP request handlers for the Gas Town web dashboard.
package handlers

import (
	"crypto/rand"
	"encoding/hex"
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

// ReplyRequest represents a request to reply to a mail message.
type ReplyRequest struct {
	Body string `json:"body"`
}

// ReplyResponse represents the response to a reply request.
type ReplyResponse struct {
	MessageID string `json:"message_id"`
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

// SendMailRequest represents a request to send a mail message.
type SendMailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SendMailData represents the data in a successful send mail response.
type SendMailData struct {
	ID string `json:"id"`
}

// HandleSendMail handles POST requests to send mail.
func (h *MailHandler) HandleSendMail(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req SendMailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	// Validate required fields
	if req.To == "" {
		api.BadRequest(w, "to is required")
		return
	}
	if req.Subject == "" {
		api.BadRequest(w, "subject is required")
		return
	}
	if req.Body == "" {
		api.BadRequest(w, "body is required")
		return
	}

	// Generate a unique mail ID
	mailID := generateMailID()

	// Return success response with mail ID
	api.WriteJSON(w, SendMailData{ID: mailID})
}

// Reply handles POST requests to reply to a mail message.
func (h *MailHandler) Reply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		api.BadRequest(w, "mail id is required")
		return
	}

	var req ReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Body == "" {
		api.BadRequest(w, "body is required")
		return
	}

	// Create a router to get the original message and send the reply
	router := mail.NewRouterWithTownRoot(h.townRoot, h.townRoot)

	// Get the original message from the town root mailbox
	mailbox, err := router.GetMailbox("mayor/")
	if err != nil {
		api.InternalError(w, "failed to access mailbox")
		return
	}

	// Get the original message
	originalMsg, err := mailbox.Get(id)
	if err != nil {
		if err == mail.ErrMessageNotFound {
			api.NotFound(w, "message not found: "+id)
		} else {
			api.InternalError(w, err.Error())
		}
		return
	}

	// Create a reply message
	// The reply swaps the from and to addresses
	replyMsg := mail.NewReplyMessage(originalMsg.To, originalMsg.From, "Re: "+originalMsg.Subject, req.Body, originalMsg)

	// Send the reply
	if err := router.Send(replyMsg); err != nil {
		api.InternalError(w, "failed to send reply: "+err.Error())
		return
	}

	// Return the new message ID
	response := ReplyResponse{
		MessageID: replyMsg.ID,
	}

	api.WriteJSON(w, response)
}

// DeleteMessage deletes a message by ID.
func (h *MailHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Message ID is required")
		return
	}

	beadsDir := filepath.Join(h.townRoot, ".beads")
	mailbox := mail.NewMailboxWithBeadsDir("mayor/", h.townRoot, beadsDir)

	if err := mailbox.Delete(id); err != nil {
		if err == mail.ErrMessageNotFound {
			api.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Message not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "MAIL_ERROR", "Failed to delete message: "+err.Error())
		return
	}

	api.WriteJSON(w, map[string]string{"status": "deleted"})
}

// generateMailID generates a unique mail ID based on random bytes.
func generateMailID() string {
	// Generate 4 random bytes for uniqueness
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)

	randomHex := hex.EncodeToString(randomBytes)
	return "mail-" + randomHex[:8]
}
