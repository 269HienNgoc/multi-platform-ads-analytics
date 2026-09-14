package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/catalog"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxJSONBodyBytes = 1 << 20

type catalogService interface {
	CreateAccount(context.Context, ads.AdAccount) (ads.AdAccount, error)
	CreateCampaign(context.Context, ads.Campaign) (ads.Campaign, error)
	CreateAdGroup(context.Context, ads.AdGroup) (ads.AdGroup, error)
	CreateAd(context.Context, ads.Ad) (ads.Ad, error)
	CreateCreative(context.Context, ads.Creative) (ads.Creative, error)
	AccountHierarchy(context.Context, string) (ads.AccountHierarchy, error)
}

type catalogHandler struct {
	service catalogService
	logger  *zap.Logger
}

type createAccountRequest struct {
	Platform     string           `json:"platform" binding:"required"`
	ExternalID   string           `json:"external_id" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Currency     string           `json:"currency" binding:"required"`
	Timezone     string           `json:"timezone" binding:"required"`
	Status       ads.Status       `json:"status" binding:"required"`
	ProviderData ads.ProviderData `json:"provider_data"`
}

type createCampaignRequest struct {
	AccountID    string           `json:"account_id" binding:"required"`
	ExternalID   string           `json:"external_id" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Objective    string           `json:"objective"`
	Status       ads.Status       `json:"status" binding:"required"`
	ProviderData ads.ProviderData `json:"provider_data"`
}

type createAdGroupRequest struct {
	CampaignID   string           `json:"campaign_id" binding:"required"`
	ExternalID   string           `json:"external_id" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Status       ads.Status       `json:"status" binding:"required"`
	ProviderData ads.ProviderData `json:"provider_data"`
}

type createAdRequest struct {
	AdGroupID    string           `json:"ad_group_id" binding:"required"`
	ExternalID   string           `json:"external_id" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Status       ads.Status       `json:"status" binding:"required"`
	ProviderData ads.ProviderData `json:"provider_data"`
}

type createCreativeRequest struct {
	AdID         string           `json:"ad_id" binding:"required"`
	ExternalID   string           `json:"external_id" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Format       string           `json:"format" binding:"required"`
	AssetURL     string           `json:"asset_url"`
	ProviderData ads.ProviderData `json:"provider_data"`
}

type accountResponse struct {
	ID           string           `json:"id"`
	Platform     ads.Platform     `json:"platform"`
	ExternalID   string           `json:"external_id"`
	Name         string           `json:"name"`
	Currency     string           `json:"currency"`
	Timezone     string           `json:"timezone"`
	Status       ads.Status       `json:"status"`
	ProviderData ads.ProviderData `json:"provider_data"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type campaignResponse struct {
	ID           string           `json:"id"`
	AccountID    string           `json:"account_id"`
	ExternalID   string           `json:"external_id"`
	Name         string           `json:"name"`
	Objective    string           `json:"objective"`
	Status       ads.Status       `json:"status"`
	ProviderData ads.ProviderData `json:"provider_data"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type adGroupResponse struct {
	ID           string           `json:"id"`
	CampaignID   string           `json:"campaign_id"`
	ExternalID   string           `json:"external_id"`
	Name         string           `json:"name"`
	Status       ads.Status       `json:"status"`
	ProviderData ads.ProviderData `json:"provider_data"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type adResponse struct {
	ID           string           `json:"id"`
	AdGroupID    string           `json:"ad_group_id"`
	ExternalID   string           `json:"external_id"`
	Name         string           `json:"name"`
	Status       ads.Status       `json:"status"`
	ProviderData ads.ProviderData `json:"provider_data"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type creativeResponse struct {
	ID           string           `json:"id"`
	AdID         string           `json:"ad_id"`
	ExternalID   string           `json:"external_id"`
	Name         string           `json:"name"`
	Format       string           `json:"format"`
	AssetURL     string           `json:"asset_url"`
	ProviderData ads.ProviderData `json:"provider_data"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type hierarchyResponse struct {
	Account   accountResponse        `json:"account"`
	Campaigns []campaignNodeResponse `json:"campaigns"`
}

type campaignNodeResponse struct {
	Campaign campaignResponse      `json:"campaign"`
	AdGroups []adGroupNodeResponse `json:"ad_groups"`
}

type adGroupNodeResponse struct {
	AdGroup adGroupResponse  `json:"ad_group"`
	Ads     []adNodeResponse `json:"ads"`
}

type adNodeResponse struct {
	Ad        adResponse         `json:"ad"`
	Creatives []creativeResponse `json:"creatives"`
}

func newCatalogHandler(service catalogService, logger *zap.Logger) *catalogHandler {
	return &catalogHandler{service: service, logger: logger}
}

func (h *catalogHandler) createAccount(c *gin.Context) {
	var request createAccountRequest
	if !bindJSON(c, &request) {
		return
	}
	platform, err := ads.ParsePlatform(request.Platform)
	if err != nil {
		h.writeError(c, catalog.ErrInvalidInput)

		return
	}
	account, err := h.service.CreateAccount(c.Request.Context(), ads.AdAccount{
		Platform: platform, ExternalID: request.ExternalID, Name: request.Name,
		Currency: request.Currency, Timezone: request.Timezone, Status: request.Status,
		ProviderData: request.ProviderData,
	})
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusCreated, accountToResponse(account))
}

func (h *catalogHandler) createCampaign(c *gin.Context) {
	var request createCampaignRequest
	if !bindJSON(c, &request) {
		return
	}
	campaign, err := h.service.CreateCampaign(c.Request.Context(), ads.Campaign{
		AccountID: request.AccountID, ExternalID: request.ExternalID, Name: request.Name,
		Objective: request.Objective, Status: request.Status, ProviderData: request.ProviderData,
	})
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusCreated, campaignToResponse(campaign))
}

func (h *catalogHandler) createAdGroup(c *gin.Context) {
	var request createAdGroupRequest
	if !bindJSON(c, &request) {
		return
	}
	group, err := h.service.CreateAdGroup(c.Request.Context(), ads.AdGroup{
		CampaignID: request.CampaignID, ExternalID: request.ExternalID,
		Name: request.Name, Status: request.Status, ProviderData: request.ProviderData,
	})
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusCreated, adGroupToResponse(group))
}

func (h *catalogHandler) createAd(c *gin.Context) {
	var request createAdRequest
	if !bindJSON(c, &request) {
		return
	}
	ad, err := h.service.CreateAd(c.Request.Context(), ads.Ad{
		AdGroupID: request.AdGroupID, ExternalID: request.ExternalID,
		Name: request.Name, Status: request.Status, ProviderData: request.ProviderData,
	})
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusCreated, adToResponse(ad))
}

func (h *catalogHandler) createCreative(c *gin.Context) {
	var request createCreativeRequest
	if !bindJSON(c, &request) {
		return
	}
	creative, err := h.service.CreateCreative(c.Request.Context(), ads.Creative{
		AdID: request.AdID, ExternalID: request.ExternalID, Name: request.Name,
		Format: request.Format, AssetURL: request.AssetURL, ProviderData: request.ProviderData,
	})
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusCreated, creativeToResponse(creative))
}

func (h *catalogHandler) accountHierarchy(c *gin.Context) {
	hierarchy, err := h.service.AccountHierarchy(c.Request.Context(), c.Param("accountID"))
	if err != nil {
		h.writeError(c, err)

		return
	}

	c.JSON(http.StatusOK, hierarchyToResponse(hierarchy))
}

func bindJSON(c *gin.Context, destination any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBodyBytes)
	if err := c.ShouldBindJSON(destination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_request", "message": "invalid json request"}})

		return false
	}

	return true
}

func (h *catalogHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, catalog.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_request", "message": "invalid catalog data"}})
	case errors.Is(err, catalog.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "not_found", "message": "catalog entity not found"}})
	case errors.Is(err, catalog.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "conflict", "message": "catalog entity already exists"}})
	default:
		h.logger.Error("catalog request failed", zap.String("request_id", requestID(c)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal_error", "message": "internal server error"}})
	}
}

func accountToResponse(account ads.AdAccount) accountResponse {
	return accountResponse{
		ID: account.ID, Platform: account.Platform, ExternalID: account.ExternalID,
		Name: account.Name, Currency: account.Currency, Timezone: account.Timezone,
		Status: account.Status, ProviderData: account.ProviderData,
		CreatedAt: account.CreatedAt, UpdatedAt: account.UpdatedAt,
	}
}

func campaignToResponse(campaign ads.Campaign) campaignResponse {
	return campaignResponse{
		ID: campaign.ID, AccountID: campaign.AccountID, ExternalID: campaign.ExternalID,
		Name: campaign.Name, Objective: campaign.Objective, Status: campaign.Status,
		ProviderData: campaign.ProviderData, CreatedAt: campaign.CreatedAt, UpdatedAt: campaign.UpdatedAt,
	}
}

func adGroupToResponse(group ads.AdGroup) adGroupResponse {
	return adGroupResponse{
		ID: group.ID, CampaignID: group.CampaignID, ExternalID: group.ExternalID,
		Name: group.Name, Status: group.Status, ProviderData: group.ProviderData,
		CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt,
	}
}

func adToResponse(ad ads.Ad) adResponse {
	return adResponse{
		ID: ad.ID, AdGroupID: ad.AdGroupID, ExternalID: ad.ExternalID,
		Name: ad.Name, Status: ad.Status, ProviderData: ad.ProviderData,
		CreatedAt: ad.CreatedAt, UpdatedAt: ad.UpdatedAt,
	}
}

func creativeToResponse(creative ads.Creative) creativeResponse {
	return creativeResponse{
		ID: creative.ID, AdID: creative.AdID, ExternalID: creative.ExternalID,
		Name: creative.Name, Format: creative.Format, AssetURL: creative.AssetURL,
		ProviderData: creative.ProviderData, CreatedAt: creative.CreatedAt, UpdatedAt: creative.UpdatedAt,
	}
}

func hierarchyToResponse(hierarchy ads.AccountHierarchy) hierarchyResponse {
	response := hierarchyResponse{
		Account: accountToResponse(hierarchy.Account),
		Campaigns: make([]campaignNodeResponse, 0, len(hierarchy.Campaigns)),
	}
	for _, campaignNode := range hierarchy.Campaigns {
		campaign := campaignNodeResponse{
			Campaign: campaignToResponse(campaignNode.Campaign),
			AdGroups: make([]adGroupNodeResponse, 0, len(campaignNode.AdGroups)),
		}
		for _, groupNode := range campaignNode.AdGroups {
			group := adGroupNodeResponse{
				AdGroup: adGroupToResponse(groupNode.AdGroup),
				Ads: make([]adNodeResponse, 0, len(groupNode.Ads)),
			}
			for _, adNode := range groupNode.Ads {
				adResponseNode := adNodeResponse{
					Ad: adToResponse(adNode.Ad),
					Creatives: make([]creativeResponse, 0, len(adNode.Creatives)),
				}
				for _, creative := range adNode.Creatives {
					adResponseNode.Creatives = append(adResponseNode.Creatives, creativeToResponse(creative))
				}
				group.Ads = append(group.Ads, adResponseNode)
			}
			campaign.AdGroups = append(campaign.AdGroups, group)
		}
		response.Campaigns = append(response.Campaigns, campaign)
	}

	return response
}
