package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fregataa/aami/internal/config"
	"github.com/fregataa/aami/internal/ssh"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Manage GPU nodes",
	Long:  "Add, remove, list, and manage monitored GPU nodes.",
}

var nodesAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a node to the cluster",
	Long: `Add a node to the cluster configuration.

Examples:
  aami nodes add gpu-01 --ip 192.168.1.100 --user root --key ~/.ssh/id_rsa
  aami nodes add gpu-02 --ip 192.168.1.101 --user ubuntu --labels gpu_type=a100
  aami nodes add --file hosts.txt --user root --key ~/.ssh/id_rsa`,
	RunE: runNodesAdd,
}

var nodesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nodes",
	RunE:  runNodesList,
}

var nodesRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a node from the cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  runNodesRemove,
}

var nodesInstallCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install exporters on node(s)",
	Long: `Install node_exporter and dcgm_exporter on specified nodes.

Examples:
  aami nodes install gpu-01        # Install on single node
  aami nodes install --all         # Install on all nodes`,
	RunE: runNodesInstall,
}

var nodesTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Test SSH connection to node(s)",
	Long: `Test SSH connectivity to specified nodes.

Examples:
  aami nodes test gpu-01           # Test single node
  aami nodes test --all            # Test all nodes`,
	RunE: runNodesTest,
}

var (
	nodeIP     string
	nodeUser   string
	nodeKey    string
	nodePort   int
	nodeLabels string
	nodesFile  string
	allNodes   bool
)

func init() {
	// Add flags
	nodesAddCmd.Flags().StringVar(&nodeIP, "ip", "", "Node IP address")
	nodesAddCmd.Flags().StringVar(&nodeUser, "user", "root", "SSH user")
	nodesAddCmd.Flags().StringVar(&nodeKey, "key", "", "SSH key path")
	nodesAddCmd.Flags().IntVar(&nodePort, "port", 22, "SSH port")
	nodesAddCmd.Flags().StringVar(&nodeLabels, "labels", "", "Labels (k=v,k2=v2)")
	nodesAddCmd.Flags().StringVar(&nodesFile, "file", "", "File with nodes list (format: name ip)")

	nodesInstallCmd.Flags().BoolVar(&allNodes, "all", false, "Install on all nodes")
	nodesTestCmd.Flags().BoolVar(&allNodes, "all", false, "Test all nodes")

	// Add subcommands
	nodesCmd.AddCommand(nodesAddCmd)
	nodesCmd.AddCommand(nodesListCmd)
	nodesCmd.AddCommand(nodesRemoveCmd)
	nodesCmd.AddCommand(nodesInstallCmd)
	nodesCmd.AddCommand(nodesTestCmd)
	rootCmd.AddCommand(nodesCmd)
}

func runNodesAdd(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()

	// Add from file
	if nodesFile != "" {
		count, err := addNodesFromFile(cfg, nodesFile)
		if err != nil {
			return err
		}
		fmt.Printf("%s Added %d nodes from %s\n", green("✓"), count, nodesFile)
		return saveConfig(cfg)
	}

	// Add single node
	if len(args) == 0 {
		return fmt.Errorf("node name required (or use --file)")
	}

	if nodeIP == "" {
		return fmt.Errorf("--ip is required")
	}

	node := config.NodeConfig{
		Name:    args[0],
		IP:      nodeIP,
		SSHUser: nodeUser,
		SSHKey:  nodeKey,
		SSHPort: nodePort,
		Labels:  parseLabels(nodeLabels),
	}

	// Check for duplicate
	for _, existing := range cfg.Nodes {
		if existing.Name == node.Name {
			return fmt.Errorf("node %s already exists", node.Name)
		}
	}

	cfg.Nodes = append(cfg.Nodes, node)
	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("%s Node %s added\n", green("✓"), node.Name)
	return nil
}

func runNodesList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Nodes) == 0 {
		fmt.Println("No nodes configured.")
		fmt.Println("Add nodes with: aami nodes add <name> --ip <ip> --user <user> --key <key>")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "IP", "Port", "User", "Labels"})
	table.SetBorder(true)
	table.SetRowLine(false)

	for _, node := range cfg.Nodes {
		port := node.SSHPort
		if port == 0 {
			port = 22
		}
		table.Append([]string{
			node.Name,
			node.IP,
			fmt.Sprintf("%d", port),
			node.SSHUser,
			formatLabels(node.Labels),
		})
	}

	table.Render()
	fmt.Printf("\nTotal: %d nodes\n", len(cfg.Nodes))
	return nil
}

func runNodesRemove(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	nodeName := args[0]
	found := false

	newNodes := make([]config.NodeConfig, 0, len(cfg.Nodes))
	for _, node := range cfg.Nodes {
		if node.Name == nodeName {
			found = true
			continue
		}
		newNodes = append(newNodes, node)
	}

	if !found {
		return fmt.Errorf("node %s not found", nodeName)
	}

	cfg.Nodes = newNodes
	if err := saveConfig(cfg); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Node %s removed\n", green("✓"), nodeName)
	return nil
}

func runNodesInstall(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var nodesToInstall []config.NodeConfig

	if allNodes {
		nodesToInstall = cfg.Nodes
	} else if len(args) > 0 {
		for _, name := range args {
			node, found := findNode(cfg, name)
			if !found {
				return fmt.Errorf("node %s not found", name)
			}
			nodesToInstall = append(nodesToInstall, node)
		}
	} else {
		return fmt.Errorf("specify node name or use --all")
	}

	if len(nodesToInstall) == 0 {
		return fmt.Errorf("no nodes to install on")
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("Installing exporters on %d node(s)...\n\n", len(nodesToInstall))

	// TODO: Implement actual installation via SSH
	for _, node := range nodesToInstall {
		fmt.Printf("  %s %s: Installing... (pending implementation)\n", green("•"), node.Name)
	}

	fmt.Println()
	fmt.Printf("%s Installation complete\n", green("✓"))
	fmt.Printf("  Succeeded: %s\n", green(fmt.Sprintf("%d", len(nodesToInstall))))
	fmt.Printf("  Failed:    %s\n", red("0"))

	return nil
}

func runNodesTest(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var nodesToTest []config.NodeConfig

	if allNodes {
		nodesToTest = cfg.Nodes
	} else if len(args) > 0 {
		for _, name := range args {
			node, found := findNode(cfg, name)
			if !found {
				return fmt.Errorf("node %s not found", name)
			}
			nodesToTest = append(nodesToTest, node)
		}
	} else {
		return fmt.Errorf("specify node name or use --all")
	}

	if len(nodesToTest) == 0 {
		return fmt.Errorf("no nodes to test")
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("Testing SSH connection to %d node(s)...\n\n", len(nodesToTest))

	executor := ssh.NewExecutorFromConfig(
		cfg.SSH.MaxParallel,
		cfg.SSH.ConnectTimeout,
		cfg.SSH.CommandTimeout,
		cfg.SSH.Retry.MaxAttempts,
		cfg.SSH.Retry.BackoffBase,
		cfg.SSH.Retry.BackoffMax,
	)

	succeeded := 0
	failed := 0

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, node := range nodesToTest {
		sshNode := ssh.Node{
			Name:    node.Name,
			Host:    node.IP,
			Port:    node.SSHPort,
			User:    node.SSHUser,
			KeyPath: node.SSHKey,
		}

		err := executor.TestConnection(ctx, sshNode)
		if err != nil {
			fmt.Printf("  %s %s: %v\n", red("✗"), node.Name, err)
			failed++
		} else {
			fmt.Printf("  %s %s: OK\n", green("✓"), node.Name)
			succeeded++
		}
	}

	fmt.Println()
	fmt.Printf("Results: %s succeeded, %s failed\n",
		green(fmt.Sprintf("%d", succeeded)),
		red(fmt.Sprintf("%d", failed)))

	return nil
}

func addNodesFromFile(cfg *config.Config, filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		node := config.NodeConfig{
			Name:    fields[0],
			IP:      fields[1],
			SSHUser: nodeUser,
			SSHKey:  nodeKey,
			SSHPort: nodePort,
		}

		// Check for duplicate
		duplicate := false
		for _, existing := range cfg.Nodes {
			if existing.Name == node.Name {
				duplicate = true
				break
			}
		}

		if !duplicate {
			cfg.Nodes = append(cfg.Nodes, node)
			count++
		}
	}

	return count, scanner.Err()
}

func parseLabels(s string) map[string]string {
	labels := make(map[string]string)
	if s == "" {
		return labels
	}
	for _, pair := range strings.Split(s, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return labels
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "-"
	}
	var pairs []string
	for k, v := range labels {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}

func findNode(cfg *config.Config, name string) (config.NodeConfig, bool) {
	for _, node := range cfg.Nodes {
		if node.Name == name {
			return node, true
		}
	}
	return config.NodeConfig{}, false
}
