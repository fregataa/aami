package health

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PrometheusClient handles queries to Prometheus.
type PrometheusClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPrometheusClient creates a new Prometheus client.
func NewPrometheusClient(baseURL string) *PrometheusClient {
	return &PrometheusClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PrometheusResponse represents the response from Prometheus API.
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"` // [timestamp, value]
		} `json:"result"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

// Query executes an instant query against Prometheus.
func (c *PrometheusClient) Query(query string) (*PrometheusResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query", c.baseURL)

	params := url.Values{}
	params.Set("query", query)

	resp, err := c.httpClient.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result PrometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus error: %s - %s", result.ErrorType, result.Error)
	}

	return &result, nil
}

// GetGPUTemperatures queries GPU temperatures.
func (c *PrometheusClient) GetGPUTemperatures() (map[string]float64, error) {
	query := "DCGM_FI_DEV_GPU_TEMP"
	result, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	return c.extractMetrics(result, "gpu")
}

// GetECCSingleBitErrors queries single-bit ECC errors.
func (c *PrometheusClient) GetECCSingleBitErrors() (map[string]float64, error) {
	query := "DCGM_FI_DEV_ECC_SBE_VOL_TOTAL"
	result, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	return c.extractMetrics(result, "gpu")
}

// GetECCDoubleBitErrors queries double-bit ECC errors.
func (c *PrometheusClient) GetECCDoubleBitErrors() (map[string]float64, error) {
	query := "DCGM_FI_DEV_ECC_DBE_VOL_TOTAL"
	result, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	return c.extractMetrics(result, "gpu")
}

// GetXidErrors queries Xid errors in the last 24 hours.
func (c *PrometheusClient) GetXidErrors() (map[string]float64, error) {
	query := "increase(DCGM_FI_DEV_XID_ERRORS[24h])"
	result, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	return c.extractMetrics(result, "gpu")
}

// GetNVLinkStatus queries NVLink status.
func (c *PrometheusClient) GetNVLinkStatus() (map[string]NVLinkMetrics, error) {
	// Query active NVLink count
	query := "DCGM_FI_DEV_NVLINK_LINK_COUNT"
	result, err := c.Query(query)
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]NVLinkMetrics)
	for _, r := range result.Data.Result {
		gpu := r.Metric["gpu"]
		if gpu == "" {
			continue
		}
		val, _ := c.parseValue(r.Value)
		metrics[gpu] = NVLinkMetrics{
			Active: int(val),
			Total:  int(val), // Assume all links are total
		}
	}
	return metrics, nil
}

// NVLinkMetrics represents NVLink metrics for a GPU.
type NVLinkMetrics struct {
	Active int
	Total  int
}

// GetGPUUptime queries GPU uptime (based on driver uptime or process start).
func (c *PrometheusClient) GetGPUUptime() (map[string]float64, error) {
	// Use process_start_time_seconds from dcgm-exporter if available
	// Fall back to node uptime
	query := "time() - process_start_time_seconds{job=~\"dcgm.*\"}"
	result, err := c.Query(query)
	if err != nil {
		// Fallback: use a default uptime
		return make(map[string]float64), nil
	}
	return c.extractMetrics(result, "instance")
}

// CollectAllMetrics collects all metrics for health calculation.
func (c *PrometheusClient) CollectAllMetrics() ([]NodeMetrics, error) {
	// Get all GPU info first
	gpuInfo, err := c.getGPUInfo()
	if err != nil {
		return nil, fmt.Errorf("get GPU info: %w", err)
	}

	// Collect all metrics in parallel
	temps, _ := c.GetGPUTemperatures()
	eccSingle, _ := c.GetECCSingleBitErrors()
	eccDouble, _ := c.GetECCDoubleBitErrors()
	xidErrors, _ := c.GetXidErrors()
	nvlinkStatus, _ := c.GetNVLinkStatus()

	// Organize by node
	nodeMap := make(map[string]*NodeMetrics)

	for key, info := range gpuInfo {
		instance := info.Instance
		if nodeMap[instance] == nil {
			nodeMap[instance] = &NodeMetrics{
				NodeName: instance,
				NodeIP:   instance,
			}
		}

		gpuMetric := GPUMetrics{
			GPU:      info.GPU,
			UUID:     info.UUID,
			Name:     info.Name,
			Instance: instance,
		}

		if v, ok := temps[key]; ok {
			gpuMetric.Temperature = v
		}
		if v, ok := eccSingle[key]; ok {
			gpuMetric.ECCSingleBit = int64(v)
		}
		if v, ok := eccDouble[key]; ok {
			gpuMetric.ECCDoubleBit = int64(v)
		}
		if v, ok := xidErrors[key]; ok {
			gpuMetric.XidErrors = int64(v)
		}
		if nv, ok := nvlinkStatus[info.GPU]; ok {
			gpuMetric.NVLinkActive = nv.Active
			gpuMetric.NVLinkTotal = nv.Total
		}

		nodeMap[instance].GPUs = append(nodeMap[instance].GPUs, gpuMetric)
	}

	// Convert map to slice
	var nodes []NodeMetrics
	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}

	return nodes, nil
}

// GPUInfo represents basic GPU information.
type GPUInfo struct {
	GPU      string
	UUID     string
	Name     string
	Instance string
}

// getGPUInfo retrieves GPU information from Prometheus.
func (c *PrometheusClient) getGPUInfo() (map[string]GPUInfo, error) {
	// Use a metric that exists on all GPUs to get labels
	query := "DCGM_FI_DEV_GPU_TEMP"
	result, err := c.Query(query)
	if err != nil {
		return nil, err
	}

	info := make(map[string]GPUInfo)
	for _, r := range result.Data.Result {
		gpu := r.Metric["gpu"]
		if gpu == "" {
			continue
		}

		key := fmt.Sprintf("%s_%s", r.Metric["instance"], gpu)
		info[key] = GPUInfo{
			GPU:      gpu,
			UUID:     r.Metric["UUID"],
			Name:     r.Metric["modelName"],
			Instance: r.Metric["instance"],
		}
	}

	return info, nil
}

// extractMetrics extracts metric values keyed by the specified label.
func (c *PrometheusClient) extractMetrics(result *PrometheusResponse, keyLabel string) (map[string]float64, error) {
	metrics := make(map[string]float64)

	for _, r := range result.Data.Result {
		key := r.Metric[keyLabel]
		if key == "" {
			continue
		}

		// For GPU metrics, include instance in key for uniqueness
		if keyLabel == "gpu" {
			instance := r.Metric["instance"]
			if instance != "" {
				key = fmt.Sprintf("%s_%s", instance, key)
			}
		}

		val, err := c.parseValue(r.Value)
		if err != nil {
			continue
		}
		metrics[key] = val
	}

	return metrics, nil
}

// parseValue parses the value from Prometheus response.
func (c *PrometheusClient) parseValue(value []interface{}) (float64, error) {
	if len(value) < 2 {
		return 0, fmt.Errorf("invalid value format")
	}

	strVal, ok := value[1].(string)
	if !ok {
		return 0, fmt.Errorf("value is not string")
	}

	return strconv.ParseFloat(strVal, 64)
}

// CheckConnection verifies connection to Prometheus.
func (c *PrometheusClient) CheckConnection() error {
	_, err := c.Query("up")
	return err
}
