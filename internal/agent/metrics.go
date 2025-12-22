package agent

import (
	"context"
	"net/http"
	"time"
)

// MetricsServer represents a metrics server (placeholder)
type MetricsServer struct {
	server *http.Server
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(addr string) *MetricsServer {
	return &MetricsServer{
		server: &http.Server{
			Addr: addr,
		},
	}
}

// Start starts the metrics server
func (m *MetricsServer) Start() error {
	if m.server == nil {
		return nil
	}
	go m.server.ListenAndServe()
	return nil
}

// Stop stops the metrics server
func (m *MetricsServer) Stop(timeout time.Duration) error {
	if m.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.server.Shutdown(ctx)
}
