// Package meta implements the Meta Graph API advertising connector.
package meta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/connector"
	automationdomain "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/automation"
)

// ErrNotImplemented marks connector operations that are not safe to execute yet.
var ErrNotImplemented = errors.New("meta: operation not implemented")

var externalIDPattern = regexp.MustCompile(`^[A-Za-z0-9_:-]{1,255}$`)

var _ connector.AdvertisingPlatform = (*Client)(nil)

// Config contains the explicit dependencies required by a Meta Graph API client.
type Config struct {
	BaseURL     string
	Version     string
	AccessToken string
	Timeout     time.Duration
}

// Client calls Meta Graph API without exposing provider types to core packages.
type Client struct {
	baseURL     *url.URL
	version     string
	accessToken string
	httpClient  *http.Client
}

// New validates connector configuration and creates a bounded HTTP client.
func New(cfg Config) (*Client, error) {
	baseURL, err := url.Parse(strings.TrimRight(cfg.BaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parsing meta base url: %w", err)
	}
	isHTTP := baseURL.Scheme == "http" || baseURL.Scheme == "https"
	if !isHTTP || baseURL.Host == "" {
		return nil, errors.New("meta base url must be an absolute http or https url")
	}
	version := strings.Trim(strings.TrimSpace(cfg.Version), "/")
	if version == "" {
		return nil, errors.New("meta graph version is required")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: baseURL, version: version, accessToken: strings.TrimSpace(cfg.AccessToken),
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}, nil
}

// SyncAssets loads advertising accounts visible to the configured connection.
func (c *Client) SyncAssets(ctx context.Context) (connector.AssetSnapshot, error) {
	if c.accessToken == "" {
		return connector.AssetSnapshot{}, errors.New("meta access token is required")
	}
	var response struct {
		Data []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status int    `json:"account_status"`
		} `json:"data"`
	}
	query := url.Values{"fields": {"id,name,account_status"}}
	if err := c.get(ctx, "/me/adaccounts", query, &response); err != nil {
		return connector.AssetSnapshot{}, err
	}

	snapshot := connector.AssetSnapshot{
		AdAccounts: make([]connector.ExternalAsset, 0, len(response.Data)),
		Pages:      []connector.ExternalAsset{},
		Pixels:     []connector.ExternalAsset{},
	}
	for _, account := range response.Data {
		snapshot.AdAccounts = append(snapshot.AdAccounts, connector.ExternalAsset{
			ExternalID: account.ID,
			Name:       account.Name,
			Status:     fmt.Sprintf("%d", account.Status),
		})
	}

	return snapshot, nil
}

// CreateSeedCampaign is blocked until deterministic publish validation is implemented.
func (c *Client) CreateSeedCampaign(context.Context, connector.CreateSeedInput) (string, error) {
	return "", ErrNotImplemented
}

// CreateMainCampaign is blocked until deterministic publish validation is implemented.
func (c *Client) CreateMainCampaign(context.Context, connector.CreateMainInput) (string, error) {
	return "", ErrNotImplemented
}

// PauseCampaign pauses one external Meta campaign.
func (c *Client) PauseCampaign(ctx context.Context, externalCampaignID string) error {
	externalCampaignID = strings.TrimSpace(externalCampaignID)
	if !externalIDPattern.MatchString(externalCampaignID) {
		return errors.New("meta campaign id is invalid")
	}

	return c.postForm(ctx, "/"+externalCampaignID, url.Values{"status": {"PAUSED"}}, &struct{}{})
}

// UpdateCampaignBudget is blocked until budget guardrails are implemented.
func (c *Client) UpdateCampaignBudget(context.Context, string, int64) error {
	return ErrNotImplemented
}

// GetCampaignMetrics is blocked until insight attribution is implemented.
func (c *Client) GetCampaignMetrics(context.Context, string) (automationdomain.CampaignMetrics, error) {
	return automationdomain.CampaignMetrics{}, ErrNotImplemented
}

func (c *Client) get(ctx context.Context, path string, query url.Values, target any) error {
	requestURL := c.endpoint(path)
	requestURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("creating meta get request: %w", err)
	}

	return c.do(req, target)
}

func (c *Client) postForm(ctx context.Context, path string, values url.Values, target any) error {
	requestURL := c.endpoint(path)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL.String(),
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return fmt.Errorf("creating meta post request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.do(req, target)
}

func (c *Client) endpoint(path string) *url.URL {
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(c.baseURL.Path, "/") + "/" + c.version + "/" + strings.TrimLeft(path, "/")

	return &endpoint
}

func (c *Client) do(req *http.Request, target any) (resultErr error) {
	if c.accessToken == "" {
		return errors.New("meta access token is required")
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling meta graph api: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("closing meta graph response: %w", err))
		}
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		var graphError struct {
			Error struct {
				Type string `json:"type"`
				Code int    `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&graphError); err != nil {
			return fmt.Errorf("meta graph api returned status %d", response.StatusCode)
		}

		return fmt.Errorf(
			"meta graph api returned status %d, type %q, code %d",
			response.StatusCode,
			graphError.Error.Type,
			graphError.Error.Code,
		)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(target); err != nil {
		return fmt.Errorf("decoding meta graph response: %w", err)
	}

	return nil
}
