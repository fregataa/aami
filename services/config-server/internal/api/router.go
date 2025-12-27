package api

import (
	"github.com/fregataa/aami/config-server/internal/api/handler"
	"github.com/fregataa/aami/config-server/internal/api/middleware"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// Server represents the API server
type Server struct {
	router *gin.Engine
	rm     *repository.Manager
}

// NewServer creates a new API server
func NewServer(rm *repository.Manager) *Server {
	return &Server{
		router: gin.New(),
		rm:     rm,
	}
}

// SetupRouter configures all routes and middleware
func (s *Server) SetupRouter() *gin.Engine {
	// Global middleware
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.Logger())
	s.router.Use(middleware.CORS())

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Initialize services
	namespaceService := service.NewNamespaceService(s.rm.Namespace, s.rm.Group, s.rm.Target)
	groupService := service.NewGroupService(s.rm.Group, s.rm.Namespace)
	targetService := service.NewTargetService(s.rm.Target, s.rm.Group)
	exporterService := service.NewExporterService(s.rm.Exporter, s.rm.Target)
	alertTemplateService := service.NewAlertTemplateService(s.rm.AlertTemplate)
	alertRuleService := service.NewAlertRuleService(s.rm.AlertRule, s.rm.AlertTemplate, s.rm.Group)
	checkSettingService := service.NewCheckSettingService(s.rm.CheckSetting, s.rm.Group)
	bootstrapTokenService := service.NewBootstrapTokenService(s.rm.BootstrapToken, s.rm.Group)

	// Initialize handlers
	namespaceHandler := handler.NewNamespaceHandler(namespaceService)
	groupHandler := handler.NewGroupHandler(groupService)
	targetHandler := handler.NewTargetHandler(targetService)
	exporterHandler := handler.NewExporterHandler(exporterService)
	alertTemplateHandler := handler.NewAlertTemplateHandler(alertTemplateService)
	alertRuleHandler := handler.NewAlertRuleHandler(alertRuleService)
	checkSettingHandler := handler.NewCheckSettingHandler(checkSettingService)
	bootstrapTokenHandler := handler.NewBootstrapTokenHandler(bootstrapTokenService)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Namespace routes
		namespaces := v1.Group("/namespaces")
		{
			namespaces.POST("", namespaceHandler.Create)
			namespaces.GET("", namespaceHandler.List)
			namespaces.GET("/all", namespaceHandler.GetAll)
			namespaces.GET("/stats", namespaceHandler.GetAllStats)
			namespaces.GET("/:id", namespaceHandler.GetByID)
			namespaces.GET("/name/:name", namespaceHandler.GetByName)
			namespaces.PUT("/:id", namespaceHandler.Update)
			namespaces.POST("/delete", namespaceHandler.DeleteResource)
			namespaces.POST("/purge", namespaceHandler.PurgeResource)
			namespaces.POST("/restore", namespaceHandler.RestoreResource)
			namespaces.GET("/:id/stats", namespaceHandler.GetStats)
		}

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
			groups.GET("/namespace/:namespace_id", groupHandler.GetByNamespaceID)
			groups.GET("/:id/children", groupHandler.GetChildren)
			groups.GET("/:id/ancestors", groupHandler.GetAncestors)
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

		// Check setting routes
		checkSettings := v1.Group("/check-settings")
		{
			checkSettings.POST("", checkSettingHandler.Create)
			checkSettings.GET("", checkSettingHandler.List)
			checkSettings.GET("/:id", checkSettingHandler.GetByID)
			checkSettings.PUT("/:id", checkSettingHandler.Update)
			checkSettings.POST("/delete", checkSettingHandler.DeleteResource)
			checkSettings.POST("/purge", checkSettingHandler.PurgeResource)
			checkSettings.POST("/restore", checkSettingHandler.RestoreResource)
			checkSettings.GET("/group/:group_id", checkSettingHandler.GetByGroupID)
			checkSettings.GET("/type/:type", checkSettingHandler.GetByCheckType)
		}

		// Bootstrap token routes
		bootstrapTokens := v1.Group("/bootstrap-tokens")
		{
			bootstrapTokens.POST("", bootstrapTokenHandler.Create)
			bootstrapTokens.GET("", bootstrapTokenHandler.List)
			bootstrapTokens.GET("/:id", bootstrapTokenHandler.GetByID)
			bootstrapTokens.GET("/token/:token", bootstrapTokenHandler.GetByToken)
			bootstrapTokens.POST("/validate", bootstrapTokenHandler.ValidateAndUse)
			bootstrapTokens.PUT("/:id", bootstrapTokenHandler.Update)
			bootstrapTokens.POST("/delete", bootstrapTokenHandler.DeleteResource)
			bootstrapTokens.POST("/purge", bootstrapTokenHandler.PurgeResource)
			bootstrapTokens.POST("/restore", bootstrapTokenHandler.RestoreResource)
			bootstrapTokens.GET("/group/:group_id", bootstrapTokenHandler.GetByGroupID)
		}
	}

	return s.router
}

// Router returns the configured Gin engine
func (s *Server) Router() *gin.Engine {
	return s.router
}
