// Package web provides the HTTP server for the Gas Town web dashboard.
package web

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/steveyegge/gastown/internal/web/api"
	"github.com/steveyegge/gastown/internal/web/handlers"
	"github.com/steveyegge/gastown/internal/web/ws"
)

//go:embed dist/*
var distFS embed.FS

// Config holds server configuration.
type Config struct {
	Host     string
	Port     int
	DevMode  bool
	TownRoot string
}

// Server represents the web dashboard HTTP server.
type Server struct {
	config     Config
	router     *mux.Router
	httpServer *http.Server
	hub        *ws.Hub
}

// NewServer creates a new web dashboard server.
func NewServer(cfg Config) *Server {
	s := &Server{
		config: cfg,
		router: mux.NewRouter(),
		hub:    ws.NewHub(),
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Status handlers
	statusHandler := handlers.NewStatusHandler(s.config.TownRoot)
	api.HandleFunc("/status", statusHandler.GetStatus).Methods("GET")
	api.HandleFunc("/status/summary", statusHandler.GetSummary).Methods("GET")

	// Rigs handlers
	rigsHandler := handlers.NewRigsHandler(s.config.TownRoot)
	api.HandleFunc("/rigs", rigsHandler.List).Methods("GET")
	api.HandleFunc("/rigs/{name}", rigsHandler.Get).Methods("GET")
	api.HandleFunc("/rigs/{name}/agents", rigsHandler.GetAgents).Methods("GET")
	api.HandleFunc("/rigs/{rig}/refinery/{action}", rigsHandler.RefinaryAction).Methods("POST")

	// Polecats handlers
	polecatsHandler := handlers.NewPolecatsHandler(s.config.TownRoot)
	api.HandleFunc("/rigs/{rig}/polecats", polecatsHandler.ListPolecats).Methods("GET")
	api.HandleFunc("/rigs/{rig}/polecats", polecatsHandler.HandlePolecats).Methods("POST")
	api.HandleFunc("/rigs/{rig}/polecats/{name}", polecatsHandler.RemovePolecat).Methods("DELETE")

	// Agents handlers
	agentsHandler := handlers.NewAgentsHandler(s.config.TownRoot)
	api.HandleFunc("/agents", agentsHandler.List).Methods("GET")
	api.HandleFunc("/agents/{address:.*}", agentsHandler.Get).Methods("GET")

	// Convoys handlers
	convoysHandler := handlers.NewConvoysHandler(s.config.TownRoot)
	api.HandleFunc("/convoys", convoysHandler.List).Methods("GET")
	api.HandleFunc("/convoys", convoysHandler.Create).Methods("POST")
	api.HandleFunc("/convoys/{id}", convoysHandler.Get).Methods("GET")

	// Events handlers
	eventsHandler := handlers.NewEventsHandler(s.config.TownRoot)
	api.HandleFunc("/events", eventsHandler.List).Methods("GET")

	// Mail handlers
	mailHandler := handlers.NewMailHandler(s.config.TownRoot)
	api.HandleFunc("/mail", mailHandler.ListInbox).Methods("GET")
	api.HandleFunc("/mail", mailHandler.HandleSendMail).Methods("POST")
	api.HandleFunc("/mail/count", mailHandler.GetCount).Methods("GET")
	api.HandleFunc("/mail/search", mailHandler.Search).Methods("POST")
	api.HandleFunc("/mail/{id}", mailHandler.GetMessage).Methods("GET")
	api.HandleFunc("/mail/{id}", mailHandler.DeleteMessage).Methods("DELETE")
	api.HandleFunc("/mail/{id}/read", mailHandler.MarkRead).Methods("POST")
	api.HandleFunc("/mail/{id}/reply", mailHandler.Reply).Methods("POST")
	api.HandleFunc("/mail/agent/{address:.*}", mailHandler.ListAgentInbox).Methods("GET")

	// Sling handlers (work dispatch)
	slingHandler := handlers.NewSlingHandler(s.config.TownRoot)
	api.HandleFunc("/sling", slingHandler.Dispatch).Methods("POST")

	// MQ handlers (merge queue)
	mqHandler := handlers.NewMQHandler(s.config.TownRoot)
	api.HandleFunc("/mq", mqHandler.List).Methods("GET")

	// Services handlers
	servicesHandler := handlers.NewServicesHandler(s.config.TownRoot)
	api.HandleFunc("/rigs/{rig}/services/witness/{action}", servicesHandler.HandleWitness).Methods("POST")
	api.HandleFunc("/rigs/{rig}/services/refinery/{action}", servicesHandler.HandleRefinery).Methods("POST")

	// Merge queue retry handler
	api.HandleFunc("/mq/{id}/retry", mqHandler.Retry).Methods("POST")

	// WebSocket handler
	api.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(s.hub, w, r)
	})

	// CORS middleware for dev mode
	if s.config.DevMode {
		s.router.Use(corsMiddleware)
	}

	// Serve static files from embedded filesystem
	s.router.PathPrefix("/").Handler(s.staticHandler())
}

// staticHandler returns an HTTP handler for static files.
func (s *Server) staticHandler() http.Handler {
	// Get the dist subdirectory
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Printf("Warning: could not load embedded assets: %v", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Dashboard assets not found. Run 'npm run build' in web/ directory.", http.StatusNotFound)
		})
	}

	fileServer := http.FileServer(http.FS(subFS))

	// Wrap to serve index.html for SPA routing
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		if _, err := fs.Stat(subFS, path[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fall back to index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers for development mode.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start WebSocket hub
	go s.hub.Run()

	// Start event broadcaster if we have a town root
	if s.config.TownRoot != "" {
		go s.startEventBroadcaster()
	}

	fmt.Printf("Starting Gas Town dashboard at http://%s\n", addr)
	return s.httpServer.ListenAndServe()
}

// StartWithGracefulShutdown starts the server and handles graceful shutdown.
func (s *Server) StartWithGracefulShutdown() error {
	// Channel to listen for errors from server
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		serverErrors <- s.Start()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		fmt.Printf("\nReceived %v, shutting down...\n", sig)

		// Give outstanding requests 5 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			s.httpServer.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

// startEventBroadcaster watches for events and broadcasts them to WebSocket clients.
func (s *Server) startEventBroadcaster() {
	eventsFile := filepath.Join(s.config.TownRoot, ".beads", "events.jsonl")

	// Track the current file size and position for tailing
	var lastSize int64
	var file *os.File
	var scanner *bufio.Scanner

	// Polling interval for checking file changes
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Try to open/reopen the file
		currentFile, err := os.Open(eventsFile)
		if err != nil {
			// File doesn't exist yet or can't be opened
			if file != nil {
				file.Close()
				file = nil
				scanner = nil
			}
			continue
		}

		// Get current file info
		fileInfo, err := currentFile.Stat()
		if err != nil {
			currentFile.Close()
			continue
		}

		currentSize := fileInfo.Size()

		// If this is a new file or file was truncated, reset position
		if file == nil || currentSize < lastSize {
			if file != nil {
				file.Close()
			}
			file = currentFile
			lastSize = 0
			scanner = bufio.NewScanner(file)
		} else if currentSize > lastSize {
			// File has new content, seek to last known position
			if file != nil {
				file.Close()
			}
			file = currentFile

			// Seek to last known position
			if _, err := file.Seek(lastSize, 0); err != nil {
				file.Close()
				file = nil
				continue
			}
			scanner = bufio.NewScanner(file)
		} else {
			// No new content
			currentFile.Close()
			continue
		}

		// Read new lines from the file
		if scanner != nil {
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}

				// Parse the JSON event
				var eventData map[string]interface{}
				if err := json.Unmarshal([]byte(line), &eventData); err != nil {
					log.Printf("Failed to parse event: %v", err)
					continue
				}

				// Create WSMessage to broadcast
				msg := &api.WSMessage{
					Type:      api.WSTypeEvent,
					Timestamp: time.Now(),
					Payload:   eventData,
				}

				// Broadcast to WebSocket clients
				s.hub.Broadcast(msg)

				// Update position
				lastSize, _ = file.Seek(0, 1) // Get current position
			}
		}

		if file != nil {
			lastSize, _ = file.Seek(0, 1) // Get final position
		}
	}

	// Cleanup
	if file != nil {
		file.Close()
	}
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}
