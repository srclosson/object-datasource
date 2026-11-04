package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/srclosson/object-datasource/pkg/models"
)

// errNoCredentials means there is no way to authenticate to Grafana's API
var errNoCredentials = errors.New("the object datasource has no Grafana service account to query other datasources with. " +
	"Paste a service account token into the datasource settings, or enable the externalServiceAccounts feature toggle " +
	"and [auth] managed_service_accounts_enabled, then restart Grafana")

// grafanaAPI runs queries against other datasources through Grafana's datasource query API
// (POST /api/ds/query), so a link can target any datasource Grafana can query.
type grafanaAPI struct {
	client   *http.Client
	settings *models.Settings
}

// target is Grafana's API base URL and the token to call it with. A token pasted into the
// settings wins over the managed service account Grafana gives the plugin.
func (g grafanaAPI) target(ctx context.Context) (string, string, error) {
	cfg := backend.GrafanaConfigFromContext(ctx)
	base := g.settings.GrafanaURL
	if base == "" {
		appURL, err := cfg.AppURL()
		if err != nil {
			return "", "", fmt.Errorf("can't find Grafana's URL (%w); set it in the datasource settings", err)
		}
		base = strings.TrimRight(appURL, "/")
	}
	token := g.settings.ServiceAccountToken
	if token == "" {
		token, _ = cfg.PluginAppClientSecret()
	}
	if token == "" {
		return "", "", errNoCredentials
	}
	return base, token, nil
}

// do sends a request to Grafana's API and returns the status and body
func (g grafanaAPI) do(ctx context.Context, orgID int64, method, path string, body []byte) (int, []byte, error) {
	base, token, err := g.target(ctx)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if orgID > 0 {
		req.Header.Set("X-Grafana-Org-Id", strconv.FormatInt(orgID, 10))
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("can't reach Grafana's API at %s: %w", base, err)
	}
	defer resp.Body.Close() //nolint:errcheck
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 256<<20))
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("reading Grafana's answer: %w", err)
	}
	return resp.StatusCode, raw, nil
}

// query runs one inner query over the time range and returns its response
func (g grafanaAPI) query(ctx context.Context, orgID int64, model map[string]any, tr backend.TimeRange) (backend.DataResponse, error) {
	body, err := json.Marshal(map[string]any{
		"queries": []any{model},
		"from":    strconv.FormatInt(tr.From.UnixMilli(), 10),
		"to":      strconv.FormatInt(tr.To.UnixMilli(), 10),
	})
	if err != nil {
		return backend.DataResponse{}, err
	}
	status, raw, err := g.do(ctx, orgID, http.MethodPost, "/api/ds/query", body)
	if err != nil {
		return backend.DataResponse{}, err
	}
	var qdr backend.QueryDataResponse
	if err := json.Unmarshal(raw, &qdr); err != nil || qdr.Responses == nil {
		return backend.DataResponse{}, apiError(status, raw)
	}
	refID, _ := model["refId"].(string)
	r, ok := qdr.Responses[refID]
	if !ok {
		return backend.DataResponse{}, fmt.Errorf("the target datasource returned nothing for %s", refID)
	}
	return r, nil
}

// apiError turns an unsuccessful API response into an error, using Grafana's message when there is one
func apiError(status int, raw []byte) error {
	var msg struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &msg)
	if msg.Message == "" {
		msg.Message = strings.TrimSpace(string(raw[:min(len(raw), 200)]))
	}
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("Grafana rejected the service account token (HTTP 401): %s", msg.Message)
	case http.StatusForbidden:
		return fmt.Errorf("the service account isn't allowed to do this (HTTP 403): %s", msg.Message)
	}
	return fmt.Errorf("Grafana's API failed (HTTP %d): %s", status, msg.Message)
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 2 * time.Minute}
}
