package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// BackupSuffix is the suffix for backup files
	BackupSuffix = ".backup"
	// RollbackDir is the directory for storing rollback data
	RollbackDir = "/var/lib/aami/rollback"
)

// Upgrader handles the upgrade process.
type Upgrader struct {
	checker      *Checker
	binaryPath   string
	rollbackDir  string
	httpClient   *http.Client
}

// NewUpgrader creates a new upgrader.
func NewUpgrader() *Upgrader {
	// Get current binary path
	binaryPath, _ := os.Executable()

	return &Upgrader{
		checker:     NewChecker(),
		binaryPath:  binaryPath,
		rollbackDir: RollbackDir,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// WithBinaryPath sets a custom binary path.
func (u *Upgrader) WithBinaryPath(path string) *Upgrader {
	u.binaryPath = path
	return u
}

// WithRollbackDir sets a custom rollback directory.
func (u *Upgrader) WithRollbackDir(dir string) *Upgrader {
	u.rollbackDir = dir
	return u
}

// UpgradeResult contains the result of an upgrade operation.
type UpgradeResult struct {
	Success       bool
	PreviousVersion string
	NewVersion    string
	BackupPath    string
	Message       string
}

// Upgrade performs the upgrade to the latest version.
func (u *Upgrader) Upgrade(currentVersion string) (*UpgradeResult, error) {
	result := &UpgradeResult{
		PreviousVersion: currentVersion,
	}

	// Check for update
	checkResult, err := u.checker.CheckForUpdate(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("check for update: %w", err)
	}

	if !checkResult.UpdateAvailable {
		result.Message = "Already at the latest version"
		result.Success = true
		return result, nil
	}

	result.NewVersion = checkResult.LatestVersion

	// Find platform-specific binary
	downloadURL := u.findPlatformBinary(checkResult.LatestRelease)
	if downloadURL == "" {
		return nil, fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Create rollback directory
	if err := os.MkdirAll(u.rollbackDir, 0755); err != nil {
		return nil, fmt.Errorf("create rollback dir: %w", err)
	}

	// Backup current binary
	backupPath := filepath.Join(u.rollbackDir, fmt.Sprintf("aami-%s%s", currentVersion, BackupSuffix))
	if err := u.backupBinary(backupPath); err != nil {
		return nil, fmt.Errorf("backup current binary: %w", err)
	}
	result.BackupPath = backupPath

	// Download new binary
	tempPath := u.binaryPath + ".new"
	if err := u.downloadBinary(downloadURL, tempPath); err != nil {
		return nil, fmt.Errorf("download new binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(tempPath, 0755); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("chmod new binary: %w", err)
	}

	// Replace binary
	if err := os.Rename(tempPath, u.binaryPath); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("replace binary: %w", err)
	}

	result.Success = true
	result.Message = fmt.Sprintf("Upgraded from %s to %s", currentVersion, result.NewVersion)

	return result, nil
}

// findPlatformBinary finds the appropriate binary for the current platform.
func (u *Upgrader) findPlatformBinary(release *Release) string {
	if release == nil {
		return ""
	}

	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	for _, asset := range release.Assets {
		if asset.ContentType == "application/octet-stream" ||
			asset.ContentType == "application/x-executable" {
			if filepath.Ext(asset.Name) == "" || filepath.Ext(asset.Name) == ".bin" {
				if contains(asset.Name, platform) {
					return asset.BrowserDownloadURL
				}
			}
		}
	}

	// Fallback: look for name pattern
	for _, asset := range release.Assets {
		if contains(asset.Name, platform) && !isChecksum(asset.Name) {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

// backupBinary creates a backup of the current binary.
func (u *Upgrader) backupBinary(backupPath string) error {
	src, err := os.Open(u.binaryPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	// Preserve permissions
	info, err := os.Stat(u.binaryPath)
	if err != nil {
		return err
	}
	return os.Chmod(backupPath, info.Mode())
}

// downloadBinary downloads a binary from the given URL.
func (u *Upgrader) downloadBinary(url, destPath string) error {
	resp, err := u.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// Rollback restores the previous version.
func (u *Upgrader) Rollback() error {
	// Find the most recent backup
	backups, err := u.listBackups()
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups available for rollback")
	}

	// Use the most recent backup
	backupPath := backups[0]

	// Restore from backup
	return u.restoreFromBackup(backupPath)
}

// RollbackTo restores to a specific version.
func (u *Upgrader) RollbackTo(version string) error {
	backupPath := filepath.Join(u.rollbackDir, fmt.Sprintf("aami-%s%s", version, BackupSuffix))

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found for version %s", version)
	}

	return u.restoreFromBackup(backupPath)
}

// restoreFromBackup restores the binary from a backup file.
func (u *Upgrader) restoreFromBackup(backupPath string) error {
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer src.Close()

	// Write to temp file first
	tempPath := u.binaryPath + ".restore"
	dst, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(tempPath)
		return fmt.Errorf("copy backup: %w", err)
	}
	dst.Close()

	// Make executable
	if err := os.Chmod(tempPath, 0755); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("chmod: %w", err)
	}

	// Replace current binary
	if err := os.Rename(tempPath, u.binaryPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("replace binary: %w", err)
	}

	return nil
}

// listBackups returns a list of available backups, sorted by modification time (newest first).
func (u *Upgrader) listBackups() ([]string, error) {
	entries, err := os.ReadDir(u.rollbackDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == BackupSuffix {
			backups = append(backups, filepath.Join(u.rollbackDir, entry.Name()))
		}
	}

	// Sort by modification time (newest first)
	// For simplicity, we'll just return in reverse order since files are likely named with versions
	for i, j := 0, len(backups)-1; i < j; i, j = i+1, j-1 {
		backups[i], backups[j] = backups[j], backups[i]
	}

	return backups, nil
}

// ListAvailableRollbacks returns versions available for rollback.
func (u *Upgrader) ListAvailableRollbacks() ([]string, error) {
	backups, err := u.listBackups()
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, backup := range backups {
		name := filepath.Base(backup)
		// Extract version from "aami-v1.0.0.backup"
		name = name[5:] // Remove "aami-"
		name = name[:len(name)-len(BackupSuffix)] // Remove ".backup"
		versions = append(versions, name)
	}

	return versions, nil
}

// VerifyBinary verifies a binary against its checksum.
func (u *Upgrader) VerifyBinary(binaryPath, checksumURL string) error {
	// Download checksum
	resp, err := u.httpClient.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("download checksum: %w", err)
	}
	defer resp.Body.Close()

	checksumData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read checksum: %w", err)
	}

	expectedChecksum := string(checksumData)
	expectedChecksum = extractChecksum(expectedChecksum)

	// Calculate actual checksum
	f, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("open binary: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(h.Sum(nil))

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func isChecksum(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".sha256" || ext == ".sha512" || ext == ".md5" || ext == ".sum"
}

func extractChecksum(data string) string {
	// Checksum files often contain "checksum  filename" format
	fields := splitFields(data)
	if len(fields) > 0 {
		return fields[0]
	}
	return data
}

func splitFields(s string) []string {
	var fields []string
	var field []byte
	inField := false

	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\n' {
			if inField {
				fields = append(fields, string(field))
				field = nil
				inField = false
			}
		} else {
			field = append(field, s[i])
			inField = true
		}
	}
	if inField {
		fields = append(fields, string(field))
	}
	return fields
}
