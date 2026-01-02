package nvlink

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Renderer handles ASCII rendering of NVLink topology.
type Renderer struct {
	useColor bool
}

// NewRenderer creates a new topology renderer.
func NewRenderer(useColor bool) *Renderer {
	return &Renderer{useColor: useColor}
}

// RenderTopology renders the topology as ASCII art.
func (r *Renderer) RenderTopology(topology *NodeTopology) string {
	gpuCount := len(topology.GPUs)

	switch {
	case gpuCount <= 2:
		return r.render2GPU(topology)
	case gpuCount <= 4:
		return r.render4GPU(topology)
	case gpuCount <= 8:
		return r.render8GPU(topology)
	default:
		return r.renderMatrix(topology)
	}
}

// render2GPU renders a 2-GPU topology.
func (r *Renderer) render2GPU(topology *NodeTopology) string {
	var sb strings.Builder

	sb.WriteString(r.header(topology.NodeName))
	sb.WriteString("\n")

	// Simple 2-GPU layout
	//  [GPU 0] ═══NV═══ [GPU 1]

	conn := r.getConnection(topology, 0, 1)
	connStr := r.colorConnection(conn)

	sb.WriteString(fmt.Sprintf("  [GPU 0] ══%s══ [GPU 1]\n", connStr))
	sb.WriteString("\n")
	sb.WriteString(r.renderHealth(topology))

	return sb.String()
}

// render4GPU renders a 4-GPU topology (2x2 layout).
func (r *Renderer) render4GPU(topology *NodeTopology) string {
	var sb strings.Builder

	sb.WriteString(r.header(topology.NodeName))
	sb.WriteString("\n")

	// 4-GPU layout (typical DGX Station / 4x GPU server)
	//
	//  [GPU 0] ══NV══ [GPU 1]
	//     ║              ║
	//    NV             NV
	//     ║              ║
	//  [GPU 2] ══NV══ [GPU 3]

	conn01 := r.colorConnection(r.getConnection(topology, 0, 1))
	conn23 := r.colorConnection(r.getConnection(topology, 2, 3))
	conn02 := r.colorConnection(r.getConnection(topology, 0, 2))
	conn13 := r.colorConnection(r.getConnection(topology, 1, 3))

	sb.WriteString(fmt.Sprintf("  [GPU 0] ══%s══ [GPU 1]\n", conn01))
	sb.WriteString(fmt.Sprintf("     ║              ║\n"))
	sb.WriteString(fmt.Sprintf("    %s             %s\n", conn02, conn13))
	sb.WriteString(fmt.Sprintf("     ║              ║\n"))
	sb.WriteString(fmt.Sprintf("  [GPU 2] ══%s══ [GPU 3]\n", conn23))
	sb.WriteString("\n")

	// Diagonal connections if present
	conn03 := r.getConnection(topology, 0, 3)
	conn12 := r.getConnection(topology, 1, 2)
	if conn03 != ConnNone || conn12 != ConnNone {
		sb.WriteString("  Diagonal Links:\n")
		if conn03 != ConnNone {
			sb.WriteString(fmt.Sprintf("    GPU 0 ↔ GPU 3: %s\n", r.colorConnection(conn03)))
		}
		if conn12 != ConnNone {
			sb.WriteString(fmt.Sprintf("    GPU 1 ↔ GPU 2: %s\n", r.colorConnection(conn12)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(r.renderHealth(topology))

	return sb.String()
}

// render8GPU renders an 8-GPU topology (DGX-1/2 style).
func (r *Renderer) render8GPU(topology *NodeTopology) string {
	var sb strings.Builder

	sb.WriteString(r.header(topology.NodeName))
	sb.WriteString("\n")

	// 8-GPU DGX-style layout (simplified cube visualization)
	//
	//  [GPU 0]══NV══[GPU 1]══NV══[GPU 2]══NV══[GPU 3]
	//     ║           ║           ║           ║
	//    NV          NV          NV          NV
	//     ║           ║           ║           ║
	//  [GPU 4]══NV══[GPU 5]══NV══[GPU 6]══NV══[GPU 7]

	// Top row connections
	conn01 := r.colorConnection(r.getConnection(topology, 0, 1))
	conn12 := r.colorConnection(r.getConnection(topology, 1, 2))
	conn23 := r.colorConnection(r.getConnection(topology, 2, 3))

	// Bottom row connections
	conn45 := r.colorConnection(r.getConnection(topology, 4, 5))
	conn56 := r.colorConnection(r.getConnection(topology, 5, 6))
	conn67 := r.colorConnection(r.getConnection(topology, 6, 7))

	// Vertical connections
	conn04 := r.colorConnection(r.getConnection(topology, 0, 4))
	conn15 := r.colorConnection(r.getConnection(topology, 1, 5))
	conn26 := r.colorConnection(r.getConnection(topology, 2, 6))
	conn37 := r.colorConnection(r.getConnection(topology, 3, 7))

	sb.WriteString(fmt.Sprintf("  [GPU 0]═%s═[GPU 1]═%s═[GPU 2]═%s═[GPU 3]\n", conn01, conn12, conn23))
	sb.WriteString(fmt.Sprintf("     ║        ║        ║        ║\n"))
	sb.WriteString(fmt.Sprintf("    %s       %s       %s       %s\n", conn04, conn15, conn26, conn37))
	sb.WriteString(fmt.Sprintf("     ║        ║        ║        ║\n"))
	sb.WriteString(fmt.Sprintf("  [GPU 4]═%s═[GPU 5]═%s═[GPU 6]═%s═[GPU 7]\n", conn45, conn56, conn67))
	sb.WriteString("\n")

	// Show cross-connections summary
	sb.WriteString("  Cross-connections:\n")
	crossLinks := []struct{ g1, g2 int }{
		{0, 5}, {0, 6}, {1, 4}, {1, 6}, {1, 7},
		{2, 4}, {2, 5}, {2, 7}, {3, 5}, {3, 6},
	}
	for _, link := range crossLinks {
		conn := r.getConnection(topology, link.g1, link.g2)
		if conn == ConnNVLink {
			sb.WriteString(fmt.Sprintf("    GPU %d ↔ GPU %d: %s\n", link.g1, link.g2, r.colorConnection(conn)))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(r.renderHealth(topology))

	return sb.String()
}

// renderMatrix renders a general NxN matrix for any GPU count.
func (r *Renderer) renderMatrix(topology *NodeTopology) string {
	var sb strings.Builder

	sb.WriteString(r.header(topology.NodeName))
	sb.WriteString("\n")

	gpuCount := len(topology.GPUs)

	// Header row
	sb.WriteString("        ")
	for i := 0; i < gpuCount; i++ {
		sb.WriteString(fmt.Sprintf("GPU%-2d ", i))
	}
	sb.WriteString("\n")

	// Matrix rows
	for i := 0; i < gpuCount; i++ {
		sb.WriteString(fmt.Sprintf("  GPU%-2d ", i))
		for j := 0; j < gpuCount; j++ {
			if i == j {
				sb.WriteString("  X   ")
			} else {
				conn := r.getConnection(topology, i, j)
				sb.WriteString(fmt.Sprintf(" %-4s ", r.colorConnection(conn)))
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString(r.renderHealth(topology))

	return sb.String()
}

// getConnection returns the connection type between two GPUs.
func (r *Renderer) getConnection(topology *NodeTopology, gpu1, gpu2 int) string {
	for _, p2p := range topology.P2PMatrix {
		if (p2p.GPU1 == gpu1 && p2p.GPU2 == gpu2) ||
			(p2p.GPU1 == gpu2 && p2p.GPU2 == gpu1) {
			return p2p.Connection
		}
	}
	return ConnNone
}

// colorConnection returns a colored connection string.
func (r *Renderer) colorConnection(conn string) string {
	if !r.useColor {
		return conn
	}

	switch conn {
	case ConnNVLink:
		return color.GreenString(conn)
	case ConnPIX, ConnPXB:
		return color.YellowString(conn)
	case ConnPHB, ConnSYS:
		return color.RedString(conn)
	default:
		return conn
	}
}

// header returns a formatted header.
func (r *Renderer) header(nodeName string) string {
	line := strings.Repeat("═", 50)
	return fmt.Sprintf("╔%s╗\n║ NVLink Topology: %-30s ║\n╚%s╝", line, nodeName, line)
}

// renderHealth renders health summary.
func (r *Renderer) renderHealth(topology *NodeTopology) string {
	health := topology.GetHealthStatus()

	var sb strings.Builder
	sb.WriteString("  Health Summary:\n")
	sb.WriteString(fmt.Sprintf("    Total Links:    %d\n", health.TotalLinks))
	sb.WriteString(fmt.Sprintf("    Active Links:   %d\n", health.ActiveLinks))

	if health.InactiveLinks > 0 {
		sb.WriteString(fmt.Sprintf("    Inactive Links: %d\n", health.InactiveLinks))
	}
	if health.ErrorLinks > 0 {
		sb.WriteString(fmt.Sprintf("    Error Links:    %d\n", health.ErrorLinks))
	}

	statusStr := health.Status
	if r.useColor {
		switch health.Status {
		case "healthy":
			statusStr = color.GreenString(health.Status)
		case "degraded":
			statusStr = color.YellowString(health.Status)
		case "critical":
			statusStr = color.RedString(health.Status)
		}
	}
	sb.WriteString(fmt.Sprintf("    Status:         %s (%.1f%%)\n", statusStr, health.HealthPercent))

	return sb.String()
}

// RenderConnectionLegend renders a legend for connection types.
func (r *Renderer) RenderConnectionLegend() string {
	var sb strings.Builder

	sb.WriteString("  Connection Types:\n")
	sb.WriteString(fmt.Sprintf("    %s  - NVLink (fastest)\n", r.colorConnection(ConnNVLink)))
	sb.WriteString(fmt.Sprintf("    %s - Same PCIe switch\n", r.colorConnection(ConnPIX)))
	sb.WriteString(fmt.Sprintf("    %s - Same PCIe bus\n", r.colorConnection(ConnPXB)))
	sb.WriteString(fmt.Sprintf("    %s - Same NUMA node\n", r.colorConnection(ConnPHB)))
	sb.WriteString(fmt.Sprintf("    %s - Cross NUMA\n", r.colorConnection(ConnSYS)))
	sb.WriteString(fmt.Sprintf("    %s   - No direct connection\n", ConnNone))

	return sb.String()
}

// RenderClusterSummary renders a summary of cluster topology.
func (r *Renderer) RenderClusterSummary(cluster *ClusterTopology) string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════╗\n")
	sb.WriteString("║          Cluster NVLink Topology Summary           ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("  Total Nodes:  %d\n", len(cluster.Nodes)))
	sb.WriteString(fmt.Sprintf("  Total GPUs:   %d\n", cluster.TotalGPUs))
	sb.WriteString(fmt.Sprintf("  Total Links:  %d\n", cluster.TotalLinks))
	sb.WriteString(fmt.Sprintf("  Active Links: %d\n", cluster.ActiveLinks))

	if cluster.ErrorLinks > 0 {
		if r.useColor {
			sb.WriteString(fmt.Sprintf("  Error Links:  %s\n", color.RedString("%d", cluster.ErrorLinks)))
		} else {
			sb.WriteString(fmt.Sprintf("  Error Links:  %d\n", cluster.ErrorLinks))
		}
	}

	sb.WriteString("\n  Per-Node Status:\n")
	for _, node := range cluster.Nodes {
		health := node.GetHealthStatus()
		statusStr := health.Status
		if r.useColor {
			switch health.Status {
			case "healthy":
				statusStr = color.GreenString("✓ " + health.Status)
			case "degraded":
				statusStr = color.YellowString("⚠ " + health.Status)
			case "critical":
				statusStr = color.RedString("✗ " + health.Status)
			}
		}
		sb.WriteString(fmt.Sprintf("    %-20s %d GPUs, %d/%d links - %s\n",
			node.NodeName, len(node.GPUs), health.ActiveLinks, health.TotalLinks, statusStr))
	}

	return sb.String()
}
