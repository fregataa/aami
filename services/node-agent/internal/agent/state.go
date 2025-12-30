package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// State represents the persisted agent state
type State struct {
	TargetID    string    `json:"target_id"`
	Hostname    string    `json:"hostname,omitempty"`
	FirstSeen   time.Time `json:"first_seen"`
	LastUpdated time.Time `json:"last_updated"`
}

// LoadState loads the agent state from file
func LoadState(stateFilePath string) (*State, error) {
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No state file exists yet
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// SaveState saves the agent state to file
func SaveState(stateFilePath string, state *State) error {
	// Ensure directory exists
	dir := filepath.Dir(stateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Update timestamp
	state.LastUpdated = time.Now()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	tempFile := stateFilePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	if err := os.Rename(tempFile, stateFilePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// NewState creates a new empty state
func NewState() *State {
	now := time.Now()
	return &State{
		FirstSeen:   now,
		LastUpdated: now,
	}
}

// SetTargetID sets the target ID after successful registration
func (s *State) SetTargetID(targetID string) {
	s.TargetID = targetID
	s.LastUpdated = time.Now()
}

// SetHostname sets the hostname
func (s *State) SetHostname(hostname string) {
	s.Hostname = hostname
	s.LastUpdated = time.Now()
}

// IsRegistered returns true if the agent has been registered (has a target ID)
func (s *State) IsRegistered() bool {
	return s.TargetID != ""
}
