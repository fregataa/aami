package api

import (
	"github.com/fregataa/aami/config-server/internal/api/handler"
	"github.com/fregataa/aami/config-server/internal/api/middleware"
	"github.com/fregataa/aami/config-server/internal/pkg/alertmanager"
	"github.com/fregataa/aami/config-server/internal/pkg/jobmanager"
	"github.com/fregataa/aami/config-server/internal/pkg/prometheus"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// Server represents the API server
type Server struct {
	router             *gin.Engine
	rm                 *repository.Manager
	ruleGenerator      *service.PrometheusRuleGenerator
	fileManager        *prometheus.RuleFileManager
	prometheusClient   *prometheus.PrometheusClient
	alertmanagerClient *alertmanager.AlertmanagerClient
	jobManager         *jobmanager.Manager
}

// NewServer creates a new API server
func NewServer(rm *repository.Manager) *Server {
	return &Server{
		router: gin.New(),
		rm:     rm,
	}
}

// NewServerWithPrometheus creates a new API server with Prometheus components
func NewServerWithPrometheus(
	rm *repository.Manager,
	ruleGenerator *service.PrometheusRuleGenerator,
	fileManager *prometheus.RuleFileManager,
	prometheusClient *prometheus.PrometheusClient,
) *Server {
	return &Server{
		router:           gin.New(),
		rm:               rm,
		ruleGenerator:    ruleGenerator,
		fileManager:      fileManager,
		prometheusClient: prometheusClient,
	}
}

// NewServerWithAlertmanager creates a new API server with Prometheus and Alertmanager components
func NewServerWithAlertmanager(
	rm *repository.Manager,
	ruleGenerator *service.PrometheusRuleGenerator,
	fileManager *prometheus.RuleFileManager,
	prometheusClient *prometheus.PrometheusClient,
	alertmanagerClient *alertmanager.AlertmanagerClient,
) *Server {
	return &Server{
		router:             gin.New(),
		rm:                 rm,
		ruleGenerator:      ruleGenerator,
		fileManager:        fileManager,
		prometheusClient:   prometheusClient,
		alertmanagerClient: alertmanagerClient,
	}
}

// SetAlertmanagerClient sets the Alertmanager client (optional)
func (s *Server) SetAlertmanagerClient(client *alertmanager.AlertmanagerClient) {
	s.alertmanagerClient = client
}

// SetJobManager sets the Job Manager (optional)
func (s *Server) SetJobManager(manager *jobmanager.Manager) {
	s.jobManager = manager
}

// JobManager returns the job manager instance
func (s *Server) JobManager() *jobmanager.Manager {
	return s.jobManager
}

// SetupRouter configures all routes and middleware
func (s *Server) SetupRouter() *gin.Engine {
	// Global middleware
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.Logger())
	s.router.Use(middleware.CORS())

	// Initialize health service
	healthService := service.NewHealthService(s.rm.GetDB(), "v1.0.0", s.rm)
	healthHandler := handler.NewHealthHandler(healthService)

	// Health check endpoints (support both GET and HEAD for compatibility)
	s.router.GET("/health", healthHandler.CheckHealth)
	s.router.HEAD("/health", healthHandler.CheckHealth)
	s.router.GET("/health/ready", healthHandler.CheckReadiness)
	s.router.HEAD("/health/ready", healthHandler.CheckReadiness)
	s.router.GET("/health/live", healthHandler.CheckLiveness)
	s.router.HEAD("/health/live", healthHandler.CheckLiveness)

	// Initialize services
	groupService := service.NewGroupService(s.rm.Group)
	targetService := service.NewTargetService(s.rm.Target, s.rm.TargetGroup, s.rm.Group)
	exporterService := service.NewExporterService(s.rm.Exporter, s.rm.Target)
	alertTemplateService := service.NewAlertTemplateService(s.rm.AlertTemplate)
	// Use Prometheus components if available, include target repo for effective rules
	alertRuleService := service.NewAlertRuleServiceWithTargetRepo(
		s.rm.AlertRule, s.rm.AlertTemplate, s.rm.Group, s.rm.Target,
		s.ruleGenerator, s.prometheusClient,
	)
	scriptTemplateService := service.NewScriptTemplateService(s.rm.ScriptTemplate, s.rm.ScriptPolicy)
	scriptPolicyService := service.NewScriptPolicyService(s.rm.ScriptPolicy, s.rm.ScriptTemplate, s.rm.Group, s.rm.Target)
	bootstrapTokenService := service.NewBootstrapTokenService(s.rm.BootstrapToken, s.rm.Group, targetService)
	serviceDiscoveryService := service.NewServiceDiscoveryService(s.rm.Target)

	// Initialize handlers
	groupHandler := handler.NewGroupHandler(groupService)
	targetHandler := handler.NewTargetHandler(targetService)
	exporterHandler := handler.NewExporterHandler(exporterService)
	alertTemplateHandler := handler.NewAlertTemplateHandler(alertTemplateService)
	alertRuleHandler := handler.NewAlertRuleHandler(alertRuleService)
	scriptTemplateHandler := handler.NewScriptTemplateHandler(scriptTemplateService)
	scriptPolicyHandler := handler.NewScriptPolicyHandler(scriptPolicyService)
	bootstrapTokenHandler := handler.NewBootstrapTokenHandler(bootstrapTokenService)
	serviceDiscoveryHandler := handler.NewServiceDiscoveryHandler(serviceDiscoveryService)
	prometheusRuleHandler := handler.NewPrometheusRuleHandlerWithAlertService(
		s.ruleGenerator, s.fileManager, s.prometheusClient, alertRuleService,
	)

	// Initialize Alertmanager service and handler (optional)
	alertmanagerService := service.NewAlertmanagerService(s.alertmanagerClient)
	activeAlertsHandler := handler.NewActiveAlertsHandler(alertmanagerService)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Group routes
		groups := v1.Group("/groups")
		{
			groups.POST("", groupHandler.Create)
			groups.GET("", groupHandler.List)
			groups.GET("/:id", groupHandler.GetByID)
			groups.PUT("/:id", groupHandler.Update)
			groups.POST("/delete", groupHandler.DeleteResource)
			groups.POST("/purge", groupHandler.PurgeResource)
			groups.POST("/restore", groupHandler.RestoreResource)
		}

		// Target routes
		targets := v1.Group("/targets")
		{
			targets.POST("", targetHandler.Create)
			targets.GET("", targetHandler.List)
			targets.GET("/:id", targetHandler.GetByID)
			targets.GET("/hostname/:hostname", targetHandler.GetByHostname)
			targets.PUT("/:id", targetHandler.Update)
			targets.POST("/delete", targetHandler.DeleteResource)
			targets.POST("/purge", targetHandler.PurgeResource)
			targets.POST("/restore", targetHandler.RestoreResource)
			targets.GET("/group/:group_id", targetHandler.GetByGroupID)
			targets.POST("/:id/status", targetHandler.UpdateStatus)
			targets.POST("/:id/heartbeat", targetHandler.Heartbeat)
		}

		// Exporter routes
		exporters := v1.Group("/exporters")
		{
			exporters.POST("", exporterHandler.Create)
			exporters.GET("", exporterHandler.List)
			exporters.GET("/:id", exporterHandler.GetByID)
			exporters.PUT("/:id", exporterHandler.Update)
			exporters.POST("/delete", exporterHandler.DeleteResource)
			exporters.POST("/purge", exporterHandler.PurgeResource)
			exporters.POST("/restore", exporterHandler.RestoreResource)
			exporters.GET("/target/:target_id", exporterHandler.GetByTargetID)
			exporters.GET("/type/:type", exporterHandler.GetByType)
		}

		// Alert template routes
		alertTemplates := v1.Group("/alert-templates")
		{
			alertTemplates.POST("", alertTemplateHandler.Create)
			alertTemplates.GET("", alertTemplateHandler.List)
			alertTemplates.GET("/:id", alertTemplateHandler.GetByID)
			alertTemplates.PUT("/:id", alertTemplateHandler.Update)
			alertTemplates.POST("/delete", alertTemplateHandler.DeleteResource)
			alertTemplates.POST("/purge", alertTemplateHandler.PurgeResource)
			alertTemplates.POST("/restore", alertTemplateHandler.RestoreResource)
			alertTemplates.GET("/severity/:severity", alertTemplateHandler.GetBySeverity)
		}

		// Alert rule routes
		alertRules := v1.Group("/alert-rules")
		{
			alertRules.POST("", alertRuleHandler.Create)
			alertRules.GET("", alertRuleHandler.List)
			alertRules.GET("/:id", alertRuleHandler.GetByID)
			alertRules.PUT("/:id", alertRuleHandler.Update)
			alertRules.POST("/delete", alertRuleHandler.DeleteResource)
			alertRules.POST("/purge", alertRuleHandler.PurgeResource)
			alertRules.POST("/restore", alertRuleHandler.RestoreResource)
			alertRules.GET("/group/:group_id", alertRuleHandler.GetByGroupID)
			alertRules.GET("/template/:template_id", alertRuleHandler.GetByTemplateID)
		}

		// Active alerts routes (from Alertmanager)
		alerts := v1.Group("/alerts")
		{
			alerts.GET("/active", activeAlertsHandler.GetActive)
		}

		// Bootstrap token routes
		bootstrapTokens := v1.Group("/bootstrap-tokens")
		{
			bootstrapTokens.POST("", bootstrapTokenHandler.Create)
			bootstrapTokens.GET("", bootstrapTokenHandler.List)
			bootstrapTokens.GET("/:id", bootstrapTokenHandler.GetByID)
			bootstrapTokens.GET("/token/:token", bootstrapTokenHandler.GetByToken)
			bootstrapTokens.POST("/validate", bootstrapTokenHandler.ValidateAndUse)
			bootstrapTokens.POST("/register", bootstrapTokenHandler.RegisterNode)
			bootstrapTokens.PUT("/:id", bootstrapTokenHandler.Update)
			bootstrapTokens.POST("/delete", bootstrapTokenHandler.DeleteResource)
			bootstrapTokens.POST("/purge", bootstrapTokenHandler.PurgeResource)
			bootstrapTokens.POST("/restore", bootstrapTokenHandler.RestoreResource)
		}

		// Script template routes
		scriptTemplates := v1.Group("/script-templates")
		{
			scriptTemplates.POST("", scriptTemplateHandler.Create)
			scriptTemplates.GET("", scriptTemplateHandler.List)
			scriptTemplates.GET("/active", scriptTemplateHandler.ListActive)
			scriptTemplates.GET("/:id", scriptTemplateHandler.GetByID)
			scriptTemplates.GET("/name/:name", scriptTemplateHandler.GetByName)
			scriptTemplates.GET("/type/:scriptType", scriptTemplateHandler.GetByScriptType)
			scriptTemplates.PUT("/:id", scriptTemplateHandler.Update)
			scriptTemplates.POST("/delete", scriptTemplateHandler.DeleteResource)
			scriptTemplates.POST("/purge", scriptTemplateHandler.PurgeResource)
			scriptTemplates.POST("/restore", scriptTemplateHandler.RestoreResource)
			scriptTemplates.GET("/:id/verify-hash", scriptTemplateHandler.VerifyHash)
		}

		// Script policy routes
		scriptPolicies := v1.Group("/script-policies")
		{
			scriptPolicies.POST("", scriptPolicyHandler.Create)
			scriptPolicies.GET("", scriptPolicyHandler.List)
			scriptPolicies.GET("/active", scriptPolicyHandler.ListActive)
			scriptPolicies.GET("/:id", scriptPolicyHandler.GetByID)
			scriptPolicies.GET("/template/:templateId", scriptPolicyHandler.GetByTemplateID)
			scriptPolicies.GET("/global", scriptPolicyHandler.GetGlobalInstances)
			scriptPolicies.GET("/group/:groupId", scriptPolicyHandler.GetByGroupID)
			scriptPolicies.GET("/effective/group/:groupId", scriptPolicyHandler.GetEffectiveChecksByGroup)
			scriptPolicies.PUT("/:id", scriptPolicyHandler.Update)
			scriptPolicies.POST("/delete", scriptPolicyHandler.DeleteResource)
			scriptPolicies.POST("/purge", scriptPolicyHandler.PurgeResource)
			scriptPolicies.POST("/restore", scriptPolicyHandler.RestoreResource)
		}

		// Node API: Get effective checks by target ID or hostname
		checks := v1.Group("/checks")
		{
			// Hostname-based lookup (for node scripts)
			checks.GET("/target/hostname/:hostname", scriptPolicyHandler.GetEffectiveChecksByHostname)
			// UUID-based lookup
			checks.GET("/target/:targetId", scriptPolicyHandler.GetEffectiveChecksByTargetID)
		}

		// Service Discovery routes
		sd := v1.Group("/sd")
		{
			// Prometheus HTTP Service Discovery
			sd.GET("/prometheus", serviceDiscoveryHandler.GetPrometheusTargets)
			sd.GET("/prometheus/active", serviceDiscoveryHandler.GetActivePrometheusTargets)
			sd.GET("/prometheus/group/:groupId", serviceDiscoveryHandler.GetPrometheusTargetsByGroup)

			// Prometheus File Service Discovery
			sd.POST("/prometheus/file", serviceDiscoveryHandler.GenerateFileSD)
			sd.POST("/prometheus/file/active", serviceDiscoveryHandler.GenerateActiveFileSD)
			sd.POST("/prometheus/file/group/:groupId", serviceDiscoveryHandler.GenerateGroupFileSD)
		}

		// Prometheus rule management routes
		prometheusRules := v1.Group("/prometheus")
		{
			prometheusRules.POST("/rules/regenerate", prometheusRuleHandler.RegenerateAllRules)
			prometheusRules.POST("/rules/regenerate/:group_id", prometheusRuleHandler.RegenerateGroupRules)
			prometheusRules.GET("/rules/files", prometheusRuleHandler.ListRuleFiles)
			prometheusRules.GET("/rules/effective/:target_id", prometheusRuleHandler.GetEffectiveRulesByTarget)
			prometheusRules.POST("/reload", prometheusRuleHandler.ReloadPrometheus)
			prometheusRules.GET("/status", prometheusRuleHandler.GetStatus)
		}

		// Job management routes (only if job manager is configured)
		if s.jobManager != nil {
			jobHandler := handler.NewJobHandler(s.jobManager)
			jobs := v1.Group("/jobs")
			{
				jobs.GET("", jobHandler.List)
				jobs.GET("/stats", jobHandler.GetStats)
				jobs.GET("/:id", jobHandler.GetByID)
				jobs.DELETE("/:id", jobHandler.Cancel)
			}
		}
	}

	return s.router
}

// Router returns the configured Gin engine
func (s *Server) Router() *gin.Engine {
	return s.router
}
