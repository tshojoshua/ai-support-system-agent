package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/config"
)

const (
	defaultTimeout = 30 * time.Second
	maxIdleConns   = 10
	idleConnTimeout = 90 * time.Second
)

// Client is an mTLS HTTP client
type Client struct {
	httpClient  *http.Client
	baseURL     string
	retryConfig *RetryConfig
}

// NewClient creates a new mTLS client
func NewClient(cfg *config.Config) (*Client, error) {
	// Load client certificate
	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %w", err)
	}

	// Load CA bundle
	caCert, err := os.ReadFile(cfg.CABundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA bundle: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA bundle")
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}

	// Create HTTP transport
	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
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
