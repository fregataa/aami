package multicluster

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client connects to a remote AAMI cluster.
type Client struct {
	config     ClusterConfig
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new remote AAMI client.
func NewClient(cfg ClusterConfig) (*Client, error) {
	client := &Client{
		config:  cfg,
		baseURL: cfg.Endpoint,
	}

	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Configure TLS if certificates are provided
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		tlsConfig, err := client.buildTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("build TLS config: %w", err)
		}
		transport.TLSClientConfig = tlsConfig
	} else if cfg.SkipTLS {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client.httpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client, nil
}

// buildTLSConfig builds TLS configuration from certificates.
func (c *Client) buildTLSConfig() (*tls.Config, error) {
	// Load client certificate
	cert, err := tls.LoadX509KeyPair(c.config.TLSCert, c.config.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("load client cert: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Load CA certificate if provided
	if c.config.TLSCACert != "" {
		caCert, err := os.ReadFile(c.config.TLSCACert)
		if err != nil {
			return nil, fmt.Errorf("read CA cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}

		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// doRequest executes an HTTP request with authentication.
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Add authentication header if API key is provided
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// Ping checks if the cluster is reachable.
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/ping", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// GetStatus retrieves the cluster status.
func (c *Client) GetStatus(ctx context.Context) (*ClusterStatus, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/status", nil)
	if err != nil {
		return &ClusterStatus{
			Name:      c.config.Name,
			Endpoint:  c.config.Endpoint,
			Connected: false,
			Error:     err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &ClusterStatus{
			Name:      c.config.Name,
			Endpoint:  c.config.Endpoint,
			Connected: false,
			Error:     fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	var status ClusterStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}

	status.Name = c.config.Name
	status.Endpoint = c.config.Endpoint
	status.Connected = true
	status.LastSync = time.Now()

	return &status, nil
}

// GetHealth retrieves detailed health information.
func (c *Client) GetHealth(ctx context.Context) (*ClusterHealth, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var health ClusterHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("decode health: %w", err)
	}

	return &health, nil
}

// GetMetrics retrieves cluster metrics.
func (c *Client) GetMetrics(ctx context.Context) (*ClusterMetrics, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/metrics", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var metrics ClusterMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("decode metrics: %w", err)
	}

	return &metrics, nil
}

// GetAlerts retrieves active alerts from the cluster.
func (c *Client) GetAlerts(ctx context.Context) ([]GlobalAlert, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/alerts", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var alerts []GlobalAlert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, fmt.Errorf("decode alerts: %w", err)
	}

	// Tag all alerts with cluster name
	for i := range alerts {
		alerts[i].Cluster = c.config.Name
	}

	return alerts, nil
}

// GetNodes retrieves node list from the cluster.
func (c *Client) GetNodes(ctx context.Context) ([]NodeInfo, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/nodes", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var nodes []NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("decode nodes: %w", err)
	}

	return nodes, nil
}

// NodeInfo represents a node in the cluster.
type NodeInfo struct {
	Name        string  `json:"name"`
	IP          string  `json:"ip"`
	GPUCount    int     `json:"gpu_count"`
	HealthScore float64 `json:"health_score"`
	Status      string  `json:"status"`
}

// GetEvents retrieves recent events from the cluster.
func (c *Client) GetEvents(ctx context.Context, limit int) ([]ClusterEvent, error) {
	path := fmt.Sprintf("/api/v1/events?limit=%d", limit)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var events []ClusterEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decode events: %w", err)
	}

	// Tag all events with cluster name
	for i := range events {
		events[i].Cluster = c.config.Name
	}

	return events, nil
}

// GetVersion retrieves the AAMI version of the cluster.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/version", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode version: %w", err)
	}

	return result.Version, nil
}

// TestConnection tests the connection to the cluster.
func (c *Client) TestConnection(ctx context.Context) error {
	// Try ping first
	if err := c.Ping(ctx); err == nil {
		return nil
	}

	// Try status endpoint as fallback
	status, err := c.GetStatus(ctx)
	if err != nil {
		return err
	}

	if !status.Connected {
		return fmt.Errorf("connection failed: %s", status.Error)
	}

	return nil
}

// Close closes the client.
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// GetConfig returns the cluster configuration.
func (c *Client) GetConfig() ClusterConfig {
	return c.config
}

// SetTimeout sets the client timeout.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}
