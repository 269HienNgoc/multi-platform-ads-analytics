package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	applicationprovidersync "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/providersync"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProviderSyncService exposes the read-only provider synchronization use cases.
type ProviderSyncService interface {
	Sync(context.Context) (applicationprovidersync.Run, applicationprovidersync.ApplyResult, error)
	ListRuns(context.Context, int) ([]applicationprovidersync.Run, error)
}

type providerSyncHandler struct {
	service ProviderSyncService
	logger  *zap.Logger
}

func newProviderSyncHandler(service ProviderSyncService, logger *zap.Logger) *providerSyncHandler {
	return &providerSyncHandler{service: service, logger: logger}
}

func (h *providerSyncHandler) status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"platform": "meta",
			"enabled":  h.service != nil,
			"mode":     "read_only",
		},
	})
}

func (h *providerSyncHandler) sync(c *gin.Context) {
	if h.service == nil {
		h.writeError(c, applicationprovidersync.ErrUnavailable)

		return
	}
	run, result, err := h.service.Sync(c.Request.Context())
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": run, "summary": result})
}

func (h *providerSyncHandler) listRuns(c *gin.Context) {
	if h.service == nil {
		h.writeError(c, applicationprovidersync.ErrUnavailable)

		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "invalid_request", "message": "invalid sync run limit"},
		})

		return
	}
	runs, err := h.service.ListRuns(c.Request.Context(), limit)
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": runs})
}

func (h *providerSyncHandler) writeError(c *gin.Context, err error) {
	if errors.Is(err, applicationprovidersync.ErrUnavailable) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{"code": "connector_unavailable", "message": "meta connector is not configured"},
		})

		return
	}

	h.logger.Error("provider sync request failed", zap.String("request_id", requestID(c)), zap.Error(err))
	c.JSON(http.StatusBadGateway, gin.H{
		"error": gin.H{"code": "provider_sync_failed", "message": "could not synchronize data from Meta"},
	})
}
