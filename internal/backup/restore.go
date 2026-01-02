package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RestoreOptions configures the restore operation.
type RestoreOptions struct {
	ConfigOnly  bool   // Only restore configuration files
	DataOnly    bool   // Only restore data files
	DryRun      bool   // Show what would be restored without making changes
	Force       bool   // Overwrite existing files without prompting
	TargetDir   string // Custom target directory (for testing)
}

// DefaultRestoreOptions returns default restore options.
func DefaultRestoreOptions() RestoreOptions {
	return RestoreOptions{
		ConfigOnly: false,
		DataOnly:   false,
		DryRun:     false,
		Force:      false,
	}
}

// RestoreResult contains the result of a restore operation.
type RestoreResult struct {
	Success       bool
	FilesRestored int
	FilesSkipped  int
	Errors        []string
	RestoredAt    time.Time
}

// Restore restores from a backup file.
func (b *Backup) Restore(backupPath string, opts RestoreOptions) (*RestoreResult, error) {
	result := &RestoreResult{
		RestoredAt: time.Now(),
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup file not found: %s", backupPath)
	}

	// Open backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("read gzip: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Process each file in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}

		// Determine if this file should be restored
		shouldRestore, destPath := b.shouldRestore(header.Name, opts)
		if !shouldRestore {
			result.FilesSkipped++
			continue
		}

		if opts.DryRun {
			fmt.Printf("Would restore: %s -> %s\n", header.Name, destPath)
			result.FilesRestored++
			continue
		}

		// Restore the file
		if err := b.restoreFile(tarReader, header, destPath, opts); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", header.Name, err))
			continue
		}

		result.FilesRestored++
	}

	result.Success = len(result.Errors) == 0
	return result, nil
}

// shouldRestore determines if a file should be restored and returns the destination path.
func (b *Backup) shouldRestore(archivePath string, opts RestoreOptions) (bool, string) {
	// Skip metadata file
	if archivePath == "metadata.yaml" {
		return false, ""
	}

	// Determine the type of file
	isConfig := strings.HasPrefix(archivePath, "config/")
	isData := strings.HasPrefix(archivePath, "data/")

	// Apply filters
	if opts.ConfigOnly && !isConfig {
		return false, ""
	}
	if opts.DataOnly && !isData {
		return false, ""
	}

	// Calculate destination path
	var destPath string
	if opts.TargetDir != "" {
		destPath = filepath.Join(opts.TargetDir, archivePath)
	} else if isConfig {
		relPath := strings.TrimPrefix(archivePath, "config/")
		destPath = filepath.Join(b.configDir, relPath)
	} else if isData {
		relPath := strings.TrimPrefix(archivePath, "data/")
		destPath = filepath.Join(b.dataDir, relPath)
	} else {
		// Unknown path structure
		return false, ""
	}

	return true, destPath
}

// restoreFile restores a single file from the archive.
func (b *Backup) restoreFile(reader io.Reader, header *tar.Header, destPath string, opts RestoreOptions) error {
	// Handle directories
	if header.Typeflag == tar.TypeDir {
		return os.MkdirAll(destPath, os.FileMode(header.Mode))
	}

	// Check if file exists and we're not forcing
	if !opts.Force {
		if _, err := os.Stat(destPath); err == nil {
			// File exists, create backup
			backupPath := destPath + ".pre-restore"
			if err := os.Rename(destPath, backupPath); err != nil {
				return fmt.Errorf("backup existing file: %w", err)
			}
		}
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(destPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Handle symlinks
	if header.Typeflag == tar.TypeSymlink {
		// Remove existing symlink if it exists
		os.Remove(destPath)
		return os.Symlink(header.Linkname, destPath)
	}

	// Handle regular files
	if header.Typeflag == tar.TypeReg || header.Typeflag == 0 {
		file, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer file.Close()

		if _, err := io.Copy(file, reader); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		// Set modification time
		return os.Chtimes(destPath, header.ModTime, header.ModTime)
	}

	return nil
}

// Verify checks if a backup file is valid.
func (b *Backup) Verify(backupPath string) (*VerifyResult, error) {
	result := &VerifyResult{
		Path: backupPath,
	}

	// Check file exists
	info, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	result.Size = info.Size()

	// Open and verify the archive
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		result.IsValid = false
		result.Error = "Invalid gzip format"
		return result, nil
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	hasMetadata := false
	hasConfig := false

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.IsValid = false
			result.Error = fmt.Sprintf("Invalid tar format: %v", err)
			return result, nil
		}

		result.FileCount++

		if header.Name == "metadata.yaml" {
			hasMetadata = true
		}
		if strings.HasPrefix(header.Name, "config/") {
			hasConfig = true
		}
		if strings.HasPrefix(header.Name, "data/") {
			result.HasData = true
		}
	}

	result.IsValid = hasMetadata || hasConfig
	result.HasMetadata = hasMetadata

	if !result.IsValid {
		result.Error = "Missing required files (config or metadata)"
	}

	return result, nil
}

// VerifyResult contains the result of a backup verification.
type VerifyResult struct {
	Path        string
	Size        int64
	FileCount   int
	IsValid     bool
	HasMetadata bool
	HasData     bool
	Error       string
}

// ReadMetadata reads metadata from a backup file.
func (b *Backup) ReadMetadata(backupPath string) (map[string]string, error) {
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "metadata.yaml" {
			content, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}
			return parseSimpleYAML(string(content)), nil
		}
	}

	return nil, fmt.Errorf("metadata not found in backup")
}

// parseSimpleYAML parses a simple YAML file (key: value format only).
func parseSimpleYAML(content string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		result[key] = value
	}

	return result
}

// ListContents lists the contents of a backup file.
func (b *Backup) ListContents(backupPath string) ([]string, error) {
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var files []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		typeStr := ""
		switch header.Typeflag {
		case tar.TypeDir:
			typeStr = "d"
		case tar.TypeSymlink:
			typeStr = "l"
		default:
			typeStr = "-"
		}

		files = append(files, fmt.Sprintf("%s %8d %s",
			typeStr, header.Size, header.Name))
	}

	return files, nil
}
