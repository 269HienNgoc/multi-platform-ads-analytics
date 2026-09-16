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
	"strconv"
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

type graphPage struct {
	Data   []json.RawMessage `json:"data"`
	Paging struct {
		Cursors struct {
			After string `json:"after"`
		} `json:"cursors"`
		Next string `json:"next"`
	} `json:"paging"`
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

// SyncAssets loads all visible accounts and their campaigns using cursor pagination.
func (c *Client) SyncAssets(ctx context.Context) (connector.AssetSnapshot, error) {
	if c.accessToken == "" {
		return connector.AssetSnapshot{}, errors.New("meta access token is required")
	}

	snapshot := connector.AssetSnapshot{
		AdAccounts: []connector.ExternalAdAccount{},
		Campaigns:  []connector.ExternalCampaign{},
		Pages:      []connector.ExternalAsset{},
		Pixels:     []connector.ExternalAsset{},
	}
	accountPayloads, err := c.getAll(
		ctx,
		"/me/adaccounts",
		url.Values{"fields": {"id,name,account_status,currency,timezone_name"}},
	)
	if err != nil {
		return connector.AssetSnapshot{}, fmt.Errorf("loading meta ad accounts: %w", err)
	}
	for _, payload := range accountPayloads {
		var account struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Status   int    `json:"account_status"`
			Currency string `json:"currency"`
			Timezone string `json:"timezone_name"`
		}
		if err := json.Unmarshal(payload, &account); err != nil {
			return connector.AssetSnapshot{}, fmt.Errorf("decoding meta ad account: %w", err)
		}
		if !externalIDPattern.MatchString(account.ID) {
			return connector.AssetSnapshot{}, fmt.Errorf("meta returned an invalid ad account id %q", account.ID)
		}
		snapshot.AdAccounts = append(snapshot.AdAccounts, connector.ExternalAdAccount{
			ExternalID: account.ID,
			Name:       account.Name,
			Currency:   strings.ToUpper(strings.TrimSpace(account.Currency)),
			Timezone:   account.Timezone,
			Status:     normalizeAccountStatus(account.Status),
			Raw:        payload,
		})

		campaigns, err := c.campaigns(ctx, account.ID)
		if err != nil {
			return connector.AssetSnapshot{}, fmt.Errorf("loading campaigns for meta account %s: %w", account.ID, err)
		}
		snapshot.Campaigns = append(snapshot.Campaigns, campaigns...)

		pixels, err := c.pixels(ctx, account.ID)
		if err != nil {
			snapshot.Warnings = append(
				snapshot.Warnings,
				fmt.Sprintf("pixels for account %s were not synchronized: %v", account.ID, err),
			)
		} else {
			snapshot.Pixels = append(snapshot.Pixels, pixels...)
		}
	}

	pages, err := c.assets(ctx, "/me/accounts")
	if err != nil {
		snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("pages were not synchronized: %v", err))
	} else {
		snapshot.Pages = pages
	}

	return snapshot, nil
}

func (c *Client) campaigns(ctx context.Context, accountID string) ([]connector.ExternalCampaign, error) {
	payloads, err := c.getAll(
		ctx,
		"/"+accountID+"/campaigns",
		url.Values{"fields": {"id,name,objective,status"}},
	)
	if err != nil {
		return nil, err
	}

	campaigns := make([]connector.ExternalCampaign, 0, len(payloads))
	for _, payload := range payloads {
		var campaign struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Objective string `json:"objective"`
			Status    string `json:"status"`
		}
		if err := json.Unmarshal(payload, &campaign); err != nil {
			return nil, fmt.Errorf("decoding meta campaign: %w", err)
		}
		campaigns = append(campaigns, connector.ExternalCampaign{
			AccountExternalID: accountID,
			ExternalID:        campaign.ID,
			Name:              campaign.Name,
			Objective:         strings.ToLower(campaign.Objective),
			Status:            normalizeCampaignStatus(campaign.Status),
			Raw:               payload,
		})
	}

	return campaigns, nil
}

func (c *Client) pixels(ctx context.Context, accountID string) ([]connector.ExternalAsset, error) {
	return c.assets(ctx, "/"+accountID+"/adspixels")
}

func (c *Client) assets(ctx context.Context, path string) ([]connector.ExternalAsset, error) {
	payloads, err := c.getAll(ctx, path, url.Values{"fields": {"id,name"}})
	if err != nil {
		return nil, err
	}

	assets := make([]connector.ExternalAsset, 0, len(payloads))
	for _, payload := range payloads {
		var asset struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(payload, &asset); err != nil {
			return nil, fmt.Errorf("decoding meta asset: %w", err)
		}
		assets = append(assets, connector.ExternalAsset{ExternalID: asset.ID, Name: asset.Name, Raw: payload})
	}

	return assets, nil
}

func (c *Client) getAll(ctx context.Context, path string, baseQuery url.Values) ([]json.RawMessage, error) {
	const maxPages = 1000

	query := cloneValues(baseQuery)
	query.Set("limit", "100")
	items := []json.RawMessage{}
	for range maxPages {
		var page graphPage
		if err := c.get(ctx, path, query, &page); err != nil {
			return nil, err
		}
		items = append(items, page.Data...)
		if page.Paging.Next == "" || page.Paging.Cursors.After == "" {
			return items, nil
		}
		query.Set("after", page.Paging.Cursors.After)
	}

	return nil, errors.New("meta pagination exceeded safety limit")
}

func cloneValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))
	for key, entries := range values {
		cloned[key] = append([]string(nil), entries...)
	}

	return cloned
}

func normalizeAccountStatus(status int) string {
	if status == 1 {
		return "active"
	}
	if status == 100 || status == 101 {
		return "archived"
	}

	return "paused"
}

func normalizeCampaignStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ACTIVE":
		return "active"
	case "ARCHIVED", "DELETED":
		return "archived"
	default:
		return "paused"
	}
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

func (c *Client) do(req *http.Request, target any) error {
	if c.accessToken == "" {
		return errors.New("meta access token is required")
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	const maxAttempts = 3
	for attempt := range maxAttempts {
		attemptRequest, err := requestForAttempt(req, attempt)
		if err != nil {
			return err
		}
		response, err := c.httpClient.Do(attemptRequest)
		if err != nil {
			return fmt.Errorf("calling meta graph api: %w", err)
		}
		if isRetryableStatus(response.StatusCode) && attempt < maxAttempts-1 {
			delay := retryDelay(response.Header.Get("Retry-After"), attempt)
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
			if closeErr := response.Body.Close(); closeErr != nil {
				return fmt.Errorf("closing meta graph retry response: %w", closeErr)
			}
			if err := waitForRetry(req.Context(), delay); err != nil {
				return err
			}

			continue
		}

		return decodeResponse(response, target)
	}

	return errors.New("meta graph api retry limit exceeded")
}

func requestForAttempt(request *http.Request, attempt int) (*http.Request, error) {
	if attempt == 0 || request.Body == nil {
		return request, nil
	}
	if request.GetBody == nil {
		return nil, errors.New("meta graph request body cannot be retried")
	}
	body, err := request.GetBody()
	if err != nil {
		return nil, fmt.Errorf("recreating meta graph request body: %w", err)
	}
	cloned := request.Clone(request.Context())
	cloned.Body = body

	return cloned, nil
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusInternalServerError ||
		status == http.StatusBadGateway || status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	const maximumDelay = 30 * time.Second

	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds >= 0 {
		return min(time.Duration(seconds)*time.Second, maximumDelay)
	}
	if retryAt, err := http.ParseTime(retryAfter); err == nil {
		return min(max(time.Until(retryAt), 0), maximumDelay)
	}

	return min(500*time.Millisecond*time.Duration(1<<attempt), maximumDelay)
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("waiting to retry meta graph api: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func decodeResponse(response *http.Response, target any) (resultErr error) {
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
