package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient implements the Client interface using HTTP/REST
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// HTTPClientConfig holds configuration for the HTTP client
type HTTPClientConfig struct {
	BaseURL    string
	Timeout    time.Duration
	TLSEnabled bool
	TLSCert    string
	TLSKey     string
	TLSCA      string
	SkipVerify bool
}

// NewHTTPClient creates a new HTTP client with the given configuration
func NewHTTPClient(cfg HTTPClientConfig) (*HTTPClient, error) {
	transport := &http.Transport{}

	// TODO: Add TLS configuration support
	// if cfg.TLSEnabled {
	//     tlsConfig, err := buildTLSConfig(cfg)
	//     if err != nil {
	//         return nil, fmt.Errorf("failed to build TLS config: %w", err)
	//     }
	//     transport.TLSClientConfig = tlsConfig
	// }

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	return &HTTPClient{
		baseURL:    cfg.BaseURL,
		httpClient: httpClient,
	}, nil
}

// Register registers the agent with the config-server using a bootstrap token
func (c *HTTPClient) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	url := fmt.Sprintf("%s/api/v1/bootstrap-tokens/register", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal register request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Heartbeat sends a heartbeat to the config-server
func (c *HTTPClient) Heartbeat(ctx context.Context, targetID string) error {
	url := fmt.Sprintf("%s/api/v1/targets/%s/heartbeat", c.baseURL, targetID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetEffectiveChecks retrieves the effective checks for the target
func (c *HTTPClient) GetEffectiveChecks(ctx context.Context, targetID string) ([]EffectiveCheck, error) {
	url := fmt.Sprintf("%s/api/v1/checks/target/%s", c.baseURL, targetID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var checks []EffectiveCheck
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return checks, nil
}

// SubmitCheckResults submits check execution results to the config-server
func (c *HTTPClient) SubmitCheckResults(ctx context.Context, results []CheckResult) error {
	url := fmt.Sprintf("%s/api/v1/check-results", c.baseURL)

	req := SubmitCheckResultsRequest{Results: results}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// handleErrorResponse extracts error information from an HTTP response
func (c *HTTPClient) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	if errResp.Code != "" {
		return fmt.Errorf("server error [%s]: %s", errResp.Code, errResp.Error)
	}
	return fmt.Errorf("server error: %s", errResp.Error)
}
