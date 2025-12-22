package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/config"
)

const (
	defaultTimeout  = 30 * time.Second
	maxIdleConns    = 10
	idleConnTimeout = 90 * time.Second
)

// Client is an HTTP client with bearer token authentication
type Client struct {
	httpClient  *http.Client
	baseURL     string
	agentToken  string
	retryConfig *RetryConfig
}

// NewClient creates a new HTTP client with bearer token authentication
func NewClient(cfg *config.Config) (*Client, error) {
	// Create HTTP transport with standard TLS
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        maxIdleConns,
		IdleConnTimeout:     idleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: maxIdleConns,
	}

	// Create HTTP client
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   defaultTimeout,
	}

	return &Client{
		httpClient:  httpClient,
		baseURL:     cfg.HubURL,
		agentToken:  cfg.AgentToken,
		retryConfig: DefaultRetryConfig(),
	}, nil
}

// Post sends a POST request with automatic retry
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	var respBody []byte
	var lastErr error

	err := RetryWithBackoff(ctx, c.retryConfig, func() error {
		data, err := c.doPost(ctx, path, body)
		if err != nil {
			lastErr = err
			return err
		}
		respBody = data
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("post request failed: %w (last error: %v)", err, lastErr)
	}

	return respBody, nil
}

// Get sends a GET request with automatic retry
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	var respBody []byte
	var lastErr error

	err := RetryWithBackoff(ctx, c.retryConfig, func() error {
		data, err := c.doGet(ctx, path)
		if err != nil {
			lastErr = err
			return err
		}
		respBody = data
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("get request failed: %w (last error: %v)", err, lastErr)
	}

	return respBody, nil
}

func (c *Client) doPost(ctx context.Context, path string, body interface{}) ([]byte, error) {
	// Marshal body
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.agentToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.agentToken)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429)")
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
		return nil, fmt.Errorf("client error (%d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return respBody, nil
}

func (c *Client) doGet(ctx context.Context, path string) ([]byte, error) {
	// Create request
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if c.agentToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.agentToken)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429)")
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
		return nil, fmt.Errorf("client error (%d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return respBody, nil
}

// TestConnection verifies mTLS connection to hub
func (c *Client) TestConnection(ctx context.Context) error {
	// Simple GET request to test connectivity
	_, err := c.Get(ctx, "/api/v1/agent/ping")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

// Upload uploads data with a specific content type
func (c *Client) Upload(ctx context.Context, url string, contentType string, data io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, data)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
