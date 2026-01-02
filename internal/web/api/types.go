// Package api provides shared types for the web API.
package api

import "time"

// Response wraps all API responses with consistent structure.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo provides structured error information.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaginatedResponse wraps paginated list responses.
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
	HasMore    bool        `json:"has_more"`
}

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// WSMessage types
const (
	WSTypeEvent        = "event"
	WSTypeStatusUpdate = "status_update"
	WSTypeAgentUpdate  = "agent_update"
	WSTypeConvoyUpdate = "convoy_update"
	WSTypeMQUpdate     = "mq_update"
	WSTypePing         = "ping"
	WSTypePong         = "pong"
)

// SubscribeMessage is sent by clients to subscribe to topics.
type SubscribeMessage struct {
	Topics []string `json:"topics"`
}

// Topics clients can subscribe to
const (
	TopicAll     = "all"
	TopicEvents  = "events"
	TopicStatus  = "status"
	TopicAgents  = "agents"
	TopicConvoys = "convoys"
	TopicMQ      = "mq"
)
