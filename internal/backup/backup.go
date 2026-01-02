// Package backup provides backup and restore functionality for AAMI configuration.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultBackupDir is the default directory for backups
	DefaultBackupDir = "/var/lib/aami/backups"
	// ConfigDir is the AAMI configuration directory
	ConfigDir = "/etc/aami"
	// DataDir is the AAMI data directory
	DataDir = "/var/lib/aami"
)

// BackupOptions configures the backup operation.
type BackupOptions struct {
	IncludeData    bool   // Include Prometheus/Grafana data
	IncludeConfigs bool   // Include configuration files (default: true)
	OutputDir      string // Output directory for backup file
	OutputFile     string // Custom output filename (optional)
}

// DefaultBackupOptions returns default backup options.
func DefaultBackupOptions() BackupOptions {
	return BackupOptions{
		IncludeData:    false,
		IncludeConfigs: true,
		OutputDir:      DefaultBackupDir,
	}
}

// BackupResult contains the result of a backup operation.
type BackupResult struct {
	FilePath    string
	Size        int64
	FileCount   int
	CreatedAt   time.Time
	IncludesData bool
}

// Backup represents the backup manager.
type Backup struct {
	configDir string
	dataDir   string
}

// NewBackup creates a new backup manager.
func NewBackup() *Backup {
	return &Backup{
		configDir: ConfigDir,
		dataDir:   DataDir,
	}
}

// WithConfigDir sets a custom config directory.
func (b *Backup) WithConfigDir(dir string) *Backup {
	b.configDir = dir
	return b
}

// WithDataDir sets a custom data directory.
func (b *Backup) WithDataDir(dir string) *Backup {
	b.dataDir = dir
	return b
}

// Create creates a new backup.
func (b *Backup) Create(opts BackupOptions) (*BackupResult, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	// Generate backup filename
	timestamp := time.Now().Format("2006-01-02-150405")
	filename := opts.OutputFile
	if filename == "" {
		if opts.IncludeData {
			filename = fmt.Sprintf("aami-full-backup-%s.tar.gz", timestamp)
		} else {
			filename = fmt.Sprintf("aami-backup-%s.tar.gz", timestamp)
		}
	}

	backupPath := filepath.Join(opts.OutputDir, filename)

	// Create the backup file
	file, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("create backup file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	result := &BackupResult{
		FilePath:     backupPath,
		CreatedAt:    time.Now(),
		IncludesData: opts.IncludeData,
	}

	// Backup configuration files
	if opts.IncludeConfigs {
		count, err := b.backupDirectory(tarWriter, b.configDir, "config")
		if err != nil {
			return nil, fmt.Errorf("backup config: %w", err)
		}
		result.FileCount += count
	}

	// Backup data if requested
	if opts.IncludeData {
		// Backup specific data directories (excluding large Prometheus data)
		dataDirs := []struct {
			src  string
			dest string
		}{
			{filepath.Join(b.dataDir, "alertmanager"), "data/alertmanager"},
			{filepath.Join(b.dataDir, "grafana"), "data/grafana"},
			{filepath.Join(b.dataDir, "rollback"), "data/rollback"},
		}

		for _, dir := range dataDirs {
			if _, err := os.Stat(dir.src); err == nil {
				count, err := b.backupDirectory(tarWriter, dir.src, dir.dest)
				if err != nil {
					// Log warning but continue
					fmt.Printf("Warning: failed to backup %s: %v\n", dir.src, err)
					continue
				}
				result.FileCount += count
			}
		}
	}

	// Add metadata file
	if err := b.addMetadata(tarWriter, result); err != nil {
		return nil, fmt.Errorf("add metadata: %w", err)
	}
	result.FileCount++

	// Get final file size
	fileInfo, err := os.Stat(backupPath)
	if err == nil {
		result.Size = fileInfo.Size()
	}

	return result, nil
}

// backupDirectory adds a directory to the tar archive.
func (b *Backup) backupDirectory(tw *tar.Writer, srcDir, destPrefix string) (int, error) {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return 0, nil // Directory doesn't exist, skip
	}

	count := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Set the name to include destination prefix
		if relPath == "." {
			header.Name = destPrefix
		} else {
			header.Name = filepath.Join(destPrefix, relPath)
		}

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
			count++
		}

		return nil
	})

	return count, err
}

// addMetadata adds a metadata file to the backup.
func (b *Backup) addMetadata(tw *tar.Writer, result *BackupResult) error {
	metadata := fmt.Sprintf(`# AAMI Backup Metadata
created_at: %s
includes_data: %v
file_count: %d
aami_version: unknown
`, result.CreatedAt.Format(time.RFC3339), result.IncludesData, result.FileCount)

	header := &tar.Header{
		Name:    "metadata.yaml",
		Mode:    0644,
		Size:    int64(len(metadata)),
		ModTime: result.CreatedAt,
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err := tw.Write([]byte(metadata))
	return err
}

// List returns a list of available backups.
func (b *Backup) List(backupDir string) ([]BackupInfo, error) {
	if backupDir == "" {
		backupDir = DefaultBackupDir
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isBackupFile(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Name:      name,
			Path:      filepath.Join(backupDir, name),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			IsFullBackup: isFullBackup(name),
		})
	}

	// Sort by creation time (newest first)
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[j].CreatedAt.After(backups[i].CreatedAt) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	return backups, nil
}

// BackupInfo contains information about a backup file.
type BackupInfo struct {
	Name         string
	Path         string
	Size         int64
	CreatedAt    time.Time
	IsFullBackup bool
}

// FormatSize returns a human-readable size string.
func (bi BackupInfo) FormatSize() string {
	const unit = 1024
	if bi.Size < unit {
		return fmt.Sprintf("%d B", bi.Size)
	}
	div, exp := int64(unit), 0
	for n := bi.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bi.Size)/float64(div), "KMGTPE"[exp])
}

// isBackupFile checks if a filename is a backup file.
func isBackupFile(name string) bool {
	return filepath.Ext(name) == ".gz" &&
		(len(name) > 7 && name[:5] == "aami-")
}

// isFullBackup checks if a backup includes data.
func isFullBackup(name string) bool {
	return len(name) > 14 && name[5:14] == "full-back"
}

// Delete removes a backup file.
func (b *Backup) Delete(backupPath string) error {
	// Verify it's a backup file
	if !isBackupFile(filepath.Base(backupPath)) {
		return fmt.Errorf("not a valid backup file: %s", backupPath)
	}

	return os.Remove(backupPath)
}
