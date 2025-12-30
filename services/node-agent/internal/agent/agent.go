package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fregataa/aami/node-agent/internal/client"
	"github.com/fregataa/aami/node-agent/internal/config"
)

// Agent represents the node agent that communicates with the config-server
type Agent struct {
	cfg    *config.Config
	logger *slog.Logger
	client client.Client
	state  *State

	mu       sync.RWMutex
	targetID string

	// Channels for coordination
	stopCh chan struct{}
	doneCh chan struct{}
}

// New creates a new Agent instance
func New(cfg *config.Config, logger *slog.Logger) (*Agent, error) {
	// Create HTTP client
	httpClient, err := client.NewHTTPClient(client.HTTPClientConfig{
		BaseURL:    cfg.Server.URL,
		Timeout:    cfg.Server.Timeout,
		TLSEnabled: cfg.Server.TLSEnabled,
		TLSCert:    cfg.Server.TLSCert,
		TLSKey:     cfg.Server.TLSKey,
		TLSCA:      cfg.Server.TLSCA,
		SkipVerify: cfg.Server.SkipVerify,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Agent{
		cfg:    cfg,
		logger: logger,
		client: httpClient,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}, nil
}

// Run starts the agent and blocks until the context is cancelled
func (a *Agent) Run(ctx context.Context) error {
	defer close(a.doneCh)

	// Step 1: Load existing state
	state, err := LoadState(a.cfg.Agent.StateFile)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state == nil {
		state = NewState()
		a.logger.Info("created new agent state")
	} else {
		a.logger.Info("loaded existing state", "target_id", state.TargetID)
	}
	a.state = state

	// Step 2: Check if already registered, otherwise register
	if state.IsRegistered() {
		a.setTargetID(state.TargetID)
		a.logger.Info("using existing registration", "target_id", state.TargetID)
	} else {
		if err := a.register(ctx); err != nil {
			return fmt.Errorf("failed to register with config-server: %w", err)
		}
	}

	// Step 3: Start background workers
	var wg sync.WaitGroup

	// Start heartbeat worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.heartbeatWorker(ctx)
	}()

	// Start check poller worker (placeholder for Phase 2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.checkPollerWorker(ctx)
	}()

	// Wait for context cancellation
	<-ctx.Done()
	a.logger.Info("shutting down agent")

	// Signal workers to stop
	close(a.stopCh)

	// Wait for workers to finish
	wg.Wait()

	// Save final state
	if err := SaveState(a.cfg.Agent.StateFile, a.state); err != nil {
		a.logger.Error("failed to save state on shutdown", "error", err)
	}

	return nil
}

// register handles the registration flow with retry logic
func (a *Agent) register(ctx context.Context) error {
	if a.cfg.Agent.BootstrapToken == "" {
		return fmt.Errorf("bootstrap token required for registration (use --bootstrap-token flag)")
	}

	req := client.RegisterRequest{
		Token:     a.cfg.Agent.BootstrapToken,
		Hostname:  a.cfg.Agent.Hostname,
		IPAddress: a.cfg.Agent.IPAddress,
		GroupID:   a.cfg.Agent.GroupID,
		Labels:    a.cfg.Agent.Labels,
		Metadata:  a.cfg.Agent.Metadata,
	}

	var lastErr error
	for attempt := 0; attempt < a.cfg.Agent.RegistrationMaxRetries; attempt++ {
		if attempt > 0 {
			a.logger.Info("retrying registration",
				"attempt", attempt+1,
				"max_retries", a.cfg.Agent.RegistrationMaxRetries,
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(a.cfg.Agent.RegistrationRetryInterval):
			}
		}

		resp, err := a.client.Register(ctx, req)
		if err != nil {
			lastErr = err
			a.logger.Warn("registration attempt failed", "attempt", attempt+1, "error", err)
			continue
		}

		// Registration successful
		a.setTargetID(resp.Target.ID)
		a.state.SetTargetID(resp.Target.ID)
		a.state.SetHostname(resp.Target.Hostname)

		// Persist state immediately
		if err := SaveState(a.cfg.Agent.StateFile, a.state); err != nil {
			a.logger.Error("failed to save state after registration", "error", err)
		}

		a.logger.Info("registered successfully",
			"target_id", resp.Target.ID,
			"hostname", resp.Target.Hostname,
			"token_usage", resp.TokenUsage,
			"remaining_uses", resp.RemainingUses,
		)

		return nil
	}

	return fmt.Errorf("registration failed after %d attempts: %w",
		a.cfg.Agent.RegistrationMaxRetries, lastErr)
}

// setTargetID thread-safely sets the target ID
func (a *Agent) setTargetID(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.targetID = id
}

// getTargetID thread-safely gets the target ID
func (a *Agent) getTargetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.targetID
}

// heartbeatWorker sends periodic heartbeats to the config-server
func (a *Agent) heartbeatWorker(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.Agent.HeartbeatInterval)
	defer ticker.Stop()

	a.logger.Info("heartbeat worker started", "interval", a.cfg.Agent.HeartbeatInterval)

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("heartbeat worker stopping")
			return
		case <-a.stopCh:
			a.logger.Info("heartbeat worker stopping")
			return
		case <-ticker.C:
			a.sendHeartbeat(ctx)
		}
	}
}

// sendHeartbeat sends a single heartbeat
func (a *Agent) sendHeartbeat(ctx context.Context) {
	targetID := a.getTargetID()
	if targetID == "" {
		a.logger.Warn("skipping heartbeat: no target ID")
		return
	}

	if err := a.client.Heartbeat(ctx, targetID); err != nil {
		a.logger.Warn("heartbeat failed", "target_id", targetID, "error", err)
	} else {
		a.logger.Debug("heartbeat sent", "target_id", targetID)
	}
}

// checkPollerWorker polls for check configurations (placeholder for Phase 2)
func (a *Agent) checkPollerWorker(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.Agent.CheckPollInterval)
	defer ticker.Stop()

	a.logger.Info("check poller worker started", "interval", a.cfg.Agent.CheckPollInterval)

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("check poller worker stopping")
			return
		case <-a.stopCh:
			a.logger.Info("check poller worker stopping")
			return
		case <-ticker.C:
			a.pollChecks(ctx)
		}
	}
}

// pollChecks fetches and processes checks (placeholder for Phase 2)
func (a *Agent) pollChecks(ctx context.Context) {
	targetID := a.getTargetID()
	if targetID == "" {
		a.logger.Warn("skipping check poll: no target ID")
		return
	}

	checks, err := a.client.GetEffectiveChecks(ctx, targetID)
	if err != nil {
		a.logger.Warn("failed to get effective checks", "target_id", targetID, "error", err)
		return
	}

	a.logger.Debug("fetched checks", "target_id", targetID, "count", len(checks))

	// TODO: Phase 2 - Execute checks and collect results
	for _, check := range checks {
		a.logger.Debug("check available",
			"name", check.Name,
			"script_type", check.ScriptType,
			"version", check.Version,
		)
	}
}
