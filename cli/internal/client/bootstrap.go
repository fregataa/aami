package client

import "fmt"

// CreateBootstrapToken creates a new bootstrap token
func (c *Client) CreateBootstrapToken(req CreateBootstrapTokenRequest) (*BootstrapToken, error) {
	var token BootstrapToken
	if err := c.Post("/api/v1/bootstrap-tokens", req, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// ListBootstrapTokens lists all bootstrap tokens
func (c *Client) ListBootstrapTokens() ([]BootstrapToken, error) {
	var response struct {
		Data []BootstrapToken `json:"data"`
	}
	if err := c.Get("/api/v1/bootstrap-tokens", &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetBootstrapToken retrieves a bootstrap token by ID
func (c *Client) GetBootstrapToken(id string) (*BootstrapToken, error) {
	var token BootstrapToken
	path := fmt.Sprintf("/api/v1/bootstrap-tokens/%s", id)
	if err := c.Get(path, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// GetBootstrapTokenByToken retrieves a bootstrap token by token string
func (c *Client) GetBootstrapTokenByToken(tokenStr string) (*BootstrapToken, error) {
	var token BootstrapToken
	path := fmt.Sprintf("/api/v1/bootstrap-tokens/token/%s", tokenStr)
	if err := c.Get(path, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateBootstrapToken updates a bootstrap token
func (c *Client) UpdateBootstrapToken(id string, req UpdateBootstrapTokenRequest) (*BootstrapToken, error) {
	var token BootstrapToken
	path := fmt.Sprintf("/api/v1/bootstrap-tokens/%s", id)
	if err := c.Put(path, req, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteBootstrapToken deletes a bootstrap token (soft delete)
func (c *Client) DeleteBootstrapToken(id string) error {
	req := struct {
		ID string `json:"id"`
	}{ID: id}
	return c.Post("/api/v1/bootstrap-tokens/delete", req, nil)
}

// ValidateBootstrapToken validates a bootstrap token
func (c *Client) ValidateBootstrapToken(tokenStr string) (*BootstrapToken, error) {
	var token BootstrapToken
	req := ValidateTokenRequest{Token: tokenStr}
	if err := c.Post("/api/v1/bootstrap-tokens/validate", req, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// RegisterNodeWithToken registers a node using a bootstrap token
func (c *Client) RegisterNodeWithToken(req BootstrapRegisterRequest) (*BootstrapRegisterResponse, error) {
	var response BootstrapRegisterResponse
	if err := c.Post("/api/v1/bootstrap-tokens/register", req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
