package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/config"
	"github.com/tshojoshua/jtnt-agent/internal/enroll"
	"github.com/tshojoshua/jtnt-agent/internal/store"
	"github.com/tshojoshua/jtnt-agent/internal/transport"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "enroll":
		if err := enrollCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := statusCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		versionCmd()
	case "test-connection":
		if err := testConnectionCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("JTNT Agent CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  jtnt-agent enroll --token <TOKEN> --hub <URL>")
	fmt.Println("  jtnt-agent status")
	fmt.Println("  jtnt-agent version")
	fmt.Println("  jtnt-agent test-connection")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  enroll            Enroll agent with hub")
	fmt.Println("  status            Show agent status")
	fmt.Println("  version           Show agent version")
	fmt.Println("  test-connection   Test connection to hub")
}

func enrollCmd() error {
	fs := flag.NewFlagSet("enroll", flag.ExitOnError)
	token := fs.String("token", "", "Enrollment token")
	hub := fs.String("hub", "", "Hub URL")

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *token == "" {
		return fmt.Errorf("--token is required")
	}
	if *hub == "" {
		return fmt.Errorf("--hub is required")
	}

	fmt.Printf("Enrolling agent with hub: %s\n", *hub)

	// Create store
	s, err := store.NewStore(config.GetCertsDir())
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	// Create enroller
	enroller := enroll.NewEnroller(*hub, s)

	// Enroll
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := enroller.Enroll(ctx, *token)
	if err != nil {
		return fmt.Errorf("enrollment failed: %w", err)
	}

	fmt.Println("✓ Enrollment successful!")
	fmt.Printf("  Agent ID: %s\n", cfg.AgentID)
	fmt.Printf("  Hub URL:  %s\n", cfg.HubURL)
	fmt.Printf("  Config:   %s\n", config.GetConfigPath())
	fmt.Println()
	fmt.Println("Start the agent with:")
	fmt.Println("  sudo systemctl start jtnt-agent")
	fmt.Println("  or run manually: sudo jtnt-agentd")

	return nil
}

func statusCmd() error {
	// Check if enrolled
	configPath := config.GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("Agent is not enrolled")
		fmt.Println()
		fmt.Println("Enroll with:")
		fmt.Println("  jtnt-agent enroll --token <TOKEN> --hub <URL>")
		return nil
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Agent Status:")
	fmt.Printf("  Agent ID:         %s\n", cfg.AgentID)
	fmt.Printf("  Hub URL:          %s\n", cfg.HubURL)
	fmt.Printf("  Heartbeat:        %ds\n", cfg.HeartbeatSec)
	fmt.Printf("  Poll Interval:    %ds\n", cfg.PollIntervalSec)
	fmt.Printf("  Config File:      %s\n", configPath)
	fmt.Printf("  Token:            %s\n", maskToken(cfg.AgentToken))

	return nil
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

func versionCmd() {
	fmt.Printf("JTNT Agent version %s\n", version)
}

func testConnectionCmd() error {
	// Load config
	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Testing connection to: %s\n", cfg.HubURL)

	// Create client
	client, err := transport.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.TestConnection(ctx); err != nil {
		fmt.Println("✗ Connection test failed")
		return err
	}

	fmt.Println("✓ Connection test successful!")
	return nil
}
