package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultMetricsAddr = "127.0.0.1:9090"
	serverReadTimeout  = 10 * time.Second
	serverWriteTimeout = 10 * time.Second
)

// Server serves Prometheus metrics on localhost only
type Server struct {
	addr   string
	server *http.Server
}

// NewServer creates a new metrics server
func NewServer(addr string) *Server {
	if addr == "" {
		addr = defaultMetricsAddr
	}

	return &Server{
		addr: addr,
	}
}

// Start starts the metrics server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(
		GetRegistry(),
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
	}

	// Start server in background
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't crash agent
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully stops the metrics server
func (s *Server) Stop(timeout time.Duration) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.addr
}
