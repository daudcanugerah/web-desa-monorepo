package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server represents the HTTP server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new HTTP server with configured timeouts.
// Read timeout: 30s, Write timeout: 30s, Idle timeout: 60s
//
// Validates: Requirements 16.3, 24.1
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start starts the HTTP server and handles graceful shutdown.
// It listens for SIGINT and SIGTERM signals and performs graceful shutdown
// with a 30-second timeout.
//
// Validates: Requirements 16.3, 24.1
func (s *Server) Start() error {
	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Starting HTTP server on %s\n", s.httpServer.Addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		fmt.Printf("\nReceived signal: %v. Starting graceful shutdown...\n", sig)

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := s.httpServer.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			s.httpServer.Close()
			return fmt.Errorf("failed to gracefully shutdown server: %w", err)
		}

		fmt.Println("Server shutdown completed")
	}

	return nil
}

// Shutdown gracefully shuts down the server with the given context.
// This is useful for testing or programmatic shutdown.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
