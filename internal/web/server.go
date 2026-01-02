// Package web provides the HTTP server for the Gas Town web dashboard.
package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
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

	// Agents handlers
	agentsHandler := handlers.NewAgentsHandler(s.config.TownRoot)
	api.HandleFunc("/agents", agentsHandler.List).Methods("GET")
	api.HandleFunc("/agents/{address:.*}", agentsHandler.Get).Methods("GET")

	// Convoys handlers
	convoysHandler := handlers.NewConvoysHandler(s.config.TownRoot)
	api.HandleFunc("/convoys", convoysHandler.List).Methods("GET")
	api.HandleFunc("/convoys/{id}", convoysHandler.Get).Methods("GET")

	// Events handlers
	eventsHandler := handlers.NewEventsHandler(s.config.TownRoot)
	api.HandleFunc("/events", eventsHandler.List).Methods("GET")

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
	// TODO: Implement event file tailing and broadcasting
	// This will be implemented in Phase 2: Real-time Events
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}
