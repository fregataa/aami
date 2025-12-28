package cmd

import (
	"fmt"
	"os"

	"github.com/fregataa/aami/cli/internal/client"
	"github.com/fregataa/aami/cli/internal/config"
	"github.com/fregataa/aami/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	serverURL    string
	outputFormat string

	// Global state
	apiClient *client.Client
	cfg       *config.Config
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "aami",
	Short: "AAMI CLI - Manage AAMI monitoring infrastructure",
	Long: `AAMI CLI is a command-line tool for managing the AAMI monitoring infrastructure.
It provides commands for managing namespaces, groups, targets, exporters, alerts, and more.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override with flags if provided
		if serverURL != "" {
			cfg.Server = serverURL
		}
		if outputFormat != "" {
			cfg.Default.Output = outputFormat
		}

		// Override with environment variable
		if envServer := os.Getenv("AAMI_SERVER"); envServer != "" {
			cfg.Server = envServer
		}

		// Initialize API client
		apiClient = client.NewClient(cfg.Server)

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "", "Server URL (default from config)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: table|json|yaml (default from config)")

	// Add subcommands
	rootCmd.AddCommand(namespaceCmd)
	rootCmd.AddCommand(groupCmd)
	rootCmd.AddCommand(targetCmd)
	rootCmd.AddCommand(bootstrapTokenCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

// getFormatter returns the appropriate formatter based on config
func getFormatter() output.Formatter {
	format := cfg.Default.Output
	if outputFormat != "" {
		format = outputFormat
	}
	return output.NewFormatter(format, os.Stdout)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aami version 0.1.0")
	},
}
