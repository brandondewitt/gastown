// Package ws provides WebSocket functionality for real-time updates.
package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/steveyegge/gastown/internal/web/api"
)

// mustMarshal marshals v to JSON, panicking on error.
func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return []byte("{}")
	}
	return b
}

// unmarshal is a helper for json.Unmarshal.
func unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (dev mode)
		// TODO: Restrict in production
		return true
	},
}

// Client represents a WebSocket client connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	topics map[string]bool
	mu     sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *api.WSMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *api.WSMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to this topic
				if client.isSubscribed(message.Type) {
					select {
					case client.send <- mustMarshal(message):
					default:
						// Client buffer full, close connection
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all subscribed clients.
func (h *Hub) Broadcast(msg *api.WSMessage) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	select {
	case h.broadcast <- msg:
	default:
		log.Println("Warning: broadcast channel full, message dropped")
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// isSubscribed checks if client is subscribed to a topic.
func (c *Client) isSubscribed(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// If subscribed to "all", receive everything
	if c.topics[api.TopicAll] {
		return true
	}

	// Check specific topic
	return c.topics[topic]
}

// subscribe adds topics to client subscription.
func (c *Client) subscribe(topics []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, topic := range topics {
		c.topics[topic] = true
	}
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle subscription messages
		var subMsg api.SubscribeMessage
		if err := unmarshal(message, &subMsg); err == nil && len(subMsg.Topics) > 0 {
			c.subscribe(subMsg.Topics)
			log.Printf("Client subscribed to topics: %v", subMsg.Topics)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain any queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS handles WebSocket requests from clients.
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		topics: map[string]bool{api.TopicAll: true}, // Subscribe to all by default
	}

	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}
