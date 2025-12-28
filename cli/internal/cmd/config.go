package cmd

import (
	"fmt"

	"github.com/fregataa/aami/cli/internal/config"
	"github.com/fregataa/aami/cli/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `View, set, and initialize CLI configuration.`,
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Example: `  # View configuration
  aami config view`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Format as YAML
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Print(string(data))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Example: `  # Set server URL
  aami config set server http://localhost:8080

  # Set default output format
  aami config set output json

  # Set default namespace
  aami config set default-namespace production`,
	ValidArgs: []string{"server", "default-namespace", "output", "no-headers", "color"},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := config.Set(key, value); err != nil {
			return fmt.Errorf("failed to set config: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Example: `  # Get server URL
  aami config get server

  # Get default output format
  aami config get output`,
	ValidArgs: []string{"server", "default-namespace", "output", "no-headers", "color"},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		value, err := config.Get(key)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		fmt.Println(value)
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file with defaults",
	Example: `  # Initialize configuration
  aami config init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Init(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		configPath, _ := config.GetConfigPath()
		output.PrintSuccess(fmt.Sprintf("Initialized configuration at: %s", configPath))
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Example: `  # Show config path
  aami config path`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}

		fmt.Println(configPath)
		return nil
	},
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
}
