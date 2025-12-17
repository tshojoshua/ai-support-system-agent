package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/tshojoshua/jtnt-agent/internal/config"
	"github.com/tshojoshua/jtnt-agent/internal/jobs"
	"github.com/tshojoshua/jtnt-agent/internal/policy"
	"github.com/tshojoshua/jtnt-agent/internal/store"
	"github.com/tshojoshua/jtnt-agent/internal/sysinfo"
	"github.com/tshojoshua/jtnt-agent/internal/transport"
)

// Agent is the main agent orchestrator
type Agent struct {
	config      *config.Config
	client      *transport.Client
	store       store.Store
	sysinfo     *sysinfo.Collector
	logger      *Logger
	jobExecutor *jobs.Executor
	resultCache *ResultCache
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// New creates a new agent instance
func New(cfg *config.Config) (*Agent, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create mTLS client
	client, err := transport.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Create store
	store, err := store.NewStore(config.GetCertsDir())
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	// Create system info collector
	collector := sysinfo.NewCollector()

	// Create logger
	logger := NewLogger(cfg.AgentID, LogLevelInfo)

	// Create result cache
	resultCache, err := NewResultCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create result cache: %w", err)
	}

	// Load or create default policy
	pol := policy.DefaultPolicy()
	// TODO: Load policy from hub or local cache if available

	// Create policy enforcer
	enforcer := policy.NewEnforcer(pol)

	// Load hub's public key for script signature verification
	// TODO: Load from secure storage or configuration
	var hubPublicKey []byte // This should be loaded from config

	// Create job executor
	jobExecutor := jobs.NewExecutor(cfg.AgentID, enforcer, client, hubPublicKey, logger)

	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		config:      cfg,
		client:      client,
		store:       store,
		sysinfo:     collector,
		logger:      logger,
		jobExecutor: jobExecutor,
		resultCache: resultCache,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start starts the agent
func (a *Agent) Start() error {
	a.logger.Info("agent", map[string]interface{}{
		"message":  "starting agent",
		"agent_id": a.config.AgentID,
	})

	// Start heartbeat loop
	a.wg.Add(1)
	go a.heartbeatLoop()

	// Start job polling loop
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.jobPollLoop(a.ctx)
	}()

	a.logger.Info("agent", map[string]interface{}{
		"message": "agent started successfully",
	})

	return nil
}

// Stop gracefully stops the agent
func (a *Agent) Stop() error {
	a.logger.Info("agent", map[string]interface{}{
		"message": "stopping agent",
	})

	// Cancel context to stop all goroutines
	a.cancel()

	// Wait for all goroutines to finish
	a.wg.Wait()

	a.logger.Info("agent", map[string]interface{}{
		"message": "agent stopped successfully",
	})

	return nil
}

// Wait blocks until the agent is stopped
func (a *Agent) Wait() {
	a.wg.Wait()
}
