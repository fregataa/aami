package prometheus

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RuleFileManager manages Prometheus rule files with atomic operations
type RuleFileManager struct {
	basePath          string
	backupPath        string
	logger            *slog.Logger
	enableValidation  bool
	enableBackup      bool
	promtoolPath      string
}

// RuleFileManagerConfig holds configuration for RuleFileManager
type RuleFileManagerConfig struct {
	BasePath         string
	BackupPath       string
	EnableValidation bool
	EnableBackup     bool
	PromtoolPath     string // Path to promtool binary (optional, will search PATH if empty)
}

// NewRuleFileManager creates a new RuleFileManager instance
func NewRuleFileManager(config RuleFileManagerConfig, logger *slog.Logger) (*RuleFileManager, error) {
	// Validate base path exists
	if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
		// Try to create the directory
		if err := os.MkdirAll(config.BasePath, 0755); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, err)
		}
	}

	// Set default backup path if not provided
	backupPath := config.BackupPath
	if backupPath == "" {
		backupPath = filepath.Join(config.BasePath, ".backup")
	}

	// Create backup directory if backup is enabled
	if config.EnableBackup {
		if err := os.MkdirAll(backupPath, 0755); err != nil {
			return nil, fmt.Errorf("%w: failed to create backup directory: %v", ErrBackupFailed, err)
		}
	}

	// Find promtool if validation is enabled
	promtoolPath := config.PromtoolPath
	if config.EnableValidation && promtoolPath == "" {
		// Try to find promtool in PATH
		path, err := exec.LookPath("promtool")
		if err != nil {
			logger.Warn("promtool not found in PATH, validation will be disabled", "error", err)
			config.EnableValidation = false
		} else {
			promtoolPath = path
		}
	}

	return &RuleFileManager{
		basePath:         config.BasePath,
		backupPath:       backupPath,
		logger:           logger,
		enableValidation: config.EnableValidation,
		enableBackup:     config.EnableBackup,
		promtoolPath:     promtoolPath,
	}, nil
}

// WriteRuleFile writes a rule file atomically with optional validation and backup
func (m *RuleFileManager) WriteRuleFile(groupID string, content []byte) error {
	filename := m.getFilePath(groupID)

	// Create backup of existing file if enabled
	if m.enableBackup {
		if _, err := os.Stat(filename); err == nil {
			if err := m.BackupRuleFile(groupID); err != nil {
				m.logger.Warn("Failed to backup existing file", "group_id", groupID, "error", err)
			}
		}
	}

	// Atomic write: write to temp file first
	tempFile := filename + ".tmp"

	// Write content to temporary file
	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		return fmt.Errorf("%w: failed to write temp file: %s", ErrAtomicWriteFailed, err)
	}

	// Ensure temp file is deleted on error
	defer func() {
		if _, err := os.Stat(tempFile); err == nil {
			os.Remove(tempFile)
		}
	}()

	// Validate the rule file if enabled
	if m.enableValidation {
		if err := m.validateRuleFileContent(tempFile); err != nil {
			// Restore from backup if validation fails
			if m.enableBackup {
				if restoreErr := m.restoreFromBackup(groupID); restoreErr != nil {
					m.logger.Error("Failed to restore from backup after validation failure",
						"group_id", groupID, "error", restoreErr)
				}
			}
			return fmt.Errorf("%w: %s", ErrValidationFailed, err)
		}
	}

	// Atomic rename: replace the target file
	if err := os.Rename(tempFile, filename); err != nil {
		return fmt.Errorf("%w: failed to rename temp file: %s", ErrAtomicWriteFailed, err)
	}

	m.logger.Info("Successfully wrote rule file", "group_id", groupID, "file", filename)
	return nil
}

// DeleteRuleFile deletes a rule file for a specific group
func (m *RuleFileManager) DeleteRuleFile(groupID string) error {
	filename := m.getFilePath(groupID)

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		m.logger.Debug("Rule file does not exist, nothing to delete", "file", filename)
		return nil
	}

	// Create backup before deletion if enabled
	if m.enableBackup {
		if err := m.BackupRuleFile(groupID); err != nil {
			m.logger.Warn("Failed to backup before deletion", "group_id", groupID, "error", err)
		}
	}

	// Delete the file
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	m.logger.Info("Successfully deleted rule file", "group_id", groupID, "file", filename)
	return nil
}

// ListRuleFiles returns a list of all rule files
func (m *RuleFileManager) ListRuleFiles() ([]string, error) {
	pattern := filepath.Join(m.basePath, "group-*.yml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrListFailed, err)
	}

	// Extract group IDs from filenames
	var groupIDs []string
	for _, file := range files {
		basename := filepath.Base(file)
		// Extract group ID from "group-{id}.yml"
		if strings.HasPrefix(basename, "group-") && strings.HasSuffix(basename, ".yml") {
			groupID := strings.TrimPrefix(basename, "group-")
			groupID = strings.TrimSuffix(groupID, ".yml")
			groupIDs = append(groupIDs, groupID)
		}
	}

	return groupIDs, nil
}

// ValidateRuleFile validates a rule file using promtool
func (m *RuleFileManager) ValidateRuleFile(groupID string) error {
	filename := m.getFilePath(groupID)
	return m.validateRuleFileContent(filename)
}

// BackupRuleFile creates a backup of the rule file
func (m *RuleFileManager) BackupRuleFile(groupID string) error {
	if !m.enableBackup {
		return nil
	}

	sourceFile := m.getFilePath(groupID)

	// Check if source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return nil // No file to backup
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(m.backupPath, fmt.Sprintf("group-%s.%s.yml", groupID, timestamp))

	// Copy the file
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("%w: failed to read source file: %s", ErrBackupFailed, err)
	}

	if err := os.WriteFile(backupFile, content, 0644); err != nil {
		return fmt.Errorf("%w: failed to write backup file: %s", ErrBackupFailed, err)
	}

	m.logger.Debug("Created backup", "group_id", groupID, "backup_file", backupFile)
	return nil
}

// CleanupOldBackups removes backup files older than the specified duration
func (m *RuleFileManager) CleanupOldBackups(maxAge time.Duration) error {
	if !m.enableBackup {
		return nil
	}

	pattern := filepath.Join(m.backupPath, "group-*.yml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrListFailed, err)
	}

	cutoff := time.Now().Add(-maxAge)
	removedCount := 0

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				m.logger.Warn("Failed to remove old backup", "file", file, "error", err)
			} else {
				removedCount++
			}
		}
	}

	if removedCount > 0 {
		m.logger.Info("Cleaned up old backups", "removed_count", removedCount, "max_age", maxAge)
	}

	return nil
}

// GetFilePath returns the full path for a group's rule file
func (m *RuleFileManager) GetFilePath(groupID string) string {
	return m.getFilePath(groupID)
}

// getFilePath returns the file path for a given group ID
func (m *RuleFileManager) getFilePath(groupID string) string {
	return filepath.Join(m.basePath, fmt.Sprintf("group-%s.yml", groupID))
}

// validateRuleFileContent validates a rule file using promtool
func (m *RuleFileManager) validateRuleFileContent(filePath string) error {
	if !m.enableValidation {
		return nil
	}

	cmd := exec.Command(m.promtoolPath, "check", "rules", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: promtool error: %v, output: %s", ErrValidationFailed, err, string(output))
	}

	return nil
}

// restoreFromBackup restores the most recent backup for a group
func (m *RuleFileManager) restoreFromBackup(groupID string) error {
	if !m.enableBackup {
		return nil
	}

	// Find the most recent backup
	pattern := filepath.Join(m.backupPath, fmt.Sprintf("group-%s.*.yml", groupID))
	backups, err := filepath.Glob(pattern)
	if err != nil || len(backups) == 0 {
		return fmt.Errorf("%w: no backup found for group %s", ErrRestoreFailed, groupID)
	}

	// Get the most recent backup (last in alphabetical order due to timestamp format)
	var mostRecent string
	var mostRecentTime time.Time

	for _, backup := range backups {
		info, err := os.Stat(backup)
		if err != nil {
			continue
		}
		if mostRecent == "" || info.ModTime().After(mostRecentTime) {
			mostRecent = backup
			mostRecentTime = info.ModTime()
		}
	}

	if mostRecent == "" {
		return fmt.Errorf("%w: no valid backup found for group %s", ErrRestoreFailed, groupID)
	}

	// Restore the backup
	content, err := os.ReadFile(mostRecent)
	if err != nil {
		return fmt.Errorf("%w: failed to read backup: %v", ErrRestoreFailed, err)
	}

	targetFile := m.getFilePath(groupID)
	if err := os.WriteFile(targetFile, content, 0644); err != nil {
		return fmt.Errorf("%w: failed to write restored file: %v", ErrRestoreFailed, err)
	}

	m.logger.Info("Restored from backup", "group_id", groupID, "backup_file", mostRecent)
	return nil
}
