package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tshojoshua/jtnt-agent/internal/agent"
	"github.com/tshojoshua/jtnt-agent/internal/config"
)

const version = "1.0.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create agent
	agentInstance, err := agent.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Start agent
	if err := agentInstance.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	fmt.Printf("Received signal: %v\n", sig)

	// Gracefully stop agent
	if err := agentInstance.Stop(); err != nil {
		return fmt.Errorf("failed to stop agent: %w", err)
	}

	return nil
}
