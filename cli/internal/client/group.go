package client

import "fmt"

// CreateGroup creates a new group
func (c *Client) CreateGroup(req CreateGroupRequest) (*Group, error) {
	var group Group
	if err := c.Post("/api/v1/groups", req, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// ListGroups lists all groups
func (c *Client) ListGroups() ([]Group, error) {
	var response struct {
		Data []Group `json:"data"`
	}
	if err := c.Get("/api/v1/groups", &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// ListGroupsByNamespace lists groups in a namespace
func (c *Client) ListGroupsByNamespace(namespaceID string) ([]Group, error) {
	var groups []Group
	path := fmt.Sprintf("/api/v1/groups/namespace/%s", namespaceID)
	if err := c.Get(path, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// GetGroup retrieves a group by ID
func (c *Client) GetGroup(id string) (*Group, error) {
	var group Group
	path := fmt.Sprintf("/api/v1/groups/%s", id)
	if err := c.Get(path, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup updates a group
func (c *Client) UpdateGroup(id string, req UpdateGroupRequest) (*Group, error) {
	var group Group
	path := fmt.Sprintf("/api/v1/groups/%s", id)
	if err := c.Put(path, req, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// DeleteGroup deletes a group (soft delete)
func (c *Client) DeleteGroup(id string) error {
	req := struct {
		ID string `json:"id"`
	}{ID: id}
	return c.Post("/api/v1/groups/delete", req, nil)
}

// GetGroupChildren retrieves child groups
func (c *Client) GetGroupChildren(id string) ([]Group, error) {
	var groups []Group
	path := fmt.Sprintf("/api/v1/groups/%s/children", id)
	if err := c.Get(path, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// GetGroupAncestors retrieves ancestor groups
func (c *Client) GetGroupAncestors(id string) ([]Group, error) {
	var groups []Group
	path := fmt.Sprintf("/api/v1/groups/%s/ancestors", id)
	if err := c.Get(path, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}
