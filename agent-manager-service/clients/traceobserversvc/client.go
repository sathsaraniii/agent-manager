package traceobserversvc

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	occlient "github.com/wso2/agent-manager/agent-manager-service/clients/openchoreosvc/client"
	"github.com/wso2/agent-manager/agent-manager-service/clients/requests"
)

// TraceObserverClient defines the interface for trace observer operations.
type TraceObserverClient interface {
	ListTraces(ctx context.Context, params TraceListParams) (map[string]any, error)
	ExportTraces(ctx context.Context, params TraceListParams) (map[string]any, error)
	GetTrace(ctx context.Context, params TraceDetailsParams) (map[string]any, error)
}

// Config contains configuration for the trace observer client.
type Config struct {
	BaseURL      string
	AuthProvider occlient.AuthProvider
	RetryConfig  requests.RequestRetryConfig
}

type traceObserverClient struct {
	baseURL      string
	httpClient   requests.HttpClient
	authProvider occlient.AuthProvider
}

// NewTraceObserverClient creates a new trace observer client.
func NewTraceObserverClient(cfg *Config) (TraceObserverClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if cfg.AuthProvider == nil {
		return nil, fmt.Errorf("auth provider is required")
	}

	retryConfig := cfg.RetryConfig
	httpClient := requests.NewRetryableHTTPClient(&http.Client{}, retryConfig)

	return &traceObserverClient{
		baseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		httpClient:   httpClient,
		authProvider: cfg.AuthProvider,
	}, nil
}

func (c *traceObserverClient) ListTraces(ctx context.Context, params TraceListParams) (map[string]any, error) {
	query := map[string]string{
		"organization": params.Organization,
		"project":      params.Project,
		"agent":        params.Component,
		"environment":  params.Environment,
		"startTime":    params.StartTime,
		"endTime":      params.EndTime,
	}
	if params.Limit > 0 {
		query["limit"] = strconv.Itoa(params.Limit)
	}
	if params.Offset >= 0 {
		query["offset"] = strconv.Itoa(params.Offset)
	}
	if strings.TrimSpace(params.SortOrder) != "" {
		query["sortOrder"] = params.SortOrder
	}

	return c.doGetMap(ctx, "traceobserversvc.ListTraces", "/api/v1/traces", query)
}

func (c *traceObserverClient) ExportTraces(ctx context.Context, params TraceListParams) (map[string]any, error) {
	query := map[string]string{
		"organization": params.Organization,
		"project":      params.Project,
		"agent":        params.Component,
		"environment":  params.Environment,
		"startTime":    params.StartTime,
		"endTime":      params.EndTime,
	}
	if params.Limit > 0 {
		query["limit"] = strconv.Itoa(params.Limit)
	}
	if params.Offset >= 0 {
		query["offset"] = strconv.Itoa(params.Offset)
	}
	if strings.TrimSpace(params.SortOrder) != "" {
		query["sortOrder"] = params.SortOrder
	}

	return c.doGetMap(ctx, "traceobserversvc.ExportTraces", "/api/v1/traces/export", query)
}

func (c *traceObserverClient) GetTrace(ctx context.Context, params TraceDetailsParams) (map[string]any, error) {
	query := map[string]string{
		"organization": params.Organization,
		"startTime":    params.StartTime,
		"endTime":      params.EndTime,
	}
	if strings.TrimSpace(params.Project) != "" {
		query["project"] = params.Project
	}
	if strings.TrimSpace(params.Component) != "" {
		query["agent"] = params.Component
	}
	if strings.TrimSpace(params.Environment) != "" {
		query["environment"] = params.Environment
	}
	if strings.TrimSpace(params.SortOrder) != "" {
		query["sortOrder"] = params.SortOrder
	}
	if params.Limit > 0 {
		query["limit"] = strconv.Itoa(params.Limit)
	}

	path := "/api/v1/traces/" + params.TraceID + "/spans"
	return c.doGetMap(ctx, "traceobserversvc.GetTrace", path, query)
}

func (c *traceObserverClient) doGetMap(ctx context.Context, name, path string, query map[string]string) (map[string]any, error) {
	if c == nil {
		return nil, fmt.Errorf("trace observer client is nil")
	}
	url := c.baseURL + path

	result, err := c.sendGet(ctx, name, url, query)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	scanErr := result.ScanResponse(&out, http.StatusOK)
	if scanErr == nil {
		return out, nil
	}
	if httpErr, ok := scanErr.(*requests.HttpError); ok && httpErr.StatusCode == http.StatusUnauthorized {
		// retry once after invalidating token
		c.authProvider.InvalidateToken()
		result, retryErr := c.sendGet(ctx, name, url, query)
		if retryErr != nil {
			return nil, retryErr
		}
		var retryOut map[string]any
		if retryErr := result.ScanResponse(&retryOut, http.StatusOK); retryErr != nil {
			return nil, retryErr
		}
		return retryOut, nil
	}

	return nil, scanErr
}

func (c *traceObserverClient) sendGet(ctx context.Context, name, url string, query map[string]string) (*requests.Result, error) {
	req := &requests.HttpRequest{
		Name:   name,
		URL:    url,
		Method: http.MethodGet,
		Query:  query,
	}

	token, err := c.authProvider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get auth token: %w", name, err)
	}
	if strings.TrimSpace(token) != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}
	req.SetHeader("Content-Type", "application/json")

	result := requests.SendRequest(ctx, c.httpClient, req)
	if result == nil {
		return nil, fmt.Errorf("%s: request returned nil result", name)
	}
	return result, nil
}
