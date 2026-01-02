// Package nvlink provides NVLink topology collection and visualization.
package nvlink

// LinkStatus represents the status of an NVLink connection.
type LinkStatus string

const (
	LinkStatusActive   LinkStatus = "active"
	LinkStatusInactive LinkStatus = "inactive"
	LinkStatusDisabled LinkStatus = "disabled"
	LinkStatusError    LinkStatus = "error"
	LinkStatusUnknown  LinkStatus = "unknown"
)

// GPUInfo holds basic GPU information.
type GPUInfo struct {
	Index       int    `json:"index"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	BusID       string `json:"bus_id"`
	NVLinkCount int    `json:"nvlink_count"`
}

// NVLinkInfo represents a single NVLink connection between GPUs.
type NVLinkInfo struct {
	SourceGPU   int        `json:"source_gpu"`
	TargetGPU   int        `json:"target_gpu"`
	LinkIndex   int        `json:"link_index"`
	Status      LinkStatus `json:"status"`
	Version     int        `json:"version"`      // NVLink version (1, 2, 3, 4)
	Bandwidth   float64    `json:"bandwidth"`    // GB/s per direction
	TXThroughput float64   `json:"tx_throughput"` // Current TX throughput
	RXThroughput float64   `json:"rx_throughput"` // Current RX throughput
}

// P2PCapability represents peer-to-peer capability between two GPUs.
type P2PCapability struct {
	GPU1       int    `json:"gpu1"`
	GPU2       int    `json:"gpu2"`
	P2PRead    bool   `json:"p2p_read"`
	P2PWrite   bool   `json:"p2p_write"`
	P2PAtomic  bool   `json:"p2p_atomic"`
	Connection string `json:"connection"` // NVLink, PIX, PXB, PHB, SYS, etc.
}

// NodeTopology represents the complete NVLink topology of a node.
type NodeTopology struct {
	NodeName     string          `json:"node_name"`
	GPUs         []GPUInfo       `json:"gpus"`
	Links        []NVLinkInfo    `json:"links"`
	P2PMatrix    []P2PCapability `json:"p2p_matrix"`
	TotalLinks   int             `json:"total_links"`
	ActiveLinks  int             `json:"active_links"`
	ErrorLinks   int             `json:"error_links"`
	CollectedAt  string          `json:"collected_at"`
}

// TopologyMatrix represents the GPU interconnect matrix.
type TopologyMatrix struct {
	Size        int        `json:"size"` // Number of GPUs
	Connections [][]string `json:"connections"` // Connection type matrix
}

// ConnectionType constants for topology matrix.
const (
	ConnNVLink = "NV"   // NVLink connection
	ConnPIX    = "PIX"  // Same PCIe switch
	ConnPXB    = "PXB"  // Same PCIe bus
	ConnPHB    = "PHB"  // Same NUMA node (PCIe Host Bridge)
	ConnSYS    = "SYS"  // Cross NUMA (system interconnect)
	ConnSelf   = "X"    // Same GPU
	ConnNone   = "-"    // No connection
)

// NVLinkVersion bandwidth specifications (per link, per direction).
var NVLinkBandwidth = map[int]float64{
	1: 20.0,  // NVLink 1.0: 20 GB/s
	2: 25.0,  // NVLink 2.0: 25 GB/s
	3: 25.0,  // NVLink 3.0: 25 GB/s (same speed, more links)
	4: 25.0,  // NVLink 4.0: 25 GB/s per sub-link
}

// ClusterTopology represents topology across multiple nodes.
type ClusterTopology struct {
	Nodes       []NodeTopology `json:"nodes"`
	TotalGPUs   int            `json:"total_gpus"`
	TotalLinks  int            `json:"total_links"`
	ActiveLinks int            `json:"active_links"`
	ErrorLinks  int            `json:"error_links"`
}

// TopologyHealth provides a health summary of the topology.
type TopologyHealth struct {
	NodeName      string  `json:"node_name"`
	TotalLinks    int     `json:"total_links"`
	ActiveLinks   int     `json:"active_links"`
	InactiveLinks int     `json:"inactive_links"`
	ErrorLinks    int     `json:"error_links"`
	HealthPercent float64 `json:"health_percent"`
	Status        string  `json:"status"` // healthy, degraded, critical
}

// GetHealthStatus returns health status based on active link percentage.
func (t *NodeTopology) GetHealthStatus() TopologyHealth {
	health := TopologyHealth{
		NodeName:    t.NodeName,
		TotalLinks:  t.TotalLinks,
		ActiveLinks: t.ActiveLinks,
		ErrorLinks:  t.ErrorLinks,
	}

	if t.TotalLinks > 0 {
		health.HealthPercent = float64(t.ActiveLinks) / float64(t.TotalLinks) * 100
	}

	health.InactiveLinks = t.TotalLinks - t.ActiveLinks - t.ErrorLinks

	switch {
	case health.HealthPercent >= 100:
		health.Status = "healthy"
	case health.HealthPercent >= 75:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health
}
