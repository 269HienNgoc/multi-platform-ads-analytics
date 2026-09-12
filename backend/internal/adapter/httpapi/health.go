package httpapi

import (
	"net/http"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/health"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type healthHandler struct {
	service *health.Service
	logger  *zap.Logger
}

type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func newHealthHandler(service *health.Service, logger *zap.Logger) *healthHandler {
	return &healthHandler{service: service, logger: logger}
}

func (h *healthHandler) liveness(c *gin.Context) {
	result := h.service.Liveness()
	c.JSON(http.StatusOK, healthResponse{Status: result.Status, Checks: result.Checks})
}

func (h *healthHandler) readiness(c *gin.Context) {
	result, err := h.service.Readiness(c.Request.Context())
	if err != nil {
		h.logger.Warn(
			"readiness check failed",
			zap.String("request_id", requestID(c)),
			zap.Error(err),
		)
		c.JSON(http.StatusServiceUnavailable, healthResponse{Status: result.Status, Checks: result.Checks})

		return
	}

	c.JSON(http.StatusOK, healthResponse{Status: result.Status, Checks: result.Checks})
}
