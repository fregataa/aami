package nvlink

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fregataa/aami/internal/ssh"
)

// Collector handles NVLink topology collection from nodes.
type Collector struct {
	executor *ssh.Executor
	nodes    map[string]ssh.Node // host -> node mapping
}

// NewCollector creates a new topology collector.
func NewCollector(executor *ssh.Executor) *Collector {
	return &Collector{
		executor: executor,
		nodes:    make(map[string]ssh.Node),
	}
}

// AddNode registers a node for SSH access.
func (c *Collector) AddNode(name, host string, port int, user, keyPath string) {
	c.nodes[host] = ssh.Node{
		Name:    name,
		Host:    host,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
	}
}

// runCommand executes a command on the specified host.
func (c *Collector) runCommand(host, cmd string, timeout time.Duration) (string, error) {
	node, ok := c.nodes[host]
	if !ok {
		return "", fmt.Errorf("node not registered: %s", host)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := c.executor.Run(ctx, node, cmd)
	if result.Error != nil {
		return "", result.Error
	}
	return result.Output, nil
}

// CollectTopology collects NVLink topology from a single node.
func (c *Collector) CollectTopology(host string) (*NodeTopology, error) {
	topology := &NodeTopology{
		NodeName:    host,
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Collect GPU list
	gpus, err := c.collectGPUs(host)
	if err != nil {
		return nil, fmt.Errorf("failed to collect GPUs: %w", err)
	}
	topology.GPUs = gpus

	// Collect topology matrix
	matrix, err := c.collectTopologyMatrix(host, len(gpus))
	if err != nil {
		return nil, fmt.Errorf("failed to collect topology matrix: %w", err)
	}
	topology.P2PMatrix = c.parseP2PMatrix(matrix, len(gpus))

	// Collect NVLink status
	links, err := c.collectNVLinks(host, len(gpus))
	if err != nil {
		// NVLink may not be available on all GPUs
		topology.Links = []NVLinkInfo{}
	} else {
		topology.Links = links
	}

	// Calculate link statistics
	c.calculateLinkStats(topology)

	return topology, nil
}

// collectGPUs runs nvidia-smi to get GPU list.
func (c *Collector) collectGPUs(host string) ([]GPUInfo, error) {
	cmd := "nvidia-smi --query-gpu=index,uuid,name,pci.bus_id --format=csv,noheader,nounits"
	output, err := c.runCommand(host, cmd, 30*time.Second)
	if err != nil {
		return nil, err
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		if len(parts) < 4 {
			continue
		}

		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		gpus = append(gpus, GPUInfo{
			Index: idx,
			UUID:  strings.TrimSpace(parts[1]),
			Name:  strings.TrimSpace(parts[2]),
			BusID: strings.TrimSpace(parts[3]),
		})
	}

	return gpus, nil
}

// collectTopologyMatrix runs nvidia-smi topo to get connection matrix.
func (c *Collector) collectTopologyMatrix(host string, gpuCount int) ([][]string, error) {
	cmd := "nvidia-smi topo -m"
	output, err := c.runCommand(host, cmd, 30*time.Second)
	if err != nil {
		return nil, err
	}

	return c.parseTopologyOutput(output, gpuCount)
}

// parseTopologyOutput parses nvidia-smi topo -m output.
func (c *Collector) parseTopologyOutput(output string, gpuCount int) ([][]string, error) {
	lines := strings.Split(output, "\n")
	matrix := make([][]string, gpuCount)
	for i := range matrix {
		matrix[i] = make([]string, gpuCount)
	}

	gpuLineRegex := regexp.MustCompile(`^GPU(\d+)`)

	for _, line := range lines {
		matches := gpuLineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		gpuIdx, _ := strconv.Atoi(matches[1])
		if gpuIdx >= gpuCount {
			continue
		}

		// Parse the row - split by whitespace
		fields := strings.Fields(line)
		if len(fields) < gpuCount+1 {
			continue
		}

		// First field is GPUx, rest are connections
		for i := 0; i < gpuCount && i+1 < len(fields); i++ {
			conn := fields[i+1]
			matrix[gpuIdx][i] = c.normalizeConnection(conn)
		}
	}

	return matrix, nil
}

// normalizeConnection normalizes connection type strings.
func (c *Collector) normalizeConnection(conn string) string {
	conn = strings.ToUpper(strings.TrimSpace(conn))
	switch {
	case strings.HasPrefix(conn, "NV"):
		return ConnNVLink
	case conn == "PIX":
		return ConnPIX
	case conn == "PXB":
		return ConnPXB
	case conn == "PHB":
		return ConnPHB
	case conn == "SYS":
		return ConnSYS
	case conn == "X":
		return ConnSelf
	default:
		return ConnNone
	}
}

// parseP2PMatrix converts topology matrix to P2P capability list.
func (c *Collector) parseP2PMatrix(matrix [][]string, gpuCount int) []P2PCapability {
	var capabilities []P2PCapability

	for i := 0; i < gpuCount; i++ {
		for j := i + 1; j < gpuCount; j++ {
			if i >= len(matrix) || j >= len(matrix[i]) {
				continue
			}
			conn := matrix[i][j]
			cap := P2PCapability{
				GPU1:       i,
				GPU2:       j,
				Connection: conn,
			}
			// NVLink provides full P2P capabilities
			if conn == ConnNVLink {
				cap.P2PRead = true
				cap.P2PWrite = true
				cap.P2PAtomic = true
			} else if conn == ConnPIX || conn == ConnPXB {
				cap.P2PRead = true
				cap.P2PWrite = true
				cap.P2PAtomic = false
			}
			capabilities = append(capabilities, cap)
		}
	}

	return capabilities
}

// collectNVLinks collects detailed NVLink information.
func (c *Collector) collectNVLinks(host string, gpuCount int) ([]NVLinkInfo, error) {
	var links []NVLinkInfo

	for gpuIdx := 0; gpuIdx < gpuCount; gpuIdx++ {
		gpuLinks, err := c.collectGPUNVLinks(host, gpuIdx)
		if err != nil {
			continue // Skip GPUs without NVLink
		}
		links = append(links, gpuLinks...)
	}

	return links, nil
}

// collectGPUNVLinks collects NVLink status for a specific GPU.
func (c *Collector) collectGPUNVLinks(host string, gpuIdx int) ([]NVLinkInfo, error) {
	cmd := fmt.Sprintf("nvidia-smi nvlink -s -i %d 2>/dev/null", gpuIdx)
	output, err := c.runCommand(host, cmd, 30*time.Second)
	if err != nil {
		return nil, err
	}

	return c.parseNVLinkStatus(output, gpuIdx)
}

// parseNVLinkStatus parses nvidia-smi nvlink -s output.
func (c *Collector) parseNVLinkStatus(output string, sourceGPU int) ([]NVLinkInfo, error) {
	var links []NVLinkInfo
	lines := strings.Split(output, "\n")

	linkRegex := regexp.MustCompile(`Link\s+(\d+):\s*(.*)`)

	for _, line := range lines {
		matches := linkRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		linkIdx, _ := strconv.Atoi(matches[1])
		statusStr := strings.ToLower(matches[2])

		link := NVLinkInfo{
			SourceGPU: sourceGPU,
			LinkIndex: linkIdx,
			TargetGPU: -1, // Will be determined from topology
		}

		switch {
		case strings.Contains(statusStr, "active"):
			link.Status = LinkStatusActive
		case strings.Contains(statusStr, "inactive"):
			link.Status = LinkStatusInactive
		case strings.Contains(statusStr, "disabled"):
			link.Status = LinkStatusDisabled
		case strings.Contains(statusStr, "error"):
			link.Status = LinkStatusError
		default:
			link.Status = LinkStatusUnknown
		}

		links = append(links, link)
	}

	return links, nil
}

// calculateLinkStats calculates link statistics for topology.
func (c *Collector) calculateLinkStats(topology *NodeTopology) {
	total := 0
	active := 0
	errors := 0

	for _, link := range topology.Links {
		total++
		switch link.Status {
		case LinkStatusActive:
			active++
		case LinkStatusError:
			errors++
		}
	}

	// Also count NVLink connections from P2P matrix
	for _, p2p := range topology.P2PMatrix {
		if p2p.Connection == ConnNVLink {
			total++
			active++ // Assume active if in topology matrix
		}
	}

	topology.TotalLinks = total
	topology.ActiveLinks = active
	topology.ErrorLinks = errors
}

// CollectClusterTopology collects topology from multiple nodes.
func (c *Collector) CollectClusterTopology(hosts []string) (*ClusterTopology, error) {
	cluster := &ClusterTopology{}

	results := make(chan struct {
		topology *NodeTopology
		err      error
	}, len(hosts))

	for _, host := range hosts {
		go func(h string) {
			topo, err := c.CollectTopology(h)
			results <- struct {
				topology *NodeTopology
				err      error
			}{topo, err}
		}(host)
	}

	for range hosts {
		result := <-results
		if result.err != nil {
			continue // Skip failed nodes
		}
		cluster.Nodes = append(cluster.Nodes, *result.topology)
		cluster.TotalGPUs += len(result.topology.GPUs)
		cluster.TotalLinks += result.topology.TotalLinks
		cluster.ActiveLinks += result.topology.ActiveLinks
		cluster.ErrorLinks += result.topology.ErrorLinks
	}

	return cluster, nil
}
