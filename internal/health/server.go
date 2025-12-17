package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultHealthAddr  = "127.0.0.1:9091"
	healthReadTimeout  = 5 * time.Second
	healthWriteTimeout = 5 * time.Second
)

// Server serves health checks on localhost only
type Server struct {
	addr    string
	checker *Checker
	server  *http.Server
}

// NewServer creates a new health check server
func NewServer(addr string, checker *Checker) *Server {
	if addr == "" {
		addr = defaultHealthAddr
	}

	return &Server{
		addr:    addr,
		checker: checker,
	}
}

// Start starts the health check server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  healthReadTimeout,
		WriteTimeout: healthWriteTimeout,
	}

	// Start server in background
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("health server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully stops the health server
func (s *Server) Stop(timeout time.Duration) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	report := s.checker.GetReport()

	statusCode := http.StatusOK
	if report.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.addr
}
