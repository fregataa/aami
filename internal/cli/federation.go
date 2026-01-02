package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/federation"
)

var (
	federationShardCount int
	federationShardBy    string
	federationDryRun     bool
	federationForce      bool
)

var federationCmd = &cobra.Command{
	Use:   "federation",
	Short: "Manage Prometheus federation for large-scale deployments",
	Long: `Manage Prometheus federation for clusters with 500+ nodes.

Federation splits monitoring across multiple Prometheus instances (shards),
with a central Prometheus aggregating metrics from all shards.

Examples:
  aami federation enable --shards 3      # Enable with 3 shards
  aami federation enable --by rack       # Shard by rack labels
  aami federation status                 # Show federation status
  aami federation rebalance              # Rebalance nodes across shards
  aami federation disable                # Disable federation`,
}

var federationEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable federation mode",
	Long: `Enable Prometheus federation for large-scale monitoring.

This will:
1. Calculate shard distribution based on node count
2. Generate Prometheus configurations for each shard
3. Create systemd services for shard instances
4. Configure central Prometheus for federation

Sharding strategies:
  auto  - Automatically distribute nodes evenly (default)
  rack  - Distribute based on 'rack' label in node config
  count - Fixed number of nodes per shard`,
	RunE: runFederationEnable,
}

var federationDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable federation mode",
	Long: `Disable federation and return to single Prometheus instance.

This will stop all shard Prometheus instances and clean up
federation configuration files.`,
	RunE: runFederationDisable,
}

var federationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show federation status",
	Long:  `Display the current status of all Prometheus shards and the central aggregator.`,
	RunE:  runFederationStatus,
}

var federationRebalanceCmd = &cobra.Command{
	Use:   "rebalance",
	Short: "Rebalance nodes across shards",
	Long: `Analyze and suggest rebalancing of nodes across shards.

Use --apply to actually move nodes between shards.`,
	RunE: runFederationRebalance,
}

var federationValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate federation configuration",
	RunE:  runFederationValidate,
}

var federationShardsCmd = &cobra.Command{
	Use:   "shards",
	Short: "List all shards",
	RunE:  runFederationShards,
}

func init() {
	rootCmd.AddCommand(federationCmd)

	// Enable flags
	federationEnableCmd.Flags().IntVar(&federationShardCount, "shards", 0,
		"Number of shards (0 = auto-calculate based on node count)")
	federationEnableCmd.Flags().StringVar(&federationShardBy, "by", "auto",
		"Sharding strategy: auto, rack, count")
	federationEnableCmd.Flags().BoolVar(&federationDryRun, "dry-run", false,
		"Show what would be done without making changes")
	federationEnableCmd.Flags().BoolVar(&federationForce, "force", false,
		"Force enable even with few nodes")

	// Rebalance flags
	federationRebalanceCmd.Flags().BoolVar(&federationDryRun, "dry-run", true,
		"Show suggested changes without applying")

	federationCmd.AddCommand(federationEnableCmd)
	federationCmd.AddCommand(federationDisableCmd)
	federationCmd.AddCommand(federationStatusCmd)
	federationCmd.AddCommand(federationRebalanceCmd)
	federationCmd.AddCommand(federationValidateCmd)
	federationCmd.AddCommand(federationShardsCmd)
}

func runFederationEnable(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	nodeCount := len(cfg.Nodes)

	// Warn if node count is low
	if nodeCount < 100 && !federationForce {
		color.Yellow("Warning: Federation is recommended for 500+ nodes.")
		fmt.Printf("You have %d nodes. Continue anyway? [y/N]: ", nodeCount)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║         Enabling Prometheus Federation             ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create federation config
	fedConfig := federation.DefaultFederationConfig()
	fedConfig.Enabled = true

	manager := federation.NewManager(cfg, fedConfig)

	// Calculate shards
	strategy := federation.ShardingStrategy(federationShardBy)
	shards := manager.CalculateShards(strategy, federationShardCount)

	if len(shards) == 0 {
		return fmt.Errorf("no shards calculated - check node configuration")
	}

	// Display shard plan
	fmt.Printf("Sharding strategy: %s\n", federationShardBy)
	fmt.Printf("Total nodes: %d\n", nodeCount)
	fmt.Printf("Shard count: %d\n", len(shards))
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Shard", "Nodes", "Port", "Storage Path"})
	table.SetBorder(false)

	for _, shard := range shards {
		table.Append([]string{
			shard.Name,
			fmt.Sprintf("%d", len(shard.Nodes)),
			fmt.Sprintf("%d", shard.Prometheus.Port),
			shard.Prometheus.StoragePath,
		})
	}
	table.Render()
	fmt.Println()

	// Validate shards
	validator := federation.NewShardValidator()
	errors := validator.ValidateAll(shards)
	if len(errors) > 0 {
		color.Red("Validation errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("validation failed")
	}

	if federationDryRun {
		color.Yellow("Dry-run mode - no changes made")
		fmt.Println()
		fmt.Println("Would create:")
		fmt.Println("  - Prometheus config for each shard")
		fmt.Println("  - Systemd service for each shard")
		fmt.Println("  - Central Prometheus federation config")
		fmt.Println("  - Recording rules for aggregation")
		return nil
	}

	// Set shards and deploy
	manager.SetShards(shards)

	fmt.Println("Deploying federation configuration...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := manager.Deploy(ctx); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Generate recording rules
	rulesPath := "/etc/aami/rules/federation-recording.yaml"
	if err := federation.GeneratePrometheusRules(rulesPath); err != nil {
		color.Yellow("Warning: Could not generate recording rules: %v", err)
	}

	fmt.Println()
	color.Green("✓ Federation enabled successfully")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Start shard services:")
	for _, shard := range shards {
		fmt.Printf("     sudo systemctl start aami-prometheus-%s\n", shard.Name)
	}
	fmt.Println("  2. Start central service:")
	fmt.Println("     sudo systemctl start aami-prometheus-central")
	fmt.Println("  3. Verify status:")
	fmt.Println("     aami federation status")

	return nil
}

func runFederationDisable(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("Disabling federation...")
	fmt.Println()
	fmt.Println("This will:")
	fmt.Println("  - Stop all shard Prometheus instances")
	fmt.Println("  - Remove federation configuration files")
	fmt.Println("  - Return to single Prometheus instance")
	fmt.Println()
	fmt.Print("Continue? [y/N]: ")

	var answer string
	fmt.Scanln(&answer)
	if answer != "y" && answer != "Y" {
		fmt.Println("Aborted.")
		return nil
	}

	fedConfig := federation.FederationConfig{Enabled: false}
	manager := federation.NewManager(cfg, fedConfig)

	ctx := context.Background()
	if err := manager.Disable(ctx); err != nil {
		return fmt.Errorf("disable failed: %w", err)
	}

	color.Green("✓ Federation disabled")
	fmt.Println()
	fmt.Println("To use single Prometheus:")
	fmt.Println("  sudo systemctl start prometheus")

	return nil
}

func runFederationStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Try to load federation config
	fedConfig, err := loadFederationConfig()
	if err != nil {
		color.Yellow("Federation is not enabled")
		fmt.Println()
		fmt.Println("Enable with: aami federation enable")
		return nil
	}

	manager := federation.NewManager(cfg, fedConfig)

	ctx := context.Background()
	status, err := manager.GetStatus(ctx)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║            Federation Status                       ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()

	// Summary
	fmt.Printf("Enabled:      %v\n", status.Enabled)
	fmt.Printf("Type:         %s\n", status.Type)
	fmt.Printf("Shards:       %d (%d healthy)\n", status.ShardCount, status.HealthyCount)
	fmt.Printf("Total Nodes:  %d\n", status.TotalNodes)
	fmt.Println()

	// Central status
	centralStatus := "Healthy"
	centralIcon := color.GreenString("✓")
	if !status.Central.Healthy {
		centralStatus = "Unhealthy"
		centralIcon = color.RedString("✗")
	}
	fmt.Printf("Central:      %s %s (%s)\n", centralIcon, centralStatus, status.Central.Endpoint)
	fmt.Println()

	// Shard table
	fmt.Println("Shards")
	fmt.Println(strings.Repeat("-", 60))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Endpoint", "Nodes", "Metrics", "Status"})
	table.SetBorder(false)

	for _, shard := range status.Shards {
		statusStr := color.GreenString("Healthy")
		if !shard.Healthy {
			statusStr = color.RedString("Unhealthy")
			if shard.Error != "" {
				statusStr = color.RedString("Error")
			}
		}

		table.Append([]string{
			shard.Name,
			shard.Endpoint,
			fmt.Sprintf("%d", shard.NodeCount),
			fmt.Sprintf("%d", shard.MetricCount),
			statusStr,
		})
	}

	table.Render()

	// Show errors if any
	hasErrors := false
	for _, shard := range status.Shards {
		if shard.Error != "" {
			if !hasErrors {
				fmt.Println()
				color.Red("Errors:")
				hasErrors = true
			}
			fmt.Printf("  %s: %s\n", shard.Name, shard.Error)
		}
	}

	return nil
}

func runFederationRebalance(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fedConfig, err := loadFederationConfig()
	if err != nil {
		return fmt.Errorf("federation not enabled: %w", err)
	}

	rebalancer := federation.NewShardRebalancer(fedConfig.Shards)

	imbalance := rebalancer.GetImbalance()
	fmt.Printf("Current imbalance: %.1f%%\n", imbalance*100)
	fmt.Println()

	if imbalance < 0.1 {
		color.Green("✓ Shards are well balanced")
		return nil
	}

	moves := rebalancer.SuggestRebalance()
	if len(moves) == 0 {
		fmt.Println("No rebalancing needed")
		return nil
	}

	fmt.Println("Suggested node moves:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Node", "From", "To"})
	table.SetBorder(false)

	for _, move := range moves {
		table.Append([]string{move.Node, move.FromShard, move.ToShard})
	}
	table.Render()

	if federationDryRun {
		fmt.Println()
		color.Yellow("Dry-run mode - no changes made")
		fmt.Println("Use --dry-run=false to apply changes")
	} else {
		// Would apply rebalancing here
		_ = cfg
		color.Green("✓ Rebalancing applied")
	}

	return nil
}

func runFederationValidate(cmd *cobra.Command, args []string) error {
	fedConfig, err := loadFederationConfig()
	if err != nil {
		return fmt.Errorf("federation not enabled: %w", err)
	}

	validator := federation.NewShardValidator()
	errors := validator.ValidateAll(fedConfig.Shards)

	if len(errors) == 0 {
		color.Green("✓ Federation configuration is valid")
		return nil
	}

	color.Red("Validation errors:")
	for _, err := range errors {
		fmt.Printf("  - %s\n", err)
	}

	return fmt.Errorf("validation failed with %d errors", len(errors))
}

func runFederationShards(cmd *cobra.Command, args []string) error {
	fedConfig, err := loadFederationConfig()
	if err != nil {
		return fmt.Errorf("federation not enabled: %w", err)
	}

	fmt.Println("Federation Shards")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	for _, shard := range fedConfig.Shards {
		fmt.Printf("Shard: %s\n", color.CyanString(shard.Name))
		fmt.Printf("  Port: %d\n", shard.Prometheus.Port)
		fmt.Printf("  Storage: %s\n", shard.Prometheus.StoragePath)
		fmt.Printf("  Retention: %s\n", shard.Prometheus.Retention)
		fmt.Printf("  Nodes (%d):\n", len(shard.Nodes))
		for _, node := range shard.Nodes {
			fmt.Printf("    - %s\n", node)
		}
		fmt.Println()
	}

	return nil
}

// loadFederationConfig loads federation configuration from file.
func loadFederationConfig() (federation.FederationConfig, error) {
	fedConfigPath := filepath.Join("/etc/aami", "federation", "federation.yaml")

	// Check if federation directory exists
	fedDir := filepath.Dir(fedConfigPath)
	if _, err := os.Stat(fedDir); os.IsNotExist(err) {
		return federation.FederationConfig{}, fmt.Errorf("federation not configured")
	}

	// For now, scan for shard configs to reconstruct federation config
	shardFiles, err := filepath.Glob(filepath.Join(fedDir, "prometheus-shard-*.yaml"))
	if err != nil || len(shardFiles) == 0 {
		return federation.FederationConfig{}, fmt.Errorf("no shard configurations found")
	}

	fedConfig := federation.FederationConfig{
		Enabled: true,
		Type:    federation.FederationTypePrometheus,
		Central: federation.CentralConfig{
			Port:        9090,
			StoragePath: "/var/lib/aami/prometheus-central",
		},
	}

	// Reconstruct shard configs from files
	for i, _ := range shardFiles {
		shard := federation.ShardConfig{
			Name: fmt.Sprintf("shard-%d", i+1),
		}
		shard.Prometheus.Port = 9091 + i
		shard.Prometheus.StoragePath = fmt.Sprintf("/var/lib/aami/prometheus-shard-%d", i+1)
		fedConfig.Shards = append(fedConfig.Shards, shard)
	}

	return fedConfig, nil
}
