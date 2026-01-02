package multicluster

import (
	"context"
	"sort"
	"sync"
	"time"
)

// Aggregator collects and aggregates metrics from multiple clusters.
type Aggregator struct {
	registry *Registry
	clients  map[string]*Client
	mu       sync.RWMutex
}

// NewAggregator creates a new metric aggregator.
func NewAggregator(registry *Registry) *Aggregator {
	return &Aggregator{
		registry: registry,
		clients:  make(map[string]*Client),
	}
}

// Initialize creates clients for all registered clusters.
func (a *Aggregator) Initialize() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Close existing clients
	for _, client := range a.clients {
		client.Close()
	}
	a.clients = make(map[string]*Client)

	// Create new clients
	for _, cfg := range a.registry.List() {
		client, err := NewClient(cfg)
		if err != nil {
			continue // Skip clusters that fail to initialize
		}
		a.clients[cfg.Name] = client
	}

	return nil
}

// Refresh re-initializes clients for any new clusters.
func (a *Aggregator) Refresh() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	clusters := a.registry.List()
	clusterMap := make(map[string]bool)

	// Add new clusters
	for _, cfg := range clusters {
		clusterMap[cfg.Name] = true
		if _, exists := a.clients[cfg.Name]; !exists {
			client, err := NewClient(cfg)
			if err != nil {
				continue
			}
			a.clients[cfg.Name] = client
		}
	}

	// Remove deleted clusters
	for name, client := range a.clients {
		if !clusterMap[name] {
			client.Close()
			delete(a.clients, name)
		}
	}

	return nil
}

// GetAggregatedStatus collects status from all clusters.
func (a *Aggregator) GetAggregatedStatus(ctx context.Context) ([]ClusterStatus, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var wg sync.WaitGroup
	results := make(chan ClusterStatus, len(a.clients))

	for name, client := range a.clients {
		wg.Add(1)
		go func(name string, c *Client) {
			defer wg.Done()

			status, err := c.GetStatus(ctx)
			if err != nil {
				results <- ClusterStatus{
					Name:      name,
					Endpoint:  c.config.Endpoint,
					Connected: false,
					Error:     err.Error(),
				}
				return
			}
			results <- *status
		}(name, client)
	}

	// Wait and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	var statuses []ClusterStatus
	for status := range results {
		statuses = append(statuses, status)
	}

	// Sort by name for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	return statuses, nil
}

// GetAggregatedMetrics computes aggregated metrics across all clusters.
func (a *Aggregator) GetAggregatedMetrics(ctx context.Context) (*AggregatedMetrics, error) {
	statuses, err := a.GetAggregatedStatus(ctx)
	if err != nil {
		return nil, err
	}

	metrics := &AggregatedMetrics{
		Timestamp:        time.Now(),
		ClusterBreakdown: make(map[string]ClusterMetrics),
	}

	var totalHealth float64
	var healthyCount int

	for _, status := range statuses {
		metrics.ClusterCount++
		metrics.TotalNodes += status.Nodes
		metrics.HealthyNodes += status.HealthyNodes
		metrics.TotalGPUs += status.TotalGPUs
		metrics.HealthyGPUs += status.HealthyGPUs
		metrics.ActiveAlerts += status.AlertsActive

		if status.Connected {
			metrics.ConnectedCount++
			totalHealth += status.HealthScore
			healthyCount++
		}

		metrics.ClusterBreakdown[status.Name] = ClusterMetrics{
			Nodes:        status.Nodes,
			HealthyNodes: status.HealthyNodes,
			GPUs:         status.TotalGPUs,
			HealthyGPUs:  status.HealthyGPUs,
			HealthScore:  status.HealthScore,
			AlertCount:   status.AlertsActive,
			Connected:    status.Connected,
		}
	}

	if healthyCount > 0 {
		metrics.AverageHealth = totalHealth / float64(healthyCount)
	}

	return metrics, nil
}

// GetAllAlerts collects alerts from all clusters.
func (a *Aggregator) GetAllAlerts(ctx context.Context) ([]GlobalAlert, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var wg sync.WaitGroup
	results := make(chan []GlobalAlert, len(a.clients))

	for _, client := range a.clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()

			alerts, err := c.GetAlerts(ctx)
			if err != nil {
				return
			}
			results <- alerts
		}(client)
	}

	// Wait and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	var allAlerts []GlobalAlert
	for alerts := range results {
		allAlerts = append(allAlerts, alerts...)
	}

	// Sort by severity and time
	sort.Slice(allAlerts, func(i, j int) bool {
		// Critical first
		if allAlerts[i].Severity != allAlerts[j].Severity {
			return severityOrder(allAlerts[i].Severity) < severityOrder(allAlerts[j].Severity)
		}
		// Then by time (newest first)
		return allAlerts[i].FiredAt.After(allAlerts[j].FiredAt)
	})

	return allAlerts, nil
}

// GetCriticalAlerts returns only critical alerts from all clusters.
func (a *Aggregator) GetCriticalAlerts(ctx context.Context) ([]GlobalAlert, error) {
	alerts, err := a.GetAllAlerts(ctx)
	if err != nil {
		return nil, err
	}

	var critical []GlobalAlert
	for _, alert := range alerts {
		if alert.Severity == "critical" {
			critical = append(critical, alert)
		}
	}

	return critical, nil
}

// GetAllEvents collects events from all clusters.
func (a *Aggregator) GetAllEvents(ctx context.Context, limit int) ([]ClusterEvent, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var wg sync.WaitGroup
	results := make(chan []ClusterEvent, len(a.clients))

	for _, client := range a.clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()

			events, err := c.GetEvents(ctx, limit)
			if err != nil {
				return
			}
			results <- events
		}(client)
	}

	// Wait and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	var allEvents []ClusterEvent
	for events := range results {
		allEvents = append(allEvents, events...)
	}

	// Sort by timestamp (newest first)
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp.After(allEvents[j].Timestamp)
	})

	// Limit total events
	if len(allEvents) > limit {
		allEvents = allEvents[:limit]
	}

	return allEvents, nil
}

// GetClusterSummaries returns brief summaries for all clusters.
func (a *Aggregator) GetClusterSummaries(ctx context.Context) ([]ClusterSummary, error) {
	statuses, err := a.GetAggregatedStatus(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]ClusterSummary, len(statuses))
	for i, status := range statuses {
		summaries[i] = status.ToSummary()
	}

	return summaries, nil
}

// GetClusterClient returns a client for a specific cluster.
func (a *Aggregator) GetClusterClient(name string) (*Client, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	client, ok := a.clients[name]
	return client, ok
}

// Close closes all clients.
func (a *Aggregator) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, client := range a.clients {
		client.Close()
	}
	a.clients = make(map[string]*Client)

	return nil
}

// TestAllConnections tests connections to all clusters.
func (a *Aggregator) TestAllConnections(ctx context.Context) map[string]error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var wg sync.WaitGroup
	type result struct {
		name string
		err  error
	}
	results := make(chan result, len(a.clients))

	for name, client := range a.clients {
		wg.Add(1)
		go func(name string, c *Client) {
			defer wg.Done()
			err := c.TestConnection(ctx)
			results <- result{name: name, err: err}
		}(name, client)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	errors := make(map[string]error)
	for r := range results {
		if r.err != nil {
			errors[r.name] = r.err
		}
	}

	return errors
}

// severityOrder returns an order value for severity (lower is more severe).
func severityOrder(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 3
	}
}

// WatchAlerts starts a goroutine that watches for new alerts.
func (a *Aggregator) WatchAlerts(ctx context.Context, interval time.Duration, callback func([]GlobalAlert)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	seenAlerts := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			alerts, err := a.GetAllAlerts(ctx)
			if err != nil {
				continue
			}

			var newAlerts []GlobalAlert
			for _, alert := range alerts {
				key := alert.Cluster + ":" + alert.AlertName + ":" + alert.Node
				if !seenAlerts[key] {
					seenAlerts[key] = true
					newAlerts = append(newAlerts, alert)
				}
			}

			if len(newAlerts) > 0 {
				callback(newAlerts)
			}
		}
	}
}

// GetUnhealthyClusters returns clusters with health score below threshold.
func (a *Aggregator) GetUnhealthyClusters(ctx context.Context, threshold float64) ([]ClusterStatus, error) {
	statuses, err := a.GetAggregatedStatus(ctx)
	if err != nil {
		return nil, err
	}

	var unhealthy []ClusterStatus
	for _, status := range statuses {
		if status.Connected && status.HealthScore < threshold {
			unhealthy = append(unhealthy, status)
		}
	}

	return unhealthy, nil
}

// GetDisconnectedClusters returns clusters that are not connected.
func (a *Aggregator) GetDisconnectedClusters(ctx context.Context) ([]ClusterStatus, error) {
	statuses, err := a.GetAggregatedStatus(ctx)
	if err != nil {
		return nil, err
	}

	var disconnected []ClusterStatus
	for _, status := range statuses {
		if !status.Connected {
			disconnected = append(disconnected, status)
		}
	}

	return disconnected, nil
}
