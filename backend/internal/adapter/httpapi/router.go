// Package httpapi exposes application use cases through Gin HTTP routes.
package httpapi

import (
	"net/http"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/health"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewServer creates a hardened HTTP server and registers application routes.
func NewServer(
	cfg config.Server,
	healthService *health.Service,
	catalogUseCases catalogService,
	automationUseCases automationService,
	providerSyncUseCases ProviderSyncService,
	logger *zap.Logger,
) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		requestMiddleware(logger),
		recoveryMiddleware(logger),
		securityHeadersMiddleware(),
	)

	healthHandler := newHealthHandler(healthService, logger)
	healthGroup := router.Group("/health")
	healthGroup.GET("/live", healthHandler.liveness)
	healthGroup.GET("/ready", healthHandler.readiness)

	catalogHandler := newCatalogHandler(catalogUseCases, logger)
	api := router.Group("/api/v1")
	api.Use(apiKeyMiddleware(cfg.APIKey))
	api.GET("/ad-accounts", catalogHandler.listAccounts)
	api.POST("/ad-accounts", catalogHandler.createAccount)
	api.POST("/campaigns", catalogHandler.createCampaign)
	api.POST("/ad-groups", catalogHandler.createAdGroup)
	api.POST("/ads", catalogHandler.createAd)
	api.POST("/creatives", catalogHandler.createCreative)
	api.GET("/ad-accounts/:accountID/hierarchy", catalogHandler.accountHierarchy)

	automationHandler := newAutomationHandler(automationUseCases, logger)
	api.GET("/workflows", automationHandler.listWorkflows)
	api.POST("/workflows/bulk", automationHandler.createBulkWorkflows)
	api.POST("/workflows/:workflowID/transition", automationHandler.transitionWorkflow)
	api.POST("/workflows/:workflowID/metrics", automationHandler.applyMetrics)

	providerSyncHandler := newProviderSyncHandler(providerSyncUseCases, logger)
	api.GET("/connectors/meta", providerSyncHandler.status)
	api.POST("/connectors/meta/sync", providerSyncHandler.sync)
	api.GET("/connectors/meta/sync-runs", providerSyncHandler.listRuns)

	return &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}
}
