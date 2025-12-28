package client

import "fmt"

// CreateNamespace creates a new namespace
func (c *Client) CreateNamespace(req CreateNamespaceRequest) (*Namespace, error) {
	var ns Namespace
	if err := c.Post("/api/v1/namespaces", req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// ListNamespaces lists all namespaces
func (c *Client) ListNamespaces() ([]Namespace, error) {
	var response struct {
		Data []Namespace `json:"data"`
	}
	if err := c.Get("/api/v1/namespaces", &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetNamespace retrieves a namespace by ID
func (c *Client) GetNamespace(id string) (*Namespace, error) {
	var ns Namespace
	path := fmt.Sprintf("/api/v1/namespaces/%s", id)
	if err := c.Get(path, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// GetNamespaceByName retrieves a namespace by name
func (c *Client) GetNamespaceByName(name string) (*Namespace, error) {
	var ns Namespace
	path := fmt.Sprintf("/api/v1/namespaces/name/%s", name)
	if err := c.Get(path, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// UpdateNamespace updates a namespace
func (c *Client) UpdateNamespace(id string, req UpdateNamespaceRequest) (*Namespace, error) {
	var ns Namespace
	path := fmt.Sprintf("/api/v1/namespaces/%s", id)
	if err := c.Put(path, req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// DeleteNamespace deletes a namespace (soft delete)
func (c *Client) DeleteNamespace(id string) error {
	req := struct {
		ID string `json:"id"`
	}{ID: id}
	return c.Post("/api/v1/namespaces/delete", req, nil)
}
