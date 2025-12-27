package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestExporterType_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		exporterType domain.ExporterType
		want         bool
	}{
		{
			name:         "node exporter is valid",
			exporterType: domain.ExporterTypeNodeExporter,
			want:         true,
		},
		{
			name:         "dcgm exporter is valid",
			exporterType: domain.ExporterTypeDCGMExporter,
			want:         true,
		},
		{
			name:         "custom exporter is valid",
			exporterType: domain.ExporterTypeCustom,
			want:         true,
		},
		{
			name:         "invalid exporter type returns false",
			exporterType: domain.ExporterType("invalid"),
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.exporterType.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExporter_GetEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		targetIP    string
		port        int
		metricsPath string
		want        string
	}{
		{
			name:        "standard endpoint with default path",
			targetIP:    "192.168.1.100",
			port:        9100,
			metricsPath: "/metrics",
			want:        "http://192.168.1.100:9100/metrics",
		},
		{
			name:        "custom metrics path",
			targetIP:    "10.0.0.50",
			port:        9400,
			metricsPath: "/custom/metrics",
			want:        "http://10.0.0.50:9400/custom/metrics",
		},
		{
			name:        "dcgm exporter endpoint",
			targetIP:    "192.168.1.200",
			port:        9400,
			metricsPath: "/metrics",
			want:        "http://192.168.1.200:9400/metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := &domain.Exporter{
				Port:        tt.port,
				MetricsPath: tt.metricsPath,
			}
			target := domain.Target{
				IPAddress: tt.targetIP,
			}
			got := exporter.GetEndpoint(target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultPortForType(t *testing.T) {
	tests := []struct {
		name         string
		exporterType domain.ExporterType
		want         int
	}{
		{
			name:         "node exporter default port",
			exporterType: domain.ExporterTypeNodeExporter,
			want:         9100,
		},
		{
			name:         "dcgm exporter default port",
			exporterType: domain.ExporterTypeDCGMExporter,
			want:         9400,
		},
		{
			name:         "custom exporter default port",
			exporterType: domain.ExporterTypeCustom,
			want:         9090,
		},
		{
			name:         "invalid type defaults to 9090",
			exporterType: domain.ExporterType("invalid"),
			want:         9090,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.DefaultPortForType(tt.exporterType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExporter_Creation(t *testing.T) {
	targetID := "target-123"
	exporter := testutil.NewTestExporter(targetID, domain.ExporterTypeNodeExporter)

	assert.NotEmpty(t, exporter.ID)
	assert.Equal(t, targetID, exporter.TargetID)
	assert.Equal(t, domain.ExporterTypeNodeExporter, exporter.Type)
	assert.Equal(t, 9100, exporter.Port) // Node exporter default port
	assert.True(t, exporter.Enabled)
	assert.Equal(t, "/metrics", exporter.MetricsPath)
	assert.Equal(t, "15s", exporter.ScrapeInterval)
	assert.Equal(t, "10s", exporter.ScrapeTimeout)
	assert.NotNil(t, exporter.Config)
	assert.NotZero(t, exporter.CreatedAt)
	assert.NotZero(t, exporter.UpdatedAt)
}

func TestExporter_TypeConstants(t *testing.T) {
	// Verify the constant values are as expected
	assert.Equal(t, domain.ExporterType("node_exporter"), domain.ExporterTypeNodeExporter)
	assert.Equal(t, domain.ExporterType("dcgm_exporter"), domain.ExporterTypeDCGMExporter)
	assert.Equal(t, domain.ExporterType("custom"), domain.ExporterTypeCustom)
}

func TestExporter_AllTypes(t *testing.T) {
	types := []domain.ExporterType{
		domain.ExporterTypeNodeExporter,
		domain.ExporterTypeDCGMExporter,
		domain.ExporterTypeCustom,
	}

	for _, exporterType := range types {
		t.Run(string(exporterType), func(t *testing.T) {
			assert.True(t, exporterType.IsValid())

			// Test that each type has a valid default port
			port := domain.DefaultPortForType(exporterType)
			assert.Greater(t, port, 0)
			assert.Less(t, port, 65536)
		})
	}
}

func TestExporter_GetEndpointWithTarget(t *testing.T) {
	// Integration test: exporter endpoint generation with target IP
	groupID := "group-123"
	target := testutil.NewTestTarget("server-01", groupID)
	exporter := testutil.NewTestExporter(target.ID, domain.ExporterTypeNodeExporter)

	endpoint := exporter.GetEndpoint(*target)

	assert.Contains(t, endpoint, target.IPAddress)
	assert.Contains(t, endpoint, "9100") // Node exporter port
	assert.Contains(t, endpoint, "/metrics")
	assert.True(t, strings.HasPrefix(endpoint, "http://"))
}
