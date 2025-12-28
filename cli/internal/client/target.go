package client

import "fmt"

// CreateTarget creates a new target
func (c *Client) CreateTarget(req CreateTargetRequest) (*Target, error) {
	var target Target
	if err := c.Post("/api/v1/targets", req, &target); err != nil {
		return nil, err
	}
	return &target, nil
}

// ListTargets lists all targets
func (c *Client) ListTargets() ([]Target, error) {
	var response struct {
		Data []Target `json:"data"`
	}
	if err := c.Get("/api/v1/targets", &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// ListTargetsByGroup lists targets in a group
func (c *Client) ListTargetsByGroup(groupID string) ([]Target, error) {
	var targets []Target
	path := fmt.Sprintf("/api/v1/targets/group/%s", groupID)
	if err := c.Get(path, &targets); err != nil {
		return nil, err
	}
	return targets, nil
}

// GetTarget retrieves a target by ID
func (c *Client) GetTarget(id string) (*Target, error) {
	var target Target
	path := fmt.Sprintf("/api/v1/targets/%s", id)
	if err := c.Get(path, &target); err != nil {
		return nil, err
	}
	return &target, nil
}

// GetTargetByHostname retrieves a target by hostname
func (c *Client) GetTargetByHostname(hostname string) (*Target, error) {
	var target Target
	path := fmt.Sprintf("/api/v1/targets/hostname/%s", hostname)
	if err := c.Get(path, &target); err != nil {
		return nil, err
	}
	return &target, nil
}

// UpdateTarget updates a target
func (c *Client) UpdateTarget(id string, req UpdateTargetRequest) (*Target, error) {
	var target Target
	path := fmt.Sprintf("/api/v1/targets/%s", id)
	if err := c.Put(path, req, &target); err != nil {
		return nil, err
	}
	return &target, nil
}

// DeleteTarget deletes a target (soft delete)
func (c *Client) DeleteTarget(id string) error {
	req := struct {
		ID string `json:"id"`
	}{ID: id}
	return c.Post("/api/v1/targets/delete", req, nil)
}

// UpdateTargetStatus updates a target's status
func (c *Client) UpdateTargetStatus(id string, status string) error {
	req := UpdateTargetStatusRequest{Status: status}
	path := fmt.Sprintf("/api/v1/targets/%s/status", id)
	return c.Post(path, req, nil)
}

// HeartbeatTarget sends a heartbeat for a target
func (c *Client) HeartbeatTarget(id string) error {
	path := fmt.Sprintf("/api/v1/targets/%s/heartbeat", id)
	return c.Post(path, nil, nil)
}
