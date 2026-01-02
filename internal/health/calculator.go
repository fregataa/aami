package health

import (
	"fmt"
	"math"
	"time"
)

// Calculator handles health score calculations.
type Calculator struct {
	weights     ScoreWeights
	tempThresh  TemperatureThresholds
	eccThresh   ECCThresholds
	xidThresh   XidThresholds
}

// NewCalculator creates a new health calculator with default settings.
func NewCalculator() *Calculator {
	return &Calculator{
		weights:    DefaultWeights(),
		tempThresh: DefaultTemperatureThresholds(),
		eccThresh:  DefaultECCThresholds(),
		xidThresh:  DefaultXidThresholds(),
	}
}

// WithWeights sets custom weights.
func (c *Calculator) WithWeights(w ScoreWeights) *Calculator {
	c.weights = w
	return c
}

// WithTemperatureThresholds sets custom temperature thresholds.
func (c *Calculator) WithTemperatureThresholds(t TemperatureThresholds) *Calculator {
	c.tempThresh = t
	return c
}

// CalculateGPUHealth calculates the health score for a single GPU.
func (c *Calculator) CalculateGPUHealth(metrics GPUMetrics) GPUHealth {
	now := time.Now()

	health := GPUHealth{
		UUID:        metrics.UUID,
		Name:        metrics.Name,
		CollectedAt: now,
	}

	// Parse GPU index
	fmt.Sscanf(metrics.GPU, "%d", &health.Index)

	// Calculate component scores
	tempScore := c.calculateTemperatureScore(metrics.Temperature)
	eccScore := c.calculateECCScore(metrics.ECCSingleBit, metrics.ECCDoubleBit)
	xidScore := c.calculateXidScore(metrics.XidErrors)
	nvlinkScore := c.calculateNVLinkScore(metrics.NVLinkActive, metrics.NVLinkTotal)
	uptimeScore := c.calculateUptimeScore(metrics.Uptime)

	health.Components = []ComponentScore{
		tempScore,
		eccScore,
		xidScore,
		nvlinkScore,
		uptimeScore,
	}

	// Calculate overall score
	var totalWeight float64
	var weightedSum float64
	for _, comp := range health.Components {
		weightedSum += comp.Weighted
		totalWeight += comp.Weight
	}

	if totalWeight > 0 {
		health.OverallScore = weightedSum / totalWeight * 100
	}

	health.Status = GetStatusFromScore(health.OverallScore)

	return health
}

// calculateTemperatureScore calculates the temperature component score.
func (c *Calculator) calculateTemperatureScore(temp float64) ComponentScore {
	score := ComponentScore{
		Name:      "Temperature",
		Weight:    c.weights.Temperature,
		RawValue:  temp,
		Threshold: c.tempThresh.Critical,
	}

	switch {
	case temp <= c.tempThresh.Optimal:
		score.Score = 100
		score.Status = StatusHealthy
		score.Message = fmt.Sprintf("%.0f°C (optimal)", temp)
	case temp <= c.tempThresh.Warning:
		// Linear interpolation between optimal and warning
		score.Score = 100 - (temp-c.tempThresh.Optimal)/(c.tempThresh.Warning-c.tempThresh.Optimal)*30
		score.Status = StatusHealthy
		score.Message = fmt.Sprintf("%.0f°C (normal)", temp)
	case temp <= c.tempThresh.Critical:
		// Linear interpolation between warning and critical
		score.Score = 70 - (temp-c.tempThresh.Warning)/(c.tempThresh.Critical-c.tempThresh.Warning)*40
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%.0f°C (elevated)", temp)
	default:
		// Above critical
		score.Score = math.Max(0, 30-(temp-c.tempThresh.Critical)*2)
		score.Status = StatusCritical
		score.Message = fmt.Sprintf("%.0f°C (CRITICAL)", temp)
	}

	score.Weighted = score.Score * score.Weight / 100
	return score
}

// calculateECCScore calculates the ECC error component score.
func (c *Calculator) calculateECCScore(singleBit, doubleBit int64) ComponentScore {
	score := ComponentScore{
		Name:     "ECC Errors",
		Weight:   c.weights.ECCErrors,
		RawValue: float64(singleBit + doubleBit*100), // Weight DBE higher
	}

	// Double-bit errors are more severe
	if doubleBit >= int64(c.eccThresh.DoubleBitCritical) {
		score.Score = 0
		score.Status = StatusCritical
		score.Message = fmt.Sprintf("%d DBE (CRITICAL)", doubleBit)
	} else if doubleBit >= int64(c.eccThresh.DoubleBitWarning) {
		score.Score = 30
		score.Status = StatusCritical
		score.Message = fmt.Sprintf("%d DBE detected", doubleBit)
	} else if singleBit >= int64(c.eccThresh.SingleBitCritical) {
		score.Score = 40
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%d SBE (high)", singleBit)
	} else if singleBit >= int64(c.eccThresh.SingleBitWarning) {
		score.Score = 70
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%d SBE detected", singleBit)
	} else if singleBit > 0 {
		score.Score = 90
		score.Status = StatusHealthy
		score.Message = fmt.Sprintf("%d SBE (low)", singleBit)
	} else {
		score.Score = 100
		score.Status = StatusHealthy
		score.Message = "No errors"
	}

	score.Weighted = score.Score * score.Weight / 100
	return score
}

// calculateXidScore calculates the Xid error component score.
func (c *Calculator) calculateXidScore(xidCount int64) ComponentScore {
	score := ComponentScore{
		Name:     "Xid Errors",
		Weight:   c.weights.XidErrors,
		RawValue: float64(xidCount),
	}

	switch {
	case xidCount >= int64(c.xidThresh.CountCritical):
		score.Score = 0
		score.Status = StatusCritical
		score.Message = fmt.Sprintf("%d errors (CRITICAL)", xidCount)
	case xidCount >= int64(c.xidThresh.CountWarning):
		// Linear degradation
		score.Score = math.Max(0, 70-float64(xidCount-int64(c.xidThresh.CountWarning))*15)
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%d errors (24h)", xidCount)
	case xidCount > 0:
		score.Score = 80
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%d error (24h)", xidCount)
	default:
		score.Score = 100
		score.Status = StatusHealthy
		score.Message = "No errors"
	}

	score.Weighted = score.Score * score.Weight / 100
	return score
}

// calculateNVLinkScore calculates the NVLink status component score.
func (c *Calculator) calculateNVLinkScore(active, total int) ComponentScore {
	score := ComponentScore{
		Name:      "NVLink",
		Weight:    c.weights.NVLink,
		RawValue:  float64(active),
		Threshold: float64(total),
	}

	if total == 0 {
		// No NVLink on this GPU
		score.Score = 100
		score.Status = StatusHealthy
		score.Message = "N/A (no NVLink)"
		score.Weighted = score.Score * score.Weight / 100
		return score
	}

	ratio := float64(active) / float64(total)
	score.Score = ratio * 100

	switch {
	case ratio >= 1.0:
		score.Status = StatusHealthy
		score.Message = fmt.Sprintf("%d/%d active", active, total)
	case ratio >= 0.75:
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%d/%d active (degraded)", active, total)
	default:
		score.Status = StatusCritical
		score.Message = fmt.Sprintf("%d/%d active (CRITICAL)", active, total)
	}

	score.Weighted = score.Score * score.Weight / 100
	return score
}

// calculateUptimeScore calculates the uptime component score.
func (c *Calculator) calculateUptimeScore(uptimeSeconds float64) ComponentScore {
	score := ComponentScore{
		Name:     "Uptime",
		Weight:   c.weights.Uptime,
		RawValue: uptimeSeconds,
	}

	// Convert to hours
	uptimeHours := uptimeSeconds / 3600

	switch {
	case uptimeHours >= 168: // 7 days
		score.Score = 100
		score.Status = StatusHealthy
		score.Message = formatUptime(uptimeSeconds)
	case uptimeHours >= 24: // 1 day
		score.Score = 90
		score.Status = StatusHealthy
		score.Message = formatUptime(uptimeSeconds)
	case uptimeHours >= 1:
		score.Score = 70
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%s (recent restart)", formatUptime(uptimeSeconds))
	default:
		score.Score = 50
		score.Status = StatusWarning
		score.Message = fmt.Sprintf("%s (just started)", formatUptime(uptimeSeconds))
	}

	score.Weighted = score.Score * score.Weight / 100
	return score
}

// formatUptime formats uptime in human readable form.
func formatUptime(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// CalculateNodeHealth calculates health for all GPUs on a node.
func (c *Calculator) CalculateNodeHealth(metrics NodeMetrics) NodeHealth {
	now := time.Now()

	health := NodeHealth{
		NodeName:    metrics.NodeName,
		NodeIP:      metrics.NodeIP,
		CollectedAt: now,
	}

	var totalScore float64
	for _, gpuMetrics := range metrics.GPUs {
		gpuHealth := c.CalculateGPUHealth(gpuMetrics)
		health.GPUs = append(health.GPUs, gpuHealth)
		totalScore += gpuHealth.OverallScore

		switch gpuHealth.Status {
		case StatusHealthy:
			health.HealthyGPUs++
		case StatusWarning:
			health.WarningGPUs++
		case StatusCritical:
			health.CriticalGPUs++
		}
	}

	if len(health.GPUs) > 0 {
		health.OverallScore = totalScore / float64(len(health.GPUs))
	}

	health.Status = GetStatusFromScore(health.OverallScore)

	return health
}

// CalculateClusterHealth calculates health for the entire cluster.
func (c *Calculator) CalculateClusterHealth(nodes []NodeMetrics) ClusterHealth {
	now := time.Now()

	cluster := ClusterHealth{
		CollectedAt: now,
	}

	var totalScore float64
	for _, nodeMetrics := range nodes {
		nodeHealth := c.CalculateNodeHealth(nodeMetrics)
		cluster.Nodes = append(cluster.Nodes, nodeHealth)
		cluster.TotalGPUs += len(nodeHealth.GPUs)
		cluster.HealthyGPUs += nodeHealth.HealthyGPUs
		cluster.WarningGPUs += nodeHealth.WarningGPUs
		cluster.CriticalGPUs += nodeHealth.CriticalGPUs
		totalScore += nodeHealth.OverallScore
	}

	if len(cluster.Nodes) > 0 {
		cluster.OverallScore = totalScore / float64(len(cluster.Nodes))
	}

	cluster.Status = GetStatusFromScore(cluster.OverallScore)

	return cluster
}
