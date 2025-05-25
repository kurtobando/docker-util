package web

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Server manages the HTTP web server
type Server struct {
	server  *http.Server
	handler *Handler
}

// NewServer creates a new web server
func NewServer(port string, handler *Handler) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ServeLogsPage)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return &Server{
		server:  server,
		handler: handler,
	}
}

// Start starts the web server
// This preserves the exact same server startup logic from the original main function
func (s *Server) Start(ctx context.Context) error {
	log.Printf("Starting web UI on http://localhost%s", s.server.Addr)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the web server
// This preserves the exact same shutdown logic from the original main function
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down web server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Give it 5 seconds to shutdown gracefully
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown error: %v", err)
		return err
	}

	return nil
}
