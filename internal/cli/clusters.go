package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/multicluster"
)

var clustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Manage multiple AAMI clusters",
	Long: `Manage multiple AAMI clusters from a central location.

This allows you to:
  - Register remote AAMI clusters
  - View aggregated status across all clusters
  - Monitor alerts from all clusters
  - Collect metrics from multiple sites

Examples:
  aami clusters add prod-east --endpoint https://aami-east.example.com
  aami clusters list
  aami clusters status
  aami clusters alerts`,
}

var clustersAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a remote cluster",
	Long: `Add a remote AAMI cluster to the registry.

The cluster must be running AAMI with the API server enabled.

Examples:
  aami clusters add prod-east --endpoint https://aami-east.example.com
  aami clusters add prod-west --endpoint https://aami-west.example.com --api-key secret123
  aami clusters add secure --endpoint https://secure.example.com --tls-cert client.crt --tls-key client.key`,
	Args: cobra.ExactArgs(1),
	RunE: runClustersAdd,
}

var clustersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered clusters",
	RunE:  runClustersList,
}

var clustersRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a cluster from the registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runClustersRemove,
}

var clustersStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show aggregated status of all clusters",
	RunE:  runClustersStatus,
}

var clustersAlertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Show alerts from all clusters",
	RunE:  runClustersAlerts,
}

var clustersTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Test connection to a cluster",
	Long: `Test connection to a cluster or all clusters.

If no cluster name is provided, tests all registered clusters.

Examples:
  aami clusters test              # Test all clusters
  aami clusters test prod-east    # Test specific cluster`,
	Args: cobra.MaximumNArgs(1),
	RunE: runClustersTest,
}

var clustersInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  runClustersInfo,
}

// Flags
var (
	clusterEndpoint  string
	clusterAPIKey    string
	clusterTLSCert   string
	clusterTLSKey    string
	clusterTLSCACert string
	clusterSkipTLS   bool
	clusterLabels    []string
	alertsSeverity   string
	alertsLimit      int
)

func init() {
	// Add flags
	clustersAddCmd.Flags().StringVar(&clusterEndpoint, "endpoint", "",
		"Cluster API endpoint (required)")
	clustersAddCmd.Flags().StringVar(&clusterAPIKey, "api-key", "",
		"API key for authentication")
	clustersAddCmd.Flags().StringVar(&clusterTLSCert, "tls-cert", "",
		"Path to TLS client certificate")
	clustersAddCmd.Flags().StringVar(&clusterTLSKey, "tls-key", "",
		"Path to TLS client key")
	clustersAddCmd.Flags().StringVar(&clusterTLSCACert, "tls-ca", "",
		"Path to TLS CA certificate")
	clustersAddCmd.Flags().BoolVar(&clusterSkipTLS, "skip-tls-verify", false,
		"Skip TLS certificate verification")
	clustersAddCmd.Flags().StringSliceVar(&clusterLabels, "label", nil,
		"Labels for the cluster (key=value)")
	clustersAddCmd.MarkFlagRequired("endpoint")

	// Alerts flags
	clustersAlertsCmd.Flags().StringVar(&alertsSeverity, "severity", "",
		"Filter by severity (critical, warning, info)")
	clustersAlertsCmd.Flags().IntVar(&alertsLimit, "limit", 50,
		"Maximum number of alerts to show")

	// Add subcommands
	clustersCmd.AddCommand(clustersAddCmd)
	clustersCmd.AddCommand(clustersListCmd)
	clustersCmd.AddCommand(clustersRemoveCmd)
	clustersCmd.AddCommand(clustersStatusCmd)
	clustersCmd.AddCommand(clustersAlertsCmd)
	clustersCmd.AddCommand(clustersTestCmd)
	clustersCmd.AddCommand(clustersInfoCmd)
	rootCmd.AddCommand(clustersCmd)
}

func getRegistry() (*multicluster.Registry, error) {
	registryPath := "/etc/aami/clusters.yaml"
	registry := multicluster.NewRegistry(registryPath)
	if err := registry.Load(); err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}
	return registry, nil
}

func runClustersAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	registry, err := getRegistry()
	if err != nil {
		return err
	}

	// Parse labels
	labels := make(map[string]string)
	for _, l := range clusterLabels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 2 {
			labels[parts[0]] = parts[1]
		}
	}

	cluster := multicluster.ClusterConfig{
		Name:      name,
		Endpoint:  clusterEndpoint,
		APIKey:    clusterAPIKey,
		TLSCert:   clusterTLSCert,
		TLSKey:    clusterTLSKey,
		TLSCACert: clusterTLSCACert,
		SkipTLS:   clusterSkipTLS,
		Labels:    labels,
	}

	// Test connection
	fmt.Printf("Testing connection to %s...\n", clusterEndpoint)
	client, err := multicluster.NewClient(cluster)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := client.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	if !status.Connected {
		return fmt.Errorf("connection failed: %s", status.Error)
	}

	// Add to registry
	if err := registry.Add(cluster); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Cluster %s added\n", green("✓"), name)
	fmt.Printf("  Endpoint: %s\n", clusterEndpoint)
	fmt.Printf("  Nodes: %d\n", status.Nodes)
	fmt.Printf("  Health: %.0f%%\n", status.HealthScore)

	return nil
}

func runClustersList(cmd *cobra.Command, args []string) error {
	registry, err := getRegistry()
	if err != nil {
		return err
	}

	clusters := registry.List()
	if len(clusters) == 0 {
		fmt.Println("No clusters registered.")
		fmt.Println("\nUse 'aami clusters add <name> --endpoint <url>' to add a cluster.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Endpoint", "Labels"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, c := range clusters {
		labelStr := ""
		if len(c.Labels) > 0 {
			var parts []string
			for k, v := range c.Labels {
				parts = append(parts, fmt.Sprintf("%s=%s", k, v))
			}
			labelStr = strings.Join(parts, ", ")
		}
		table.Append([]string{c.Name, c.Endpoint, labelStr})
	}

	table.Render()
	return nil
}

func runClustersRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	registry, err := getRegistry()
	if err != nil {
		return err
	}

	if !registry.Exists(name) {
		return fmt.Errorf("cluster not found: %s", name)
	}

	if err := registry.Remove(name); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Cluster %s removed\n", green("✓"), name)
	return nil
}

func runClustersStatus(cmd *cobra.Command, args []string) error {
	registry, err := getRegistry()
	if err != nil {
		return err
	}

	clusters := registry.List()
	if len(clusters) == 0 {
		fmt.Println("No clusters registered.")
		return nil
	}

	aggregator := multicluster.NewAggregator(registry)
	if err := aggregator.Initialize(); err != nil {
		return err
	}
	defer aggregator.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Multi-Cluster Status")
	fmt.Println(strings.Repeat("━", 70))

	statuses, err := aggregator.GetAggregatedStatus(ctx)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Cluster", "Nodes", "GPUs", "Health", "Alerts", "Status"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
	})

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	var totalNodes, totalGPUs, totalAlerts int
	var totalHealth float64
	var connectedCount int

	for _, status := range statuses {
		statusStr := ""
		healthStr := ""

		if status.Connected {
			connectedCount++
			totalNodes += status.Nodes
			totalGPUs += status.TotalGPUs
			totalAlerts += status.AlertsActive
			totalHealth += status.HealthScore

			if status.HealthScore >= 90 {
				statusStr = green("●") + " Connected"
				healthStr = green(fmt.Sprintf("%.0f%%", status.HealthScore))
			} else if status.HealthScore >= 70 {
				statusStr = yellow("●") + " Connected"
				healthStr = yellow(fmt.Sprintf("%.0f%%", status.HealthScore))
			} else {
				statusStr = red("●") + " Connected"
				healthStr = red(fmt.Sprintf("%.0f%%", status.HealthScore))
			}
		} else {
			statusStr = red("○") + " Offline"
			healthStr = "-"
		}

		alertStr := "-"
		if status.Connected {
			if status.AlertsActive > 0 {
				alertStr = red(fmt.Sprintf("%d", status.AlertsActive))
			} else {
				alertStr = green("0")
			}
		}

		table.Append([]string{
			status.Name,
			fmt.Sprintf("%d", status.Nodes),
			fmt.Sprintf("%d", status.TotalGPUs),
			healthStr,
			alertStr,
			statusStr,
		})
	}

	table.Render()

	// Summary
	fmt.Println()
	fmt.Printf("Total: %d clusters, %d nodes, %d GPUs, %d active alerts\n",
		len(statuses), totalNodes, totalGPUs, totalAlerts)

	if connectedCount > 0 {
		avgHealth := totalHealth / float64(connectedCount)
		fmt.Printf("Average Health: %.0f%%\n", avgHealth)
	}

	if connectedCount < len(statuses) {
		fmt.Printf("\n%s %d cluster(s) offline\n",
			red("⚠"),
			len(statuses)-connectedCount)
	}

	return nil
}

func runClustersAlerts(cmd *cobra.Command, args []string) error {
	registry, err := getRegistry()
	if err != nil {
		return err
	}

	aggregator := multicluster.NewAggregator(registry)
	if err := aggregator.Initialize(); err != nil {
		return err
	}
	defer aggregator.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	alerts, err := aggregator.GetAllAlerts(ctx)
	if err != nil {
		return err
	}

	// Filter by severity if specified
	if alertsSeverity != "" {
		var filtered []multicluster.GlobalAlert
		for _, a := range alerts {
			if strings.EqualFold(a.Severity, alertsSeverity) {
				filtered = append(filtered, a)
			}
		}
		alerts = filtered
	}

	// Limit
	if len(alerts) > alertsLimit {
		alerts = alerts[:alertsLimit]
	}

	if len(alerts) == 0 {
		fmt.Println("No active alerts across all clusters.")
		return nil
	}

	fmt.Printf("Active Alerts (%d)\n", len(alerts))
	fmt.Println(strings.Repeat("━", 80))

	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for _, alert := range alerts {
		severityStr := alert.Severity
		switch alert.Severity {
		case "critical":
			severityStr = red("CRITICAL")
		case "warning":
			severityStr = yellow("WARNING")
		case "info":
			severityStr = cyan("INFO")
		}

		fmt.Printf("[%s] %s\n", severityStr, alert.AlertName)
		fmt.Printf("  Cluster: %s | Node: %s\n", alert.Cluster, alert.Node)
		fmt.Printf("  %s\n", alert.Description)
		fmt.Printf("  Since: %s\n", alert.FiredAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

func runClustersTest(cmd *cobra.Command, args []string) error {
	registry, err := getRegistry()
	if err != nil {
		return err
	}

	var clustersToTest []multicluster.ClusterConfig

	if len(args) > 0 {
		// Test specific cluster
		cluster, ok := registry.Get(args[0])
		if !ok {
			return fmt.Errorf("cluster not found: %s", args[0])
		}
		clustersToTest = append(clustersToTest, cluster)
	} else {
		// Test all clusters
		clustersToTest = registry.List()
	}

	if len(clustersToTest) == 0 {
		fmt.Println("No clusters to test.")
		return nil
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Println("Testing cluster connections...")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, cfg := range clustersToTest {
		client, err := multicluster.NewClient(cfg)
		if err != nil {
			fmt.Printf("%s %s: Failed to create client: %v\n", red("✗"), cfg.Name, err)
			continue
		}

		err = client.TestConnection(ctx)
		client.Close()

		if err != nil {
			fmt.Printf("%s %s: %v\n", red("✗"), cfg.Name, err)
		} else {
			fmt.Printf("%s %s: Connection successful\n", green("✓"), cfg.Name)
		}
	}

	return nil
}

func runClustersInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	registry, err := getRegistry()
	if err != nil {
		return err
	}

	cfg, ok := registry.Get(name)
	if !ok {
		return fmt.Errorf("cluster not found: %s", name)
	}

	client, err := multicluster.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status, err := client.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("Cluster: %s\n", cyan(name))
	fmt.Println(strings.Repeat("━", 50))

	fmt.Printf("Endpoint:    %s\n", cfg.Endpoint)

	if status.Connected {
		fmt.Printf("Status:      %s\n", green("Connected"))
		fmt.Printf("Version:     %s\n", status.Version)
		fmt.Printf("Nodes:       %d (%d healthy)\n", status.Nodes, status.HealthyNodes)
		fmt.Printf("GPUs:        %d (%d healthy)\n", status.TotalGPUs, status.HealthyGPUs)

		healthStr := fmt.Sprintf("%.0f%%", status.HealthScore)
		if status.HealthScore >= 90 {
			healthStr = green(healthStr)
		} else if status.HealthScore >= 70 {
			healthStr = yellow(healthStr)
		} else {
			healthStr = red(healthStr)
		}
		fmt.Printf("Health:      %s\n", healthStr)

		alertStr := fmt.Sprintf("%d", status.AlertsActive)
		if status.AlertsActive > 0 {
			alertStr = red(alertStr)
		} else {
			alertStr = green(alertStr)
		}
		fmt.Printf("Alerts:      %s\n", alertStr)
		fmt.Printf("Last Sync:   %s\n", status.LastSync.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Status:      %s\n", red("Offline"))
		if status.Error != "" {
			fmt.Printf("Error:       %s\n", status.Error)
		}
	}

	if len(cfg.Labels) > 0 {
		fmt.Println("\nLabels:")
		for k, v := range cfg.Labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	return nil
}
