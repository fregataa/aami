package config

import (
	"fmt"
	"net"
	"os"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate validates the configuration and returns any validation errors
func (c *Config) Validate() []ValidationError {
	var errors []ValidationError

	if c.Cluster.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "cluster.name",
			Message: "required",
		})
	}

	for i, node := range c.Nodes {
		if node.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("nodes[%d].name", i),
				Message: "required",
			})
		}
		if node.IP == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("nodes[%d].ip", i),
				Message: "required",
			})
		} else if net.ParseIP(node.IP) == nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("nodes[%d].ip", i),
				Message: "invalid IP address",
			})
		}
		if node.SSHKey != "" {
			if _, err := os.Stat(node.SSHKey); os.IsNotExist(err) {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("nodes[%d].ssh_key", i),
					Message: "file not found",
				})
			}
		}
		if node.SSHPort != 0 && (node.SSHPort < 1 || node.SSHPort > 65535) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("nodes[%d].ssh_port", i),
				Message: "invalid port number",
			})
		}
	}

	if c.SSH.MaxParallel < 0 {
		errors = append(errors, ValidationError{
			Field:   "ssh.max_parallel",
			Message: "must be non-negative",
		})
	}

	if c.SSH.ConnectTimeout < 0 {
		errors = append(errors, ValidationError{
			Field:   "ssh.connect_timeout",
			Message: "must be non-negative",
		})
	}

	if c.Prometheus.Port != 0 && (c.Prometheus.Port < 1 || c.Prometheus.Port > 65535) {
		errors = append(errors, ValidationError{
			Field:   "prometheus.port",
			Message: "invalid port number",
		})
	}

	if c.Grafana.Port != 0 && (c.Grafana.Port < 1 || c.Grafana.Port > 65535) {
		errors = append(errors, ValidationError{
			Field:   "grafana.port",
			Message: "invalid port number",
		})
	}

	return errors
}

// IsValid returns true if the configuration is valid
func (c *Config) IsValid() bool {
	return len(c.Validate()) == 0
}
