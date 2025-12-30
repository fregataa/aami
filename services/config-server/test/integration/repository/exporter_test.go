package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestExporterRepository_Create(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("production")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server1", "192.168.1.10", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Create exporter
	exporter := testutil.NewTestExporter(target.ID, domain.ExporterTypeNodeExporter, 9100)

	err := exporterRepo.Create(ctx, exporter)
	require.NoError(t, err)
	assert.NotEmpty(t, exporter.ID)
	assert.True(t, exporter.Enabled)
}

func TestExporterRepository_GetByID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group, target, and exporter
	group := testutil.NewTestGroup("staging")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server2", "192.168.1.20", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	exporter := testutil.NewTestExporter(target.ID, domain.ExporterTypeDCGMExporter, 9400)
	require.NoError(t, exporterRepo.Create(ctx, exporter))

	// Retrieve it
	retrieved, err := exporterRepo.GetByID(ctx, exporter.ID)
	require.NoError(t, err)
	assert.Equal(t, exporter.ID, retrieved.ID)
	assert.Equal(t, exporter.TargetID, retrieved.TargetID)
	assert.Equal(t, exporter.Type, retrieved.Type)
	assert.Equal(t, exporter.Port, retrieved.Port)
}

func TestExporterRepository_GetByID_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	_, err := exporterRepo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
}

func TestExporterRepository_Update(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group, target, and exporter
	group := testutil.NewTestGroup("dev")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server3", "192.168.1.30", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	exporter := testutil.NewTestExporter(target.ID, domain.ExporterTypeNodeExporter, 9100)
	require.NoError(t, exporterRepo.Create(ctx, exporter))

	// Update it
	exporter.ScrapeInterval = "30s"
	if exporter.Config.CustomParams == nil {
		exporter.Config.CustomParams = make(map[string]interface{})
	}
	exporter.Config.CustomParams["custom_option"] = "value"
	err := exporterRepo.Update(ctx, exporter)
	require.NoError(t, err)

	// Verify update
	retrieved, err := exporterRepo.GetByID(ctx, exporter.ID)
	require.NoError(t, err)
	assert.Equal(t, "30s", retrieved.ScrapeInterval)
	assert.Equal(t, "value", retrieved.Config.CustomParams["custom_option"])
}

func TestExporterRepository_Delete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group, target, and exporter
	group := testutil.NewTestGroup("temp")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server-temp", "192.168.1.99", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	exporter := testutil.NewTestExporter(target.ID, domain.ExporterTypeCustom, 9090)
	require.NoError(t, exporterRepo.Create(ctx, exporter))

	// Delete it
	err := exporterRepo.Delete(ctx, exporter.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = exporterRepo.GetByID(ctx, exporter.ID)
	assert.Error(t, err)
}

func TestExporterRepository_GetByTargetID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Create multiple exporters for the same target
	exporter1 := testutil.NewTestExporter(target.ID, domain.ExporterTypeNodeExporter, 9100)
	exporter2 := testutil.NewTestExporter(target.ID, domain.ExporterTypeDCGMExporter, 9400)
	exporter3 := testutil.NewTestExporter(target.ID, domain.ExporterTypeCustom, 9090)

	require.NoError(t, exporterRepo.Create(ctx, exporter1))
	require.NoError(t, exporterRepo.Create(ctx, exporter2))
	require.NoError(t, exporterRepo.Create(ctx, exporter3))

	// Get exporters by target ID
	exporters, err := exporterRepo.GetByTargetID(ctx, target.ID)
	require.NoError(t, err)
	assert.Len(t, exporters, 3)

	// Verify all exporters belong to the target
	for _, e := range exporters {
		assert.Equal(t, target.ID, e.TargetID)
	}
}

func TestExporterRepository_GetByType(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target1 := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group})
	target2 := testutil.NewTestTarget("server2", "192.168.1.2", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target1))
	require.NoError(t, targetRepo.Create(ctx, target2))

	// Create exporters of different types
	nodeExporter1 := testutil.NewTestExporter(target1.ID, domain.ExporterTypeNodeExporter, 9100)
	nodeExporter2 := testutil.NewTestExporter(target2.ID, domain.ExporterTypeNodeExporter, 9100)
	dcgmExporter := testutil.NewTestExporter(target1.ID, domain.ExporterTypeDCGMExporter, 9400)

	require.NoError(t, exporterRepo.Create(ctx, nodeExporter1))
	require.NoError(t, exporterRepo.Create(ctx, nodeExporter2))
	require.NoError(t, exporterRepo.Create(ctx, dcgmExporter))

	// Get node exporters
	nodeExporters, err := exporterRepo.GetByType(ctx, domain.ExporterTypeNodeExporter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(nodeExporters), 2)

	// Verify all are node exporters
	for _, e := range nodeExporters {
		assert.Equal(t, domain.ExporterTypeNodeExporter, e.Type)
	}
}

func TestExporterRepository_List(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Create multiple exporters
	exporter1 := testutil.NewTestExporter(target.ID, domain.ExporterTypeNodeExporter, 9100)
	exporter2 := testutil.NewTestExporter(target.ID, domain.ExporterTypeDCGMExporter, 9400)

	require.NoError(t, exporterRepo.Create(ctx, exporter1))
	require.NoError(t, exporterRepo.Create(ctx, exporter2))

	// List all exporters
	exporters, total, err := exporterRepo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(exporters), 2)
	assert.GreaterOrEqual(t, total, 2)
}

func TestExporterRepository_Validation(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Create exporter with validation
	exporter := testutil.NewTestExporter(target.ID, domain.ExporterTypeNodeExporter, 9100)
	exporter.MetricsPath = "" // Will be set to default by Validate()
	exporter.ScrapeInterval = ""
	exporter.ScrapeTimeout = ""

	// Validate before creating
	err := exporter.Validate()
	require.NoError(t, err)
	assert.Equal(t, "/metrics", exporter.MetricsPath)
	assert.Equal(t, "15s", exporter.ScrapeInterval)
	assert.Equal(t, "10s", exporter.ScrapeTimeout)

	// Create exporter
	err = exporterRepo.Create(ctx, exporter)
	require.NoError(t, err)

	// Verify validation worked
	retrieved, err := exporterRepo.GetByID(ctx, exporter.ID)
	require.NoError(t, err)
	assert.Equal(t, "/metrics", retrieved.MetricsPath)
	assert.Equal(t, "15s", retrieved.ScrapeInterval)
	assert.Equal(t, "10s", retrieved.ScrapeTimeout)
}

func TestExporterRepository_GetEndpoint(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	exporterRepo := repoManager.Exporter
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server1", "192.168.1.10", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Create exporter
	exporter := testutil.NewTestExporter(target.ID, domain.ExporterTypeNodeExporter, 9100)
	exporter.MetricsPath = "/metrics"
	require.NoError(t, exporterRepo.Create(ctx, exporter))

	// Test GetEndpoint method
	endpoint := exporter.GetEndpoint(*target)
	assert.Equal(t, "http://192.168.1.10:9100/metrics", endpoint)
}
