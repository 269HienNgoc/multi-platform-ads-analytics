package meta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/platform"
)

var ErrNotImplemented = errors.New("meta operation not implemented yet")

type Client struct {
	BaseURL     string
	Version     string
	AccessToken string
	HTTPClient  *http.Client
}

func NewClient(baseURL, version, accessToken string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Version: version,
		AccessToken: accessToken,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) SyncAssets(ctx context.Context) (platform.AssetSnapshot, error) {
	if c.AccessToken == "" {
		return platform.AssetSnapshot{}, errors.New("META_ACCESS_TOKEN is required")
	}
	var response struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Status int `json:"account_status"`
		} `json:"data"`
	}
	if err := c.get(ctx, "/me/adaccounts", url.Values{"fields": {"id,name,account_status"}}, &response); err != nil {
		return platform.AssetSnapshot{}, err
	}

	snapshot := platform.AssetSnapshot{}
	for _, account := range response.Data {
		snapshot.AdAccounts = append(snapshot.AdAccounts, platform.ExternalAsset{
			ExternalID: account.ID,
			Name: account.Name,
			Status: fmt.Sprintf("%d", account.Status),
		})
	}
	return snapshot, nil
}

func (c *Client) CreateSeedCampaign(context.Context, platform.CreateSeedInput) (string, error) {
	return "", ErrNotImplemented
}

func (c *Client) CreateMainCampaign(context.Context, platform.CreateMainInput) (string, error) {
	return "", ErrNotImplemented
}

func (c *Client) PauseCampaign(ctx context.Context, externalCampaignID string) error {
	values := url.Values{"status": {"PAUSED"}}
	var response map[string]any
	return c.postForm(ctx, "/"+externalCampaignID, values, &response)
}

func (c *Client) UpdateCampaignBudget(context.Context, string, int64) error {
	return ErrNotImplemented
}

func (c *Client) GetCampaignMetrics(context.Context, string) (domain.CampaignMetrics, error) {
	return domain.CampaignMetrics{}, ErrNotImplemented
}

func (c *Client) get(ctx context.Context, path string, query url.Values, target any) error {
	query.Set("access_token", c.AccessToken)
	requestURL := fmt.Sprintf("%s/%s%s?%s", c.BaseURL, c.Version, path, query.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	return c.do(req, target)
}

func (c *Client) postForm(ctx context.Context, path string, values url.Values, target any) error {
	values.Set("access_token", c.AccessToken)
	requestURL := fmt.Sprintf("%s/%s%s", c.BaseURL, c.Version, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, target)
}

func (c *Client) do(req *http.Request, target any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var graphErr any
		_ = json.NewDecoder(resp.Body).Decode(&graphErr)
		return fmt.Errorf("meta graph api returned %s: %v", resp.Status, graphErr)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
