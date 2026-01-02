package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AAMI configuration",
	Long: `Initialize AAMI configuration and directories.

This command creates:
  - /etc/aami/config.yaml      Configuration file
  - /var/lib/aami/prometheus   Prometheus data directory
  - /var/lib/aami/grafana      Grafana data directory
  - /var/lib/aami/targets      Service discovery targets

Examples:
  aami init                      # Online initialization
  aami init --offline bundle.tar.gz  # Offline initialization`,
	RunE: runInit,
}

var offlineBundle string

func init() {
	initCmd.Flags().StringVar(&offlineBundle, "offline", "",
		"Path to offline bundle for air-gap installation")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println("Initializing AAMI...")
	fmt.Println()

	// 1. Create directories
	dirs := []string{
		"/etc/aami",
		"/etc/aami/rules",
		"/var/lib/aami/prometheus",
		"/var/lib/aami/grafana",
		"/var/lib/aami/targets",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
		fmt.Printf("  %s Created %s\n", green("✓"), dir)
	}

	// 2. Create default config if not exists
	configPath := config.DefaultConfigPath
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return err
		}
		fmt.Printf("  %s Created %s\n", green("✓"), configPath)
	} else {
		fmt.Printf("  %s Config already exists: %s\n", yellow("•"), configPath)
	}

	fmt.Println()

	// 3. Install components
	if offlineBundle != "" {
		return installOffline(offlineBundle)
	}
	return installOnline()
}

func createDefaultConfig(path string) error {
	defaultConfig := `# AAMI Configuration
# See https://github.com/fregataa/aami for documentation

cluster:
  name: my-gpu-cluster

# Nodes to monitor (add via 'aami nodes add' or manually)
nodes: []

# SSH settings for connecting to nodes
ssh:
  max_parallel: 50
  connect_timeout: 10
  command_timeout: 300
  retry:
    max_attempts: 3
    backoff_base: 2
    backoff_max: 30

# Alert configuration
alerts:
  presets:
    - gpu-production

# Notification channels
notifications:
  slack:
    enabled: false
    webhook_url: ""
    channel: ""
  email:
    enabled: false
    smtp_host: ""
    smtp_port: 587
    from: ""
    to: []

# Prometheus settings
prometheus:
  retention: 15d
  storage_path: /var/lib/aami/prometheus
  port: 9090

# Grafana settings
grafana:
  port: 3000
  admin_password: ${GRAFANA_ADMIN_PASSWORD}
`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

func installOnline() error {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println("Installing components (online)...")
	fmt.Println()

	// TODO: Implement actual component installation
	// For now, just print what would be installed
	components := []string{
		"Prometheus v2.48.0",
		"Alertmanager v0.26.0",
		"Grafana v10.2.3",
	}

	for _, comp := range components {
		fmt.Printf("  %s %s (pending implementation)\n", green("•"), comp)
	}

	fmt.Println()
	fmt.Printf("%s AAMI initialized successfully!\n", green("✓"))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit /etc/aami/config.yaml to configure your cluster")
	fmt.Println("  2. Add nodes: aami nodes add <name> --ip <ip> --user root --key ~/.ssh/id_rsa")
	fmt.Println("  3. Check status: aami status")
	fmt.Println()

	return nil
}

func installOffline(bundlePath string) error {
	green := color.New(color.FgGreen).SprintFunc()

	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle not found: %s", bundlePath)
	}

	fmt.Printf("Installing from offline bundle: %s\n", bundlePath)
	fmt.Println()

	// TODO: Implement actual offline installation
	fmt.Printf("  %s Extracting bundle...\n", green("•"))
	fmt.Printf("  %s Installing components...\n", green("•"))

	fmt.Println()
	fmt.Printf("%s AAMI initialized successfully (offline mode)!\n", green("✓"))
	fmt.Println()

	return nil
}
