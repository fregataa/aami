package domain

import "fmt"

// PrometheusSDTarget represents a single target in Prometheus Service Discovery format
type PrometheusSDTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

// NewPrometheusSDTarget creates a new Prometheus SD target from Target and Exporter
func NewPrometheusSDTarget(target *Target, exporter *Exporter) *PrometheusSDTarget {
	// Create target address (IP:Port)
	targetAddr := fmt.Sprintf("%s:%d", target.IPAddress, exporter.Port)

	// Build labels
	labels := make(map[string]string)

	// Add target labels
	labels["job"] = string(exporter.Type)
	labels["instance"] = target.Hostname
	labels["hostname"] = target.Hostname
	labels["ip_address"] = target.IPAddress
	labels["target_id"] = target.ID
	labels["exporter_type"] = string(exporter.Type)
	labels["exporter_id"] = exporter.ID
	labels["metrics_path"] = exporter.MetricsPath

	// Add target status
	labels["target_status"] = string(target.Status)

	// Add user-defined labels from target
	for k, v := range target.Labels {
		// Prefix user labels to avoid conflicts
		labels["target_label_"+k] = v
	}

	// Add group information
	if len(target.Groups) > 0 {
		// Add all group names
		for i, group := range target.Groups {
			labels[fmt.Sprintf("group_%d", i)] = group.Name
			labels[fmt.Sprintf("group_%d_id", i)] = group.ID
			if group.Namespace != nil {
				labels[fmt.Sprintf("group_%d_namespace", i)] = group.Namespace.Name
			}
		}
		// Add primary group (first group) as default
		labels["group"] = target.Groups[0].Name
		labels["group_id"] = target.Groups[0].ID
		if target.Groups[0].Namespace != nil {
			labels["namespace"] = target.Groups[0].Namespace.Name
			labels["namespace_id"] = target.Groups[0].Namespace.ID
		}
	}

	return &PrometheusSDTarget{
		Targets: []string{targetAddr},
		Labels:  labels,
	}
}

// ServiceDiscoveryFilter represents filters for service discovery query
type ServiceDiscoveryFilter struct {
	Status        *TargetStatus         `json:"status,omitempty"`
	ExporterType  *ExporterType         `json:"exporter_type,omitempty"`
	GroupID       *string               `json:"group_id,omitempty"`
	NamespaceID   *string               `json:"namespace_id,omitempty"`
	Labels        map[string]string     `json:"labels,omitempty"`
	EnabledOnly   bool                  `json:"enabled_only"`
}
