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
func NewServer(cfg config.Server, healthService *health.Service, catalogUseCases catalogService, logger *zap.Logger) *http.Server {
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
	api.POST("/ad-accounts", catalogHandler.createAccount)
	api.POST("/campaigns", catalogHandler.createCampaign)
	api.POST("/ad-groups", catalogHandler.createAdGroup)
	api.POST("/ads", catalogHandler.createAd)
	api.POST("/creatives", catalogHandler.createCreative)
	api.GET("/ad-accounts/:accountID/hierarchy", catalogHandler.accountHierarchy)

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
