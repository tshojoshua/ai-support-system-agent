package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jtnt/agent/internal/config"
	"github.com/jtnt/agent/internal/enroll"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "enroll":
		enrollCmd()
	case "status":
		statusCmd()
	case "version":
		versionCmd()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("JTNT Agent CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  jtnt-agent enroll --token <TOKEN> --hub <URL>")
	fmt.Println("  jtnt-agent status")
	fmt.Println("  jtnt-agent version")
}

func enrollCmd() {
	enrollFlags := flag.NewFlagSet("enroll", flag.ExitOnError)
	token := enrollFlags.String("token", "", "Enrollment token")
	hubURL := enrollFlags.String("hub", "https://hub.jtnt.us", "Hub URL")

	enrollFlags.Parse(os.Args[2:])

	if *token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token is required\n")
		enrollFlags.Usage()
		os.Exit(1)
	}

	if err := enroll.Enroll(*hubURL, *token); err != nil {
		fmt.Fprintf(os.Stderr, "Enrollment failed: %v\n", err)
		os.Exit(1)
	}
}

func statusCmd() {
	if !config.Exists() {
		fmt.Println("Status: Not enrolled")
		fmt.Println("\nRun enrollment:")
		fmt.Println("  jtnt-agent enroll --token <TOKEN> --hub https://hub.jtnt.us")
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Status: Enrolled")
	fmt.Printf("Agent ID: %s\n", cfg.AgentID)
	fmt.Printf("Hub URL: %s\n", cfg.HubURL)
	fmt.Printf("Tenant ID: %s\n", cfg.TenantID)
	if cfg.SiteID != "" {
		fmt.Printf("Site ID: %s\n", cfg.SiteID)
	}
	fmt.Printf("Enrolled At: %s\n", cfg.EnrolledAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Config Path: %s\n", config.GetConfigPath())
}

func versionCmd() {
	fmt.Printf("JTNT Agent v%s\n", config.Version)
}
