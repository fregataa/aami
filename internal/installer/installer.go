package installer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// Component represents an installable component
type Component struct {
	Name    string
	Version string
	Binary  string
}

// Components contains the default component versions
var Components = map[string]Component{
	"prometheus": {
		Name:    "prometheus",
		Version: "2.48.0",
		Binary:  "prometheus",
	},
	"alertmanager": {
		Name:    "alertmanager",
		Version: "0.26.0",
		Binary:  "alertmanager",
	},
	"node_exporter": {
		Name:    "node_exporter",
		Version: "1.7.0",
		Binary:  "node_exporter",
	},
	"dcgm_exporter": {
		Name:    "dcgm-exporter",
		Version: "3.3.0",
		Binary:  "dcgm-exporter",
	},
}

// DownloadURL returns the download URL for a component
func (c Component) DownloadURL() string {
	arch := runtime.GOARCH
	goos := runtime.GOOS

	// Map to expected naming
	osName := goos
	if osName == "darwin" {
		osName = "darwin"
	}

	switch c.Name {
	case "prometheus", "alertmanager", "node_exporter":
		return fmt.Sprintf(
			"https://github.com/prometheus/%s/releases/download/v%s/%s-%s.%s-%s.tar.gz",
			c.Name, c.Version, c.Name, c.Version, osName, arch)
	case "dcgm-exporter":
		// DCGM exporter is typically installed via container or package
		return fmt.Sprintf(
			"https://github.com/NVIDIA/dcgm-exporter/releases/download/v%s/dcgm-exporter",
			c.Version)
	default:
		return ""
	}
}

// Installer handles component installation
type Installer struct {
	installDir string
	binDir     string
}

// NewInstaller creates a new installer
func NewInstaller(installDir, binDir string) *Installer {
	if installDir == "" {
		installDir = "/opt/aami"
	}
	if binDir == "" {
		binDir = "/usr/local/bin"
	}
	return &Installer{
		installDir: installDir,
		binDir:     binDir,
	}
}

// Install downloads and installs a component
func (i *Installer) Install(name string) error {
	component, ok := Components[name]
	if !ok {
		return fmt.Errorf("unknown component: %s", name)
	}

	fmt.Printf("Installing %s v%s...\n", name, component.Version)

	// Create install directory
	componentDir := filepath.Join(i.installDir, name)
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Download
	url := component.DownloadURL()
	tarPath := filepath.Join(componentDir, fmt.Sprintf("%s.tar.gz", name))

	if err := downloadFile(url, tarPath); err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer os.Remove(tarPath)

	// Extract
	if err := extractTarGz(tarPath, componentDir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	// Find and link binary
	binaryPath := findBinary(componentDir, component.Binary)
	if binaryPath == "" {
		return fmt.Errorf("binary not found: %s", component.Binary)
	}

	// Create symlink
	linkPath := filepath.Join(i.binDir, component.Binary)
	os.Remove(linkPath) // Remove existing symlink if any
	if err := os.Symlink(binaryPath, linkPath); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}

	fmt.Printf("Installed %s to %s\n", name, linkPath)
	return nil
}

// downloadFile downloads a file from URL to destination
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarGz extracts a tar.gz file
func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

// findBinary finds a binary in a directory tree
func findBinary(dir, name string) string {
	var result string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == name && !info.IsDir() {
			result = path
			return filepath.SkipDir
		}
		return nil
	})
	return result
}
