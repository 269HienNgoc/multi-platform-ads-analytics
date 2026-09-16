package httpapi

import (
	"context"
	"errors"
	"net/http"

	applicationautomation "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/automation"
	automationdomain "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/automation"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type automationService interface {
	List(context.Context) ([]automationdomain.CampaignWorkflow, error)
	CreateBulk(
		context.Context,
		applicationautomation.CreateBulkInput,
	) (applicationautomation.CreateBulkResult, error)
	Transition(
		context.Context,
		string,
		automationdomain.WorkflowState,
	) (automationdomain.CampaignWorkflow, error)
	ApplyMetrics(
		context.Context,
		string,
		automationdomain.CampaignMetrics,
	) (automationdomain.CampaignWorkflow, error)
}

type automationHandler struct {
	service automationService
	logger  *zap.Logger
}

type createBulkWorkflowsRequest struct {
	OrganizationID    string   `json:"organization_id" binding:"required"`
	AdAccountIDs      []string `json:"ad_account_ids" binding:"required,min=1"`
	PageExternalID    string   `json:"page_external_id" binding:"required"`
	PixelExternalID   string   `json:"pixel_external_id"`
	PixelEvent        string   `json:"pixel_event"`
	ExistingPostID    string   `json:"existing_post_id"`
	SeedSpendLimitUSD float64  `json:"seed_spend_limit_usd"`
}

type transitionWorkflowRequest struct {
	State automationdomain.WorkflowState `json:"state" binding:"required"`
}

func newAutomationHandler(service automationService, logger *zap.Logger) *automationHandler {
	return &automationHandler{service: service, logger: logger}
}

func (h *automationHandler) listWorkflows(c *gin.Context) {
	workflows, err := h.service.List(c.Request.Context())
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflows})
}

func (h *automationHandler) createBulkWorkflows(c *gin.Context) {
	var request createBulkWorkflowsRequest
	if !bindJSON(c, &request) {
		return
	}
	requestKey := c.GetHeader("Idempotency-Key")
	if requestKey == "" {
		requestKey = requestID(c)
	}
	created, err := h.service.CreateBulk(c.Request.Context(), applicationautomation.CreateBulkInput{
		RequestKey:     requestKey,
		OrganizationID: request.OrganizationID, AdAccountIDs: request.AdAccountIDs,
		PageExternalID: request.PageExternalID, PixelExternalID: request.PixelExternalID,
		PixelEvent: request.PixelEvent, ExistingPostID: request.ExistingPostID,
		SeedSpendLimitUSD: request.SeedSpendLimitUSD,
	})
	if err != nil {
		h.writeError(c, err)

		return
	}
	if len(created.Failures) > 0 {
		c.JSON(http.StatusMultiStatus, gin.H{
			"data": created.Workflows, "count": len(created.Workflows), "failures": created.Failures,
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": created.Workflows, "count": len(created.Workflows), "failures": created.Failures})
}

func (h *automationHandler) transitionWorkflow(c *gin.Context) {
	var request transitionWorkflowRequest
	if !bindJSON(c, &request) {
		return
	}
	workflow, err := h.service.Transition(c.Request.Context(), c.Param("workflowID"), request.State)
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflow})
}

func (h *automationHandler) applyMetrics(c *gin.Context) {
	var metrics automationdomain.CampaignMetrics
	if !bindJSON(c, &metrics) {
		return
	}
	workflow, err := h.service.ApplyMetrics(c.Request.Context(), c.Param("workflowID"), metrics)
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflow})
}

func (h *automationHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, applicationautomation.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_request", "message": "invalid automation data"}})
	case errors.Is(err, applicationautomation.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "not_found", "message": "workflow or account not found"}})
	case errors.Is(err, automationdomain.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "invalid_transition", "message": "workflow transition is not allowed"},
		})
	default:
		h.logger.Error("automation request failed", zap.String("request_id", requestID(c)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal_error", "message": "internal server error"},
		})
	}
}
