package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jtnt/agent/internal/agent"
	"github.com/jtnt/agent/internal/config"
)

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[jtnt-agentd] ")

	// Load configuration
	if !config.Exists() {
		fmt.Fprintf(os.Stderr, "Error: Agent not enrolled\n")
		fmt.Fprintf(os.Stderr, "Run: jtnt-agent enroll --token <TOKEN> --hub https://hub.jtnt.us\n")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create and start agent
	agt := agent.New(cfg)
	if err := agt.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting agent: %v\n", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	// Graceful shutdown
	if err := agt.Stop(); err != nil {
		log.Printf("Error stopping agent: %v", err)
		os.Exit(1)
	}

	log.Println("Agent shutdown complete")
}
