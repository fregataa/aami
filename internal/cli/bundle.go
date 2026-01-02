package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/installer"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Manage offline bundles",
	Long:  "Create and manage offline bundles for air-gap installations.",
}

var bundleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create offline bundle",
	Long: `Create an offline bundle containing all required components.

The bundle includes:
  - Prometheus
  - Alertmanager
  - Grafana
  - Node Exporter
  - DCGM Exporter
  - Alert rules and dashboards

Example:
  aami bundle create --output aami-offline-v0.1.0.tar.gz`,
	RunE: runBundleCreate,
}

var bundleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bundle contents",
	Args:  cobra.ExactArgs(1),
	RunE:  runBundleList,
}

var bundleOutput string
var includeGrafana bool

func init() {
	bundleCreateCmd.Flags().StringVar(&bundleOutput, "output",
		"aami-offline.tar.gz", "Output file path")
	bundleCreateCmd.Flags().BoolVar(&includeGrafana, "include-grafana",
		true, "Include Grafana in bundle")

	bundleCmd.AddCommand(bundleCreateCmd)
	bundleCmd.AddCommand(bundleListCmd)
	rootCmd.AddCommand(bundleCmd)
}

func runBundleCreate(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println("Creating offline bundle...")
	fmt.Println()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "aami-bundle-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download components
	components := []string{"prometheus", "alertmanager", "node_exporter"}
	for _, name := range components {
		component := installer.Components[name]
		url := component.DownloadURL()
		dest := filepath.Join(tmpDir, fmt.Sprintf("%s.tar.gz", name))

		fmt.Printf("  %s Downloading %s v%s...\n", yellow("•"), name, component.Version)

		if err := downloadToFile(url, dest); err != nil {
			return fmt.Errorf("download %s: %w", name, err)
		}
		fmt.Printf("  %s Downloaded %s\n", green("✓"), name)
	}

	// Copy alert rules
	rulesDir := filepath.Join(tmpDir, "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("create rules dir: %w", err)
	}

	// Generate default rules
	fmt.Printf("  %s Including alert rules...\n", yellow("•"))
	for presetName, preset := range presets {
		content := generatePrometheusRules(preset)
		rulePath := filepath.Join(rulesDir, fmt.Sprintf("%s.yaml", presetName))
		if err := os.WriteFile(rulePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write rules: %w", err)
		}
	}
	fmt.Printf("  %s Alert rules included\n", green("✓"))

	// Create manifest
	manifest := fmt.Sprintf(`# AAMI Offline Bundle Manifest
version: %s
components:
  - prometheus: %s
  - alertmanager: %s
  - node_exporter: %s
`,
		Version,
		installer.Components["prometheus"].Version,
		installer.Components["alertmanager"].Version,
		installer.Components["node_exporter"].Version,
	)

	if err := os.WriteFile(filepath.Join(tmpDir, "MANIFEST"), []byte(manifest), 0644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	// Create tar.gz bundle
	fmt.Printf("  %s Creating bundle archive...\n", yellow("•"))
	if err := createTarGz(tmpDir, bundleOutput); err != nil {
		return fmt.Errorf("create bundle: %w", err)
	}

	// Get file size
	info, err := os.Stat(bundleOutput)
	if err != nil {
		return err
	}

	fmt.Printf("  %s Bundle created\n", green("✓"))
	fmt.Println()
	fmt.Printf("Bundle: %s (%.1f MB)\n", bundleOutput, float64(info.Size())/(1024*1024))
	fmt.Println()
	fmt.Println("To install on air-gapped machine:")
	fmt.Printf("  aami init --offline %s\n", bundleOutput)

	return nil
}

func runBundleList(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]

	file, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("open bundle: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("read gzip: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	fmt.Printf("Contents of %s:\n\n", bundlePath)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		size := ""
		if header.Size > 0 {
			if header.Size > 1024*1024 {
				size = fmt.Sprintf("%.1f MB", float64(header.Size)/(1024*1024))
			} else if header.Size > 1024 {
				size = fmt.Sprintf("%.1f KB", float64(header.Size)/1024)
			} else {
				size = fmt.Sprintf("%d B", header.Size)
			}
		}

		fmt.Printf("  %-40s %s\n", header.Name, size)
	}

	return nil
}

func downloadToFile(url, dest string) error {
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

func createTarGz(srcDir, dest string) error {
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}
