package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/nvlink"
	"github.com/fregataa/aami/internal/ssh"
)

var (
	topologyOutput   string
	topologyNoColor  bool
	topologyShowLegend bool
)

var topologyCmd = &cobra.Command{
	Use:   "topology [node|all]",
	Short: "Display NVLink topology",
	Long: `Display NVLink topology for GPU nodes.

Shows the interconnection topology between GPUs including NVLink,
PCIe connections, and P2P capabilities.

Examples:
  aami topology gpu-node-01     # Show topology for a specific node
  aami topology all             # Show topology for all nodes
  aami topology --legend        # Show with connection legend`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTopology,
}

func init() {
	rootCmd.AddCommand(topologyCmd)

	topologyCmd.Flags().StringVarP(&topologyOutput, "output", "o", "ascii",
		"Output format: ascii, table, json")
	topologyCmd.Flags().BoolVar(&topologyNoColor, "no-color", false,
		"Disable colored output")
	topologyCmd.Flags().BoolVar(&topologyShowLegend, "legend", false,
		"Show connection type legend")
}

func runTopology(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("no nodes configured. Use 'aami nodes add' to add nodes")
	}

	// Create SSH executor
	executor := ssh.NewExecutorFromConfig(
		cfg.SSH.MaxParallel,
		cfg.SSH.ConnectTimeout,
		cfg.SSH.CommandTimeout,
		cfg.SSH.Retry.MaxAttempts,
		cfg.SSH.Retry.BackoffBase,
		cfg.SSH.Retry.BackoffMax,
	)

	collector := nvlink.NewCollector(executor)

	// Register all nodes with the collector
	for _, node := range cfg.Nodes {
		port := node.SSHPort
		if port == 0 {
			port = 22
		}
		collector.AddNode(node.Name, node.IP, port, node.SSHUser, node.SSHKey)
	}

	renderer := nvlink.NewRenderer(!topologyNoColor && !color.NoColor)

	// Show legend if requested
	if topologyShowLegend {
		fmt.Println(renderer.RenderConnectionLegend())
		fmt.Println()
	}

	// Determine target nodes
	var targetNodes []string
	if len(args) == 0 || args[0] == "all" {
		for _, node := range cfg.Nodes {
			targetNodes = append(targetNodes, node.IP)
		}
	} else {
		nodeName := args[0]
		found := false
		for _, node := range cfg.Nodes {
			if node.Name == nodeName || node.IP == nodeName {
				targetNodes = append(targetNodes, node.IP)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("node not found: %s", nodeName)
		}
	}

	// Collect and render
	if len(targetNodes) == 1 {
		return renderSingleNode(collector, renderer, targetNodes[0])
	}

	return renderCluster(collector, renderer, targetNodes)
}

func renderSingleNode(collector *nvlink.Collector, renderer *nvlink.Renderer, host string) error {
	fmt.Printf("Collecting topology from %s...\n\n", host)

	topology, err := collector.CollectTopology(host)
	if err != nil {
		return fmt.Errorf("failed to collect topology: %w", err)
	}

	switch topologyOutput {
	case "ascii":
		fmt.Println(renderer.RenderTopology(topology))
	case "table":
		renderTopologyTable(topology)
	case "json":
		return renderJSON(topology)
	default:
		return fmt.Errorf("unknown output format: %s", topologyOutput)
	}

	return nil
}

func renderCluster(collector *nvlink.Collector, renderer *nvlink.Renderer, hosts []string) error {
	fmt.Printf("Collecting topology from %d nodes...\n\n", len(hosts))

	cluster, err := collector.CollectClusterTopology(hosts)
	if err != nil {
		return fmt.Errorf("failed to collect cluster topology: %w", err)
	}

	switch topologyOutput {
	case "ascii":
		fmt.Println(renderer.RenderClusterSummary(cluster))
		fmt.Println()
		for _, node := range cluster.Nodes {
			fmt.Println(renderer.RenderTopology(&node))
			fmt.Println()
		}
	case "table":
		renderClusterTable(cluster)
	case "json":
		return renderJSON(cluster)
	default:
		return fmt.Errorf("unknown output format: %s", topologyOutput)
	}

	return nil
}

func renderTopologyTable(topology *nvlink.NodeTopology) {
	fmt.Printf("Node: %s\n", topology.NodeName)
	fmt.Printf("Collected: %s\n\n", topology.CollectedAt)

	// GPU table
	fmt.Println("GPUs:")
	gpuTable := tablewriter.NewWriter(os.Stdout)
	gpuTable.SetHeader([]string{"Index", "Name", "Bus ID", "UUID"})
	gpuTable.SetBorder(false)
	for _, gpu := range topology.GPUs {
		gpuTable.Append([]string{
			fmt.Sprintf("%d", gpu.Index),
			gpu.Name,
			gpu.BusID,
			truncateUUID(gpu.UUID),
		})
	}
	gpuTable.Render()
	fmt.Println()

	// Connection matrix
	fmt.Println("Connection Matrix:")
	gpuCount := len(topology.GPUs)
	matrixTable := tablewriter.NewWriter(os.Stdout)

	headers := []string{""}
	for i := 0; i < gpuCount; i++ {
		headers = append(headers, fmt.Sprintf("GPU%d", i))
	}
	matrixTable.SetHeader(headers)
	matrixTable.SetBorder(false)

	// Build matrix from P2P capabilities
	matrix := make([][]string, gpuCount)
	for i := range matrix {
		matrix[i] = make([]string, gpuCount)
		for j := range matrix[i] {
			if i == j {
				matrix[i][j] = "X"
			} else {
				matrix[i][j] = "-"
			}
		}
	}

	for _, p2p := range topology.P2PMatrix {
		matrix[p2p.GPU1][p2p.GPU2] = p2p.Connection
		matrix[p2p.GPU2][p2p.GPU1] = p2p.Connection
	}

	for i := 0; i < gpuCount; i++ {
		row := []string{fmt.Sprintf("GPU%d", i)}
		row = append(row, matrix[i]...)
		matrixTable.Append(row)
	}
	matrixTable.Render()
	fmt.Println()

	// Health summary
	health := topology.GetHealthStatus()
	fmt.Println("Health Summary:")
	fmt.Printf("  Total Links:  %d\n", health.TotalLinks)
	fmt.Printf("  Active Links: %d\n", health.ActiveLinks)
	fmt.Printf("  Status:       %s (%.1f%%)\n", health.Status, health.HealthPercent)
}

func renderClusterTable(cluster *nvlink.ClusterTopology) {
	fmt.Println("Cluster NVLink Topology Summary")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total Nodes:  %d\n", len(cluster.Nodes))
	fmt.Printf("Total GPUs:   %d\n", cluster.TotalGPUs)
	fmt.Printf("Total Links:  %d\n", cluster.TotalLinks)
	fmt.Printf("Active Links: %d\n", cluster.ActiveLinks)
	fmt.Printf("Error Links:  %d\n", cluster.ErrorLinks)
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Node", "GPUs", "Active/Total", "Status"})
	table.SetBorder(false)

	for _, node := range cluster.Nodes {
		health := node.GetHealthStatus()
		table.Append([]string{
			node.NodeName,
			fmt.Sprintf("%d", len(node.GPUs)),
			fmt.Sprintf("%d/%d", health.ActiveLinks, health.TotalLinks),
			health.Status,
		})
	}
	table.Render()
}

func truncateUUID(uuid string) string {
	if len(uuid) > 12 {
		return uuid[:12] + "..."
	}
	return uuid
}

func renderJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}
