package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/fregataa/aami/internal/config"
)

var (
	diffShowAll    bool
	diffOutputYAML bool
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show pending configuration changes",
	Long: `Compare current configuration with generated/applied configurations.

Shows differences between:
  - AAMI config and generated Prometheus config
  - Current config and last applied config
  - Node list changes

Examples:
  aami diff              # Show all pending changes
  aami diff prometheus   # Show Prometheus config changes
  aami diff nodes        # Show node changes`,
	RunE: runDiff,
}

var diffPrometheusCmd = &cobra.Command{
	Use:   "prometheus",
	Short: "Show Prometheus configuration differences",
	RunE:  runDiffPrometheus,
}

var diffNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Show node configuration differences",
	RunE:  runDiffNodes,
}

var diffAlertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Show alert rule differences",
	RunE:  runDiffAlerts,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().BoolVar(&diffShowAll, "all", false,
		"Show all differences including unchanged items")
	diffCmd.Flags().BoolVar(&diffOutputYAML, "yaml", false,
		"Output differences in YAML format")

	diffCmd.AddCommand(diffPrometheusCmd)
	diffCmd.AddCommand(diffNodesCmd)
	diffCmd.AddCommand(diffAlertsCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║           Configuration Differences                ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()

	// Check Prometheus config
	fmt.Println("Prometheus Configuration")
	fmt.Println(strings.Repeat("-", 50))
	if err := showPrometheusDiff(cfg); err != nil {
		fmt.Printf("  %s Could not compare: %v\n", color.YellowString("⚠"), err)
	}
	fmt.Println()

	// Check nodes
	fmt.Println("Node Configuration")
	fmt.Println(strings.Repeat("-", 50))
	if err := showNodesDiff(cfg); err != nil {
		fmt.Printf("  %s Could not compare: %v\n", color.YellowString("⚠"), err)
	}
	fmt.Println()

	// Check alerts
	fmt.Println("Alert Rules")
	fmt.Println(strings.Repeat("-", 50))
	if err := showAlertsDiff(cfg); err != nil {
		fmt.Printf("  %s Could not compare: %v\n", color.YellowString("⚠"), err)
	}
	fmt.Println()

	return nil
}

func runDiffPrometheus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("Prometheus Configuration Differences")
	fmt.Println(strings.Repeat("=", 50))
	return showPrometheusDiff(cfg)
}

func runDiffNodes(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("Node Configuration Differences")
	fmt.Println(strings.Repeat("=", 50))
	return showNodesDiff(cfg)
}

func runDiffAlerts(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("Alert Rules Differences")
	fmt.Println(strings.Repeat("=", 50))
	return showAlertsDiff(cfg)
}

func showPrometheusDiff(cfg *config.Config) error {
	// Compare expected targets with current targets file
	targetsPath := filepath.Join(cfg.Prometheus.StoragePath, "targets", "nodes.json")
	if cfg.Prometheus.StoragePath == "" {
		targetsPath = "/var/lib/aami/prometheus/targets/nodes.json"
	}

	// Check if targets file exists
	if _, err := os.Stat(targetsPath); os.IsNotExist(err) {
		fmt.Printf("  %s Targets file not found: %s\n", color.YellowString("+"), targetsPath)
		fmt.Println("    Run 'aami nodes install' to generate targets")
		return nil
	}

	// Read current targets
	currentData, err := os.ReadFile(targetsPath)
	if err != nil {
		return err
	}

	// Generate expected targets
	expectedTargets := generateExpectedTargets(cfg)

	// Compare
	if string(currentData) == expectedTargets {
		fmt.Printf("  %s Targets file is up to date\n", color.GreenString("✓"))
	} else {
		fmt.Printf("  %s Targets file needs update\n", color.YellowString("~"))
		if diffShowAll {
			fmt.Println("\n  Current:")
			printIndented(string(currentData), "    ")
			fmt.Println("\n  Expected:")
			printIndented(expectedTargets, "    ")
		}
	}

	return nil
}

func showNodesDiff(cfg *config.Config) error {
	// Compare configured nodes with discovered/last-applied nodes
	lastAppliedPath := "/var/lib/aami/.last-applied-nodes.yaml"

	// Check if last applied file exists
	if _, err := os.Stat(lastAppliedPath); os.IsNotExist(err) {
		fmt.Printf("  %s No previous node configuration found\n", color.YellowString("!"))
		fmt.Printf("  Current configuration has %d node(s)\n", len(cfg.Nodes))
		for _, node := range cfg.Nodes {
			fmt.Printf("    %s %s (%s)\n", color.GreenString("+"), node.Name, node.IP)
		}
		return nil
	}

	// Read last applied
	lastData, err := os.ReadFile(lastAppliedPath)
	if err != nil {
		return err
	}

	var lastNodes []config.NodeConfig
	if err := yaml.Unmarshal(lastData, &lastNodes); err != nil {
		return err
	}

	// Compare nodes
	added, removed, changed := compareNodes(lastNodes, cfg.Nodes)

	if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
		fmt.Printf("  %s No changes to node configuration\n", color.GreenString("✓"))
		return nil
	}

	for _, node := range added {
		fmt.Printf("  %s Added: %s (%s)\n", color.GreenString("+"), node.Name, node.IP)
	}
	for _, node := range removed {
		fmt.Printf("  %s Removed: %s (%s)\n", color.RedString("-"), node.Name, node.IP)
	}
	for _, node := range changed {
		fmt.Printf("  %s Changed: %s\n", color.YellowString("~"), node.Name)
	}

	return nil
}

func showAlertsDiff(cfg *config.Config) error {
	// Compare configured presets with applied rules
	appliedRulesPath := "/etc/prometheus/rules/aami-alerts.yaml"

	// Check if applied rules exist
	if _, err := os.Stat(appliedRulesPath); os.IsNotExist(err) {
		fmt.Printf("  %s No alert rules applied yet\n", color.YellowString("!"))
		if len(cfg.Alerts.Presets) > 0 {
			fmt.Printf("  Configured presets: %s\n", strings.Join(cfg.Alerts.Presets, ", "))
			fmt.Println("  Run 'aami alerts apply-preset' to apply rules")
		}
		return nil
	}

	// Check configured presets
	if len(cfg.Alerts.Presets) == 0 {
		fmt.Printf("  %s No alert presets configured\n", color.YellowString("!"))
		fmt.Println("  Run 'aami alerts apply-preset <preset>' to configure alerts")
		return nil
	}

	fmt.Printf("  %s Configured presets: %s\n", color.GreenString("✓"), strings.Join(cfg.Alerts.Presets, ", "))

	// Count custom rules
	if len(cfg.Alerts.Custom) > 0 {
		fmt.Printf("  %s Custom rules: %d\n", color.CyanString("•"), len(cfg.Alerts.Custom))
	}

	return nil
}

func generateExpectedTargets(cfg *config.Config) string {
	if len(cfg.Nodes) == 0 {
		return "[]"
	}

	var targets []map[string]interface{}
	for _, node := range cfg.Nodes {
		target := map[string]interface{}{
			"targets": []string{fmt.Sprintf("%s:9100", node.IP)},
			"labels": map[string]string{
				"node": node.Name,
			},
		}
		// Add custom labels
		for k, v := range node.Labels {
			target["labels"].(map[string]string)[k] = v
		}
		targets = append(targets, target)
	}

	data, _ := yaml.Marshal(targets)
	return string(data)
}

func compareNodes(old, new []config.NodeConfig) (added, removed, changed []config.NodeConfig) {
	oldMap := make(map[string]config.NodeConfig)
	newMap := make(map[string]config.NodeConfig)

	for _, n := range old {
		oldMap[n.Name] = n
	}
	for _, n := range new {
		newMap[n.Name] = n
	}

	// Find added
	for name, node := range newMap {
		if _, exists := oldMap[name]; !exists {
			added = append(added, node)
		}
	}

	// Find removed
	for name, node := range oldMap {
		if _, exists := newMap[name]; !exists {
			removed = append(removed, node)
		}
	}

	// Find changed
	for name, newNode := range newMap {
		if oldNode, exists := oldMap[name]; exists {
			if nodeChanged(oldNode, newNode) {
				changed = append(changed, newNode)
			}
		}
	}

	return
}

func nodeChanged(old, new config.NodeConfig) bool {
	if old.IP != new.IP {
		return true
	}
	if old.SSHUser != new.SSHUser {
		return true
	}
	if old.SSHPort != new.SSHPort {
		return true
	}
	if old.SSHKey != new.SSHKey {
		return true
	}
	return false
}

func printIndented(text, indent string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fmt.Printf("%s%s\n", indent, line)
	}
}

// DiffSummary provides a summary of all differences
type DiffSummary struct {
	PrometheusChanged bool
	NodesAdded        int
	NodesRemoved      int
	NodesChanged      int
	AlertsChanged     bool
	HasChanges        bool
}

// GetDiffSummary returns a summary of pending changes
func GetDiffSummary(cfg *config.Config) DiffSummary {
	summary := DiffSummary{}

	// Check nodes
	lastAppliedPath := "/var/lib/aami/.last-applied-nodes.yaml"
	if _, err := os.Stat(lastAppliedPath); err == nil {
		lastData, _ := os.ReadFile(lastAppliedPath)
		var lastNodes []config.NodeConfig
		if yaml.Unmarshal(lastData, &lastNodes) == nil {
			added, removed, changed := compareNodes(lastNodes, cfg.Nodes)
			summary.NodesAdded = len(added)
			summary.NodesRemoved = len(removed)
			summary.NodesChanged = len(changed)
		}
	} else {
		summary.NodesAdded = len(cfg.Nodes)
	}

	summary.HasChanges = summary.NodesAdded > 0 || summary.NodesRemoved > 0 ||
		summary.NodesChanged > 0 || summary.PrometheusChanged || summary.AlertsChanged

	return summary
}
