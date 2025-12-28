package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/internal/repository"
	"gorm.io/gorm"
)

// HealthService handles health check operations
type HealthService struct {
	db        *gorm.DB
	startTime time.Time
	version   string
	repoMgr   *repository.Manager
}

// NewHealthService creates a new HealthService instance
func NewHealthService(db *gorm.DB, version string, repoMgr *repository.Manager) *HealthService {
	return &HealthService{
		db:        db,
		startTime: time.Now(),
		version:   version,
		repoMgr:   repoMgr,
	}
}

// CheckHealth performs a complete health check
func (s *HealthService) CheckHealth(ctx context.Context) *domain.HealthCheckResponse {
	components := make(map[string]domain.ComponentHealth)
	overallStatus := domain.HealthStatusHealthy

	// Check database
	dbHealth := s.checkDatabase(ctx)
	components["database"] = dbHealth
	if dbHealth.Status != domain.HealthStatusHealthy {
		overallStatus = domain.HealthStatusDegraded
		if dbHealth.Status == domain.HealthStatusUnhealthy {
			overallStatus = domain.HealthStatusUnhealthy
		}
	}

	// Check goroutines
	goroutineHealth := s.checkGoroutines()
	components["goroutines"] = goroutineHealth
	if goroutineHealth.Status == domain.HealthStatusDegraded {
		if overallStatus == domain.HealthStatusHealthy {
			overallStatus = domain.HealthStatusDegraded
		}
	}

	// Check memory
	memoryHealth := s.checkMemory()
	components["memory"] = memoryHealth
	if memoryHealth.Status == domain.HealthStatusDegraded {
		if overallStatus == domain.HealthStatusHealthy {
			overallStatus = domain.HealthStatusDegraded
		}
	}

	uptime := time.Since(s.startTime).Seconds()

	return &domain.HealthCheckResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    s.version,
		Components: components,
		Uptime:     uptime,
	}
}

// CheckReadiness performs a readiness check (critical components only)
func (s *HealthService) CheckReadiness(ctx context.Context) *domain.HealthCheckResponse {
	components := make(map[string]domain.ComponentHealth)
	overallStatus := domain.HealthStatusHealthy

	// Check database (critical)
	dbHealth := s.checkDatabase(ctx)
	components["database"] = dbHealth
	if dbHealth.Status == domain.HealthStatusUnhealthy {
		overallStatus = domain.HealthStatusUnhealthy
	}

	uptime := time.Since(s.startTime).Seconds()

	return &domain.HealthCheckResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    s.version,
		Components: components,
		Uptime:     uptime,
	}
}

// CheckLiveness performs a liveness check (basic server operation)
func (s *HealthService) CheckLiveness(ctx context.Context) *domain.HealthCheckResponse {
	components := make(map[string]domain.ComponentHealth)

	// Simple liveness check - if we can respond, we're alive
	components["server"] = domain.ComponentHealth{
		Name:      "server",
		Status:    domain.HealthStatusHealthy,
		Message:   "Server is responding",
		Timestamp: time.Now(),
	}

	// Check goroutine count (warn if too high)
	goroutineHealth := s.checkGoroutines()
	components["goroutines"] = goroutineHealth

	overallStatus := domain.HealthStatusHealthy
	if goroutineHealth.Status == domain.HealthStatusDegraded {
		overallStatus = domain.HealthStatusDegraded
	}

	uptime := time.Since(s.startTime).Seconds()

	return &domain.HealthCheckResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    s.version,
		Components: components,
		Uptime:     uptime,
	}
}

// checkDatabase checks the database connection
func (s *HealthService) checkDatabase(ctx context.Context) domain.ComponentHealth {
	health := domain.ComponentHealth{
		Name:      "database",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Get underlying SQL database
	sqlDB, err := s.db.DB()
	if err != nil {
		health.Status = domain.HealthStatusUnhealthy
		health.Error = fmt.Sprintf("Failed to get database: %v", err)
		return health
	}

	// Check if database is reachable
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		health.Status = domain.HealthStatusUnhealthy
		health.Error = fmt.Sprintf("Database ping failed: %v", err)
		return health
	}

	// Get database stats
	stats := sqlDB.Stats()
	health.Status = domain.HealthStatusHealthy
	health.Message = "Database connection is healthy"
	health.Metadata["open_connections"] = stats.OpenConnections
	health.Metadata["in_use"] = stats.InUse
	health.Metadata["idle"] = stats.Idle
	health.Metadata["max_open_connections"] = stats.MaxOpenConnections

	// Warn if connection pool is running low
	if stats.MaxOpenConnections > 0 && stats.OpenConnections >= stats.MaxOpenConnections*9/10 {
		health.Status = domain.HealthStatusDegraded
		health.Message = "Database connection pool is almost full"
	}

	return health
}

// checkGoroutines checks the number of goroutines
func (s *HealthService) checkGoroutines() domain.ComponentHealth {
	health := domain.ComponentHealth{
		Name:      "goroutines",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	numGoroutines := runtime.NumGoroutine()
	health.Metadata["count"] = numGoroutines

	// Warn if too many goroutines (potential leak)
	if numGoroutines > 10000 {
		health.Status = domain.HealthStatusDegraded
		health.Message = fmt.Sprintf("High goroutine count: %d", numGoroutines)
	} else {
		health.Status = domain.HealthStatusHealthy
		health.Message = fmt.Sprintf("Goroutine count is normal: %d", numGoroutines)
	}

	return health
}

// checkMemory checks memory usage
func (s *HealthService) checkMemory() domain.ComponentHealth {
	health := domain.ComponentHealth{
		Name:      "memory",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health.Metadata["alloc_mb"] = m.Alloc / 1024 / 1024
	health.Metadata["total_alloc_mb"] = m.TotalAlloc / 1024 / 1024
	health.Metadata["sys_mb"] = m.Sys / 1024 / 1024
	health.Metadata["num_gc"] = m.NumGC

	// Warn if memory usage is high (>1GB)
	if m.Alloc > 1024*1024*1024 {
		health.Status = domain.HealthStatusDegraded
		health.Message = fmt.Sprintf("High memory usage: %d MB", m.Alloc/1024/1024)
	} else {
		health.Status = domain.HealthStatusHealthy
		health.Message = fmt.Sprintf("Memory usage is normal: %d MB", m.Alloc/1024/1024)
	}

	return health
}
