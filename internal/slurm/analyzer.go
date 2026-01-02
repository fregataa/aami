package slurm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Analyzer correlates Slurm jobs with GPU events.
type Analyzer struct {
	slurmClient    *Client
	prometheusURL  string
	httpClient     *http.Client
}

// NewAnalyzer creates a new job-GPU analyzer.
func NewAnalyzer(slurm *Client, prometheusURL string) *Analyzer {
	return &Analyzer{
		slurmClient:   slurm,
		prometheusURL: prometheusURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AnalyzeJob correlates a job with GPU events during its execution.
func (a *Analyzer) AnalyzeJob(ctx context.Context, jobID int64) (*JobGPUCorrelation, error) {
	job, err := a.slurmClient.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}

	result := &JobGPUCorrelation{
		Job:         *job,
		Correlation: CorrelationNone,
	}

	// Query GPU events during job execution
	events, err := a.queryGPUEvents(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("query GPU events: %w", err)
	}
	result.GPUEvents = events

	// Analyze correlation
	a.analyzeCorrelation(result)

	return result, nil
}

// queryGPUEvents queries Prometheus for GPU events during job execution.
func (a *Analyzer) queryGPUEvents(ctx context.Context, job *Job) ([]GPUEvent, error) {
	var allEvents []GPUEvent

	// Determine time range
	startTime := job.StartTime
	endTime := job.EndTime
	if endTime.IsZero() {
		endTime = time.Now()
	}

	// Add buffer for events that may have occurred just before/after
	startTime = startTime.Add(-1 * time.Minute)
	endTime = endTime.Add(1 * time.Minute)

	for _, node := range job.Nodes {
		// Query Xid errors
		xidEvents, err := a.queryXidErrors(ctx, node, startTime, endTime)
		if err == nil {
			allEvents = append(allEvents, xidEvents...)
		}

		// Query high temperature events
		tempEvents, err := a.queryTemperatureEvents(ctx, node, startTime, endTime)
		if err == nil {
			allEvents = append(allEvents, tempEvents...)
		}

		// Query ECC errors
		eccEvents, err := a.queryECCErrors(ctx, node, startTime, endTime)
		if err == nil {
			allEvents = append(allEvents, eccEvents...)
		}

		// Query power throttling
		throttleEvents, err := a.queryThrottleEvents(ctx, node, startTime, endTime)
		if err == nil {
			allEvents = append(allEvents, throttleEvents...)
		}
	}

	// Sort by timestamp
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp.Before(allEvents[j].Timestamp)
	})

	return allEvents, nil
}

// queryXidErrors queries for Xid errors on a node.
func (a *Analyzer) queryXidErrors(ctx context.Context, node string, start, end time.Time) ([]GPUEvent, error) {
	query := fmt.Sprintf(`DCGM_FI_DEV_XID_ERRORS{instance=~"%s.*"} > 0`, node)
	return a.queryRangeEvents(ctx, query, start, end, "xid", "critical")
}

// queryTemperatureEvents queries for high temperature events.
func (a *Analyzer) queryTemperatureEvents(ctx context.Context, node string, start, end time.Time) ([]GPUEvent, error) {
	query := fmt.Sprintf(`DCGM_FI_DEV_GPU_TEMP{instance=~"%s.*"} > 83`, node)
	return a.queryRangeEvents(ctx, query, start, end, "temperature", "warning")
}

// queryECCErrors queries for ECC errors.
func (a *Analyzer) queryECCErrors(ctx context.Context, node string, start, end time.Time) ([]GPUEvent, error) {
	// Double-bit ECC errors (uncorrectable)
	query := fmt.Sprintf(`increase(DCGM_FI_DEV_ECC_DBE_VOL_TOTAL{instance=~"%s.*"}[5m]) > 0`, node)
	return a.queryRangeEvents(ctx, query, start, end, "ecc_dbe", "critical")
}

// queryThrottleEvents queries for power/thermal throttling.
func (a *Analyzer) queryThrottleEvents(ctx context.Context, node string, start, end time.Time) ([]GPUEvent, error) {
	query := fmt.Sprintf(`DCGM_FI_DEV_POWER_VIOLATION{instance=~"%s.*"} > 0`, node)
	return a.queryRangeEvents(ctx, query, start, end, "throttle", "warning")
}

// queryRangeEvents executes a range query and returns events.
func (a *Analyzer) queryRangeEvents(ctx context.Context, query string, start, end time.Time, eventType, severity string) ([]GPUEvent, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query_range", a.prometheusURL)

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", "60") // 1 minute resolution

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"` // [[timestamp, value], ...]
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var events []GPUEvent
	for _, series := range result.Data.Result {
		node := extractNode(series.Metric)
		gpu := extractGPU(series.Metric)

		for _, v := range series.Values {
			if len(v) < 2 {
				continue
			}

			ts, ok := v[0].(float64)
			if !ok {
				continue
			}

			value := ""
			switch val := v[1].(type) {
			case string:
				value = val
			case float64:
				value = strconv.FormatFloat(val, 'f', 2, 64)
			}

			events = append(events, GPUEvent{
				Timestamp: time.Unix(int64(ts), 0),
				Node:      node,
				GPUIndex:  gpu,
				Type:      eventType,
				Value:     value,
				Severity:  severity,
				Message:   formatEventMessage(eventType, value, gpu),
			})
		}
	}

	return events, nil
}

// analyzeCorrelation determines the correlation between job failure and GPU events.
func (a *Analyzer) analyzeCorrelation(result *JobGPUCorrelation) {
	if len(result.GPUEvents) == 0 {
		result.Correlation = CorrelationNone
		result.Confidence = 0
		result.Recommendation = "No GPU issues detected during job execution"
		result.Summary = "Job completed without detected GPU anomalies"
		return
	}

	// Categorize events
	var xidEvents, eccEvents, tempEvents, throttleEvents []GPUEvent
	for _, event := range result.GPUEvents {
		switch event.Type {
		case "xid":
			xidEvents = append(xidEvents, event)
		case "ecc_dbe":
			eccEvents = append(eccEvents, event)
		case "temperature":
			tempEvents = append(tempEvents, event)
		case "throttle":
			throttleEvents = append(throttleEvents, event)
		}
	}

	// Analyze based on job state
	jobFailed := result.Job.State == JobStateFailed ||
		result.Job.State == JobStateNodeFail ||
		result.Job.State == JobStateTimeout ||
		result.Job.ExitCode != 0

	// Determine correlation level
	if len(xidEvents) > 0 || len(eccEvents) > 0 {
		if jobFailed {
			result.Correlation = CorrelationConfirmed
			result.Confidence = 0.95
		} else {
			result.Correlation = CorrelationLikely
			result.Confidence = 0.7
		}
	} else if len(tempEvents) > 0 {
		if jobFailed {
			result.Correlation = CorrelationLikely
			result.Confidence = 0.6
		} else {
			result.Correlation = CorrelationPossible
			result.Confidence = 0.3
		}
	} else if len(throttleEvents) > 0 {
		result.Correlation = CorrelationPossible
		result.Confidence = 0.4
	} else {
		result.Correlation = CorrelationUnlikely
		result.Confidence = 0.2
	}

	// Generate recommendation
	result.Recommendation = a.generateRecommendation(result, xidEvents, eccEvents, tempEvents, throttleEvents)
	result.Summary = a.generateSummary(result, xidEvents, eccEvents, tempEvents, throttleEvents)

	// Identify affected GPUs
	affectedGPUs := make(map[string]GPUAllocation)
	for _, event := range result.GPUEvents {
		if event.Severity == "critical" {
			key := fmt.Sprintf("%s:%d", event.Node, event.GPUIndex)
			affectedGPUs[key] = GPUAllocation{
				Node:     event.Node,
				GPUIndex: event.GPUIndex,
			}
		}
	}
	for _, gpu := range affectedGPUs {
		result.AffectedGPUs = append(result.AffectedGPUs, gpu)
	}
}

// generateRecommendation creates actionable recommendations.
func (a *Analyzer) generateRecommendation(result *JobGPUCorrelation, xid, ecc, temp, throttle []GPUEvent) string {
	var recommendations []string

	if len(xid) > 0 {
		// Get unique nodes with Xid errors
		nodes := make(map[string]bool)
		for _, e := range xid {
			nodes[e.Node] = true
		}
		nodeList := make([]string, 0, len(nodes))
		for n := range nodes {
			nodeList = append(nodeList, n)
		}
		recommendations = append(recommendations,
			fmt.Sprintf("CRITICAL: Xid errors detected on %s. Drain node(s) and inspect GPU hardware.",
				strings.Join(nodeList, ", ")))
	}

	if len(ecc) > 0 {
		nodes := make(map[string]bool)
		for _, e := range ecc {
			nodes[e.Node] = true
		}
		nodeList := make([]string, 0, len(nodes))
		for n := range nodes {
			nodeList = append(nodeList, n)
		}
		recommendations = append(recommendations,
			fmt.Sprintf("CRITICAL: Uncorrectable ECC errors on %s. GPU memory may be failing.",
				strings.Join(nodeList, ", ")))
	}

	if len(temp) > 0 {
		recommendations = append(recommendations,
			"WARNING: High GPU temperatures detected. Check cooling systems and airflow.")
	}

	if len(throttle) > 0 {
		recommendations = append(recommendations,
			"INFO: Power throttling detected. Consider reducing job workload or checking power limits.")
	}

	if len(recommendations) == 0 {
		return "No specific action required"
	}

	return strings.Join(recommendations, "\n")
}

// generateSummary creates a human-readable summary.
func (a *Analyzer) generateSummary(result *JobGPUCorrelation, xid, ecc, temp, throttle []GPUEvent) string {
	parts := []string{}

	if len(xid) > 0 {
		parts = append(parts, fmt.Sprintf("%d Xid error(s)", len(xid)))
	}
	if len(ecc) > 0 {
		parts = append(parts, fmt.Sprintf("%d ECC error(s)", len(ecc)))
	}
	if len(temp) > 0 {
		parts = append(parts, fmt.Sprintf("%d high temperature event(s)", len(temp)))
	}
	if len(throttle) > 0 {
		parts = append(parts, fmt.Sprintf("%d throttling event(s)", len(throttle)))
	}

	if len(parts) == 0 {
		return "No GPU issues detected"
	}

	return fmt.Sprintf("Detected: %s", strings.Join(parts, ", "))
}

// FindAffectedJobs finds jobs that may have been affected by GPU issues on a node.
func (a *Analyzer) FindAffectedJobs(ctx context.Context, node string, since time.Time) ([]Job, error) {
	// Get jobs that ran on this node
	jobs, err := a.slurmClient.GetJobs(ctx, JobFilter{Node: node})
	if err != nil {
		return nil, err
	}

	var affected []Job
	for _, job := range jobs {
		if job.StartTime.After(since) || job.StartTime.IsZero() {
			affected = append(affected, job)
		}
	}

	return affected, nil
}

// AnalyzeNode analyzes all recent jobs on a node for GPU correlation.
func (a *Analyzer) AnalyzeNode(ctx context.Context, node string, hours int) ([]JobGPUCorrelation, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	jobs, err := a.FindAffectedJobs(ctx, node, since)
	if err != nil {
		return nil, err
	}

	var correlations []JobGPUCorrelation
	for _, job := range jobs {
		correlation, err := a.AnalyzeJob(ctx, job.ID)
		if err != nil {
			continue
		}
		correlations = append(correlations, *correlation)
	}

	return correlations, nil
}

// GetJobStatistics calculates job statistics for a time period.
func (a *Analyzer) GetJobStatistics(ctx context.Context, partition string, hours int) (*JobStatistics, error) {
	// This would typically query sacct or Prometheus
	// Placeholder implementation
	stats := &JobStatistics{
		Period: fmt.Sprintf("Last %d hours", hours),
	}

	// Query completed and failed jobs
	// In production, this would use sacct or stored metrics

	return stats, nil
}

// Helper functions

func extractNode(metric map[string]string) string {
	if instance, ok := metric["instance"]; ok {
		// Remove port
		return strings.Split(instance, ":")[0]
	}
	if node, ok := metric["node"]; ok {
		return node
	}
	if hostname, ok := metric["hostname"]; ok {
		return hostname
	}
	return "unknown"
}

func extractGPU(metric map[string]string) int {
	if gpu, ok := metric["gpu"]; ok {
		idx, _ := strconv.Atoi(gpu)
		return idx
	}
	if device, ok := metric["device"]; ok {
		// Format: nvidia0, nvidia1, etc.
		idx, _ := strconv.Atoi(strings.TrimPrefix(device, "nvidia"))
		return idx
	}
	return 0
}

func formatEventMessage(eventType, value string, gpu int) string {
	switch eventType {
	case "xid":
		return fmt.Sprintf("Xid error %s on GPU %d", value, gpu)
	case "ecc_dbe":
		return fmt.Sprintf("Uncorrectable ECC error on GPU %d (count: %s)", gpu, value)
	case "temperature":
		return fmt.Sprintf("High temperature on GPU %d: %s°C", gpu, value)
	case "throttle":
		return fmt.Sprintf("Power throttling on GPU %d", gpu)
	default:
		return fmt.Sprintf("%s event on GPU %d: %s", eventType, gpu, value)
	}
}
