package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fregataa/aami/internal/config"
)

var cfgFile string
var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "aami",
	Short: "AI Accelerator Monitoring Infrastructure",
	Long: `AAMI - GPU cluster monitoring tool with Prometheus stack.

Simplifies the installation, configuration, and operation of the Prometheus
stack through a single CLI, with GPU-specific diagnostic features.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: /etc/aami/config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile(config.DefaultConfigPath)
	}
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

// loadConfig loads the configuration file
func loadConfig() (*config.Config, error) {
	path := cfgFile
	if path == "" {
		path = config.DefaultConfigPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s\nRun 'aami init' to create one", path)
	}

	return config.Load(path)
}

// saveConfig saves the configuration to file
func saveConfig(c *config.Config) error {
	path := cfgFile
	if path == "" {
		path = config.DefaultConfigPath
	}
	return config.Save(c, path)
}
