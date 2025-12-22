package agent

import (
	"context"
	"log"
	"sync"

	"github.com/jtnt/agent/internal/config"
	"github.com/jtnt/agent/internal/transport"
)

type Agent struct {
	config *config.Config
	client *transport.Client
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(cfg *config.Config) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	client := transport.NewClient(cfg.HubURL)
	client.SetAgentToken(cfg.AgentToken)

	return &Agent{
		config: cfg,
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (a *Agent) Start() error {
	log.Printf("Starting JTNT Agent v%s", config.Version)
	log.Printf("Agent ID: %s", a.config.AgentID)
	log.Printf("Hub URL: %s", a.config.HubURL)

	// Start heartbeat loop
	a.wg.Add(1)
	go a.heartbeatLoop()

	// TODO: Start job polling loop (Phase 2)
	// a.wg.Add(1)
	// go a.jobPollLoop()

	log.Println("Agent started successfully")
	return nil
}

func (a *Agent) Stop() error {
	log.Println("Stopping agent...")
	a.cancel()
	a.wg.Wait()
	log.Println("Agent stopped")
	return nil
}

func (a *Agent) Wait() {
	<-a.ctx.Done()
}
