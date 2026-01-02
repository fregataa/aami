package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/health"
)

var (
	healthOutput   string
	healthNoColor  bool
	healthDetailed bool
)

var healthCmd = &cobra.Command{
	Use:   "health [node]",
	Short: "Display GPU health scores",
	Long: `Display GPU health scores for the cluster or a specific node.

Health scores are calculated based on:
  - Temperature (20%)
  - ECC Errors (25%)
  - Xid Errors (25%)
  - NVLink Status (15%)
  - Uptime (15%)

Examples:
  aami health              # Show cluster health summary
  aami health gpu-node-01  # Show detailed health for a node
  aami health --detailed   # Show all component scores`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.Flags().StringVarP(&healthOutput, "output", "o", "table",
		"Output format: table, json")
	healthCmd.Flags().BoolVar(&healthNoColor, "no-color", false,
		"Disable colored output")
	healthCmd.Flags().BoolVar(&healthDetailed, "detailed", false,
		"Show detailed component scores")
}

func runHealth(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Build Prometheus URL
	promURL := fmt.Sprintf("http://localhost:%d", cfg.Prometheus.Port)
	if cfg.Prometheus.Port == 0 {
		promURL = "http://localhost:9090"
	}

	// Create clients
	promClient := health.NewPrometheusClient(promURL)
	calculator := health.NewCalculator()

	// Check Prometheus connection
	if err := promClient.CheckConnection(); err != nil {
		return fmt.Errorf("cannot connect to Prometheus at %s: %w", promURL, err)
	}

	// Collect metrics
	nodeMetrics, err := promClient.CollectAllMetrics()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	if len(nodeMetrics) == 0 {
		fmt.Println("No GPU metrics found. Ensure DCGM exporter is running.")
		return nil
	}

	// Filter by node if specified
	if len(args) > 0 {
		nodeName := args[0]
		var filtered []health.NodeMetrics
		for _, n := range nodeMetrics {
			if strings.Contains(n.NodeName, nodeName) || strings.Contains(n.NodeIP, nodeName) {
				filtered = append(filtered, n)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("node not found: %s", nodeName)
		}
		nodeMetrics = filtered
	}

	// Calculate health
	clusterHealth := calculator.CalculateClusterHealth(nodeMetrics)

	// Render output
	switch healthOutput {
	case "json":
		return renderHealthJSON(clusterHealth)
	case "table":
		if len(args) > 0 || healthDetailed {
			renderDetailedHealth(clusterHealth, !healthNoColor && !color.NoColor)
		} else {
			renderClusterHealth(clusterHealth, !healthNoColor && !color.NoColor)
		}
	default:
		return fmt.Errorf("unknown output format: %s", healthOutput)
	}

	return nil
}

func renderHealthJSON(cluster health.ClusterHealth) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cluster)
}

func renderClusterHealth(cluster health.ClusterHealth, useColor bool) {
	// Header
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║              GPU Cluster Health Summary            ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
	fmt.Println()

	// Overall score
	scoreStr := fmt.Sprintf("%.1f", cluster.OverallScore)
	statusStr := cluster.Status
	if useColor {
		scoreStr = colorScore(cluster.OverallScore, scoreStr)
		statusStr = colorStatus(cluster.Status)
	}
	fmt.Printf("  Overall Score: %s / 100  [%s]\n", scoreStr, statusStr)
	fmt.Println()

	// Summary stats
	fmt.Printf("  Total GPUs:    %d\n", cluster.TotalGPUs)
	if useColor {
		fmt.Printf("  Healthy:       %s\n", color.GreenString("%d", cluster.HealthyGPUs))
		if cluster.WarningGPUs > 0 {
			fmt.Printf("  Warning:       %s\n", color.YellowString("%d", cluster.WarningGPUs))
		}
		if cluster.CriticalGPUs > 0 {
			fmt.Printf("  Critical:      %s\n", color.RedString("%d", cluster.CriticalGPUs))
		}
	} else {
		fmt.Printf("  Healthy:       %d\n", cluster.HealthyGPUs)
		fmt.Printf("  Warning:       %d\n", cluster.WarningGPUs)
		fmt.Printf("  Critical:      %d\n", cluster.CriticalGPUs)
	}
	fmt.Println()

	// Node summary table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Node", "GPUs", "Score", "Status", "Issues"})
	table.SetBorder(false)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_LEFT,
	})

	for _, node := range cluster.Nodes {
		issues := []string{}
		if node.WarningGPUs > 0 {
			issues = append(issues, fmt.Sprintf("%d warning", node.WarningGPUs))
		}
		if node.CriticalGPUs > 0 {
			issues = append(issues, fmt.Sprintf("%d critical", node.CriticalGPUs))
		}
		issueStr := "-"
		if len(issues) > 0 {
			issueStr = strings.Join(issues, ", ")
		}

		scoreStr := fmt.Sprintf("%.1f", node.OverallScore)
		statusStr := node.Status
		if useColor {
			scoreStr = colorScore(node.OverallScore, scoreStr)
			statusStr = colorStatus(node.Status)
		}

		table.Append([]string{
			node.NodeName,
			fmt.Sprintf("%d", len(node.GPUs)),
			scoreStr,
			statusStr,
			issueStr,
		})
	}
	table.Render()

	fmt.Println()
	fmt.Printf("  Use 'aami health <node>' for detailed view\n")
	fmt.Printf("  Collected at: %s\n", cluster.CollectedAt.Format("2006-01-02 15:04:05"))
}

func renderDetailedHealth(cluster health.ClusterHealth, useColor bool) {
	for _, node := range cluster.Nodes {
		// Node header
		fmt.Println()
		fmt.Printf("═══ %s ═══\n", node.NodeName)
		fmt.Printf("Score: %.1f / 100  Status: %s\n\n",
			node.OverallScore, colorStatus(node.Status))

		// GPU table with component details
		for _, gpu := range node.GPUs {
			fmt.Printf("  GPU %d: %s\n", gpu.Index, gpu.Name)
			scoreStr := fmt.Sprintf("%.1f", gpu.OverallScore)
			if useColor {
				scoreStr = colorScore(gpu.OverallScore, scoreStr)
			}
			fmt.Printf("  Score: %s  Status: %s\n", scoreStr, colorStatus(gpu.Status))
			fmt.Println()

			// Component scores
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Component", "Score", "Weight", "Status", "Details"})
			table.SetBorder(false)
			table.SetColumnAlignment([]int{
				tablewriter.ALIGN_LEFT,
				tablewriter.ALIGN_RIGHT,
				tablewriter.ALIGN_RIGHT,
				tablewriter.ALIGN_CENTER,
				tablewriter.ALIGN_LEFT,
			})

			for _, comp := range gpu.Components {
				scoreStr := fmt.Sprintf("%.0f", comp.Score)
				weightStr := fmt.Sprintf("%.0f%%", comp.Weight*100)
				statusStr := comp.Status
				if useColor {
					scoreStr = colorScore(comp.Score, scoreStr)
					statusStr = colorStatus(comp.Status)
				}

				table.Append([]string{
					comp.Name,
					scoreStr,
					weightStr,
					statusStr,
					comp.Message,
				})
			}
			table.Render()
			fmt.Println()
		}
	}

	fmt.Printf("Collected at: %s\n", cluster.CollectedAt.Format("2006-01-02 15:04:05"))
}

func colorScore(score float64, text string) string {
	switch {
	case score >= 80:
		return color.GreenString(text)
	case score >= 50:
		return color.YellowString(text)
	default:
		return color.RedString(text)
	}
}

func colorStatus(status string) string {
	switch status {
	case health.StatusHealthy:
		return color.GreenString(status)
	case health.StatusWarning:
		return color.YellowString(status)
	case health.StatusCritical:
		return color.RedString(status)
	default:
		return status
	}
}
