package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

const (
	hubURL       = "https://hub.jtnt.us"
	apiBaseURL   = hubURL + "/api/v1"
	pollInterval = 5 * time.Minute
)

var (
	agentToken string
	userToken  string
	version    = "4.0.0"
)

type AgentStatus struct {
	AgentID  string    `json:"agent_id"`
	Status   string    `json:"status"`
	Enrolled bool      `json:"enrolled"`
	LastHB   time.Time `json:"last_heartbeat"`
	HubURL   string    `json:"hub_url"`
}

type Ticket struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Created  string `json:"created_at"`
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// Set icon and title
	systray.SetIcon(icon.Data)
	systray.SetTitle("JTNT Agent")
	systray.SetTooltip("JTNT Remote Management Agent")

	// Load tokens from environment or credential store
	loadTokens()

	// Menu structure
	mStatus := systray.AddMenuItem("Agent Status: Checking...", "View current agent status")
	mStatus.Disable()

	systray.AddSeparator()

	// Support ticket menu
	mTickets := systray.AddMenuItem("My Support Tickets", "View your support tickets")
	mNewTicket := systray.AddMenuItem("Create Support Ticket", "Submit a new support request")

	systray.AddSeparator()

	// Quick actions
	mOpenHub := systray.AddMenuItem("Open Hub Portal", "Open hub.jtnt.us in browser")
	mViewLogs := systray.AddMenuItem("View Logs", "Open agent log directory")

	systray.AddSeparator()

	// Agent control
	mRestart := systray.AddMenuItem("Restart Agent Service", "Restart the agent daemon")

	systray.AddSeparator()

	mAbout := systray.AddMenuItem(fmt.Sprintf("About (v%s)", version), "About JTNT Agent")
	mQuit := systray.AddMenuItem("Quit Tray App", "Exit the system tray application")

	// Update status immediately
	go updateAgentStatus(mStatus)

	// Start background refresh timer
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Event loop
	for {
		select {
		case <-mStatus.ClickedCh:
			showAgentDetails()

		case <-mTickets.ClickedCh:
			openURL(hubURL + "/support/tickets")

		case <-mNewTicket.ClickedCh:
			openURL(hubURL + "/support/new")

		case <-mOpenHub.ClickedCh:
			openURL(hubURL)

		case <-mViewLogs.ClickedCh:
			openLogsDirectory()

		case <-mRestart.ClickedCh:
			restartAgentService()

		case <-mAbout.ClickedCh:
			showAbout()

		case <-mQuit.ClickedCh:
			systray.Quit()
			return

		case <-ticker.C:
			go updateAgentStatus(mStatus)
		}
	}
}

func onExit() {
	// Cleanup if needed
}

func loadTokens() {
	// Try to load agent token from credential store or environment
	agentToken = os.Getenv("JTNT_AGENT_TOKEN")

	// Try to load user token for support ticket access
	userToken = os.Getenv("JTNT_USER_TOKEN")
}

func updateAgentStatus(mStatus *systray.MenuItem) {
	status, err := getAgentStatus()
	if err != nil {
		mStatus.SetTitle("Agent Status: Offline")
		mStatus.SetTooltip(fmt.Sprintf("Error: %v", err))
		return
	}

	if status.Enrolled {
		mStatus.SetTitle(fmt.Sprintf("Agent Status: %s", status.Status))
		mStatus.SetTooltip(fmt.Sprintf("Agent ID: %s\nLast Heartbeat: %s",
			status.AgentID, status.LastHB.Format("15:04:05")))
	} else {
		mStatus.SetTitle("Agent Status: Not Enrolled")
		mStatus.SetTooltip("Agent has not been enrolled yet")
	}
}

func getAgentStatus() (*AgentStatus, error) {
	// Try to get status from local agent CLI
	cmd := exec.Command("jtnt-agent", "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent status: %w", err)
	}

	var status AgentStatus
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return &status, nil
}

func showAgentDetails() {
	status, err := getAgentStatus()
	if err != nil {
		showNotification("Agent Status", fmt.Sprintf("Error: %v", err))
		return
	}

	details := fmt.Sprintf(
		"Agent ID: %s\nStatus: %s\nEnrolled: %v\nHub URL: %s\nLast Heartbeat: %s",
		status.AgentID,
		status.Status,
		status.Enrolled,
		status.HubURL,
		status.LastHB.Format("2006-01-02 15:04:05"),
	)

	showNotification("JTNT Agent Status", details)
}

func getMyTickets() ([]Ticket, error) {
	if userToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	req, err := http.NewRequest("GET", apiBaseURL+"/support/tickets", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+userToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Tickets []Ticket `json:"tickets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Tickets, nil
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		showNotification("Error", fmt.Sprintf("Failed to open URL: %v", err))
	}
}

func openLogsDirectory() {
	var logPath string
	switch runtime.GOOS {
	case "windows":
		logPath = `C:\ProgramData\JTNT\Agent\logs`
	case "darwin":
		logPath = "/var/log/jtnt-agent"
	default:
		logPath = "/var/log/jtnt-agent"
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", logPath)
	case "darwin":
		cmd = exec.Command("open", logPath)
	default:
		cmd = exec.Command("xdg-open", logPath)
	}

	if err := cmd.Start(); err != nil {
		showNotification("Error", fmt.Sprintf("Failed to open logs: %v", err))
	}
}

func restartAgentService() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("net", "stop", "JTNTAgent")
		if err := cmd.Run(); err != nil {
			showNotification("Error", fmt.Sprintf("Failed to stop service: %v", err))
			return
		}
		time.Sleep(2 * time.Second)
		cmd = exec.Command("net", "start", "JTNTAgent")
	case "darwin":
		cmd = exec.Command("launchctl", "kickstart", "-k", "system/us.jtnt.agentd")
	default:
		cmd = exec.Command("systemctl", "restart", "jtnt-agentd")
	}

	if err := cmd.Run(); err != nil {
		showNotification("Error", fmt.Sprintf("Failed to restart service: %v", err))
		return
	}

	showNotification("Service Restarted", "JTNT Agent service has been restarted")
}

func showAbout() {
	about := fmt.Sprintf(
		"JTNT Remote Management Agent\n\nVersion: %s\nHub: %s\n\nCopyright © 2025 JT&T Communications\nAll rights reserved.",
		version,
		hubURL,
	)
	showNotification("About JTNT Agent", about)
}

func showNotification(title, message string) {
	// Simple notification - for production, use github.com/gen2brain/beeep
	// which provides cross-platform native notifications
	fmt.Printf("[%s] %s\n", title, message)
}
