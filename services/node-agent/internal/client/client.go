package client

import "context"

// Client defines the interface for communicating with the config-server.
// This interface allows for different implementations (HTTP, gRPC) and testing.
type Client interface {
	// Register registers the agent with the config-server using a bootstrap token.
	// Returns the assigned TargetID on success.
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)

	// Heartbeat sends a heartbeat to the config-server to indicate the agent is alive.
	Heartbeat(ctx context.Context, targetID string) error

	// GetEffectiveChecks retrieves the effective checks for the target.
	GetEffectiveChecks(ctx context.Context, targetID string) ([]EffectiveCheck, error)

	// SubmitCheckResults submits check execution results to the config-server.
	// This is optional and may not be implemented in the initial version.
	SubmitCheckResults(ctx context.Context, results []CheckResult) error
}
