package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/srclosson/object-datasource/pkg/models"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource resolves queries that reference query links, and runs each link's query against
// its target datasource through Grafana's API
type Datasource struct {
	settings *models.Settings
	grafana  grafanaAPI
}

// NewDatasource creates a datasource instance. Grafana creates a new one whenever the
// datasource's settings change.
func NewDatasource(_ context.Context, s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := models.LoadSettings(s)
	if err != nil {
		return nil, err
	}
	return &Datasource{
		settings: settings,
		grafana:  grafanaAPI{client: newHTTPClient(), settings: settings},
	}, nil
}

// Dispose is called before a new instance replaces this one
func (d *Datasource) Dispose() {}

// QueryData runs each query's link concurrently and returns the responses by RefID
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, q := range req.Queries {
		wg.Add(1)
		go func(q backend.DataQuery) {
			defer wg.Done()
			res := d.query(ctx, req.PluginContext.OrgID, q)
			mu.Lock()
			response.Responses[q.RefID] = res
			mu.Unlock()
		}(q)
	}
	wg.Wait()
	return response, nil
}

func (d *Datasource) query(ctx context.Context, orgID int64, q backend.DataQuery) backend.DataResponse {
	var oq models.ObjectQuery
	if err := json.Unmarshal(q.JSON, &oq); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("can't read the query: %v", err))
	}
	if oq.LinkID == "" && oq.Name == "" {
		return backend.ErrDataResponse(backend.StatusBadRequest, "no query link selected")
	}
	link, ok := d.settings.FindLink(oq.LinkID, oq.Name)
	if !ok {
		return backend.ErrDataResponse(backend.StatusBadRequest,
			fmt.Sprintf("query link %q no longer exists in the datasource settings", oq.Name))
	}
	model, err := innerQuery(link, q)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}
	interpolate(model, oq.Variables, link.Datasource().Type)
	res, err := d.grafana.query(ctx, orgID, model, q.TimeRange)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadGateway, err.Error())
	}
	for _, f := range res.Frames {
		f.RefID = q.RefID
	}
	return res
}

// innerQuery is the link's query as the target datasource expects it, with the fields
// Grafana's query API owns taken from the outer query
func innerQuery(link models.QueryLink, q backend.DataQuery) (map[string]any, error) {
	model := map[string]any{}
	if len(link.Query) > 0 {
		if err := json.Unmarshal(link.Query, &model); err != nil {
			return nil, fmt.Errorf("can't read query link %q: %w", link.Name, err)
		}
	}
	ds := link.Datasource()
	if ds.UID == "" {
		return nil, fmt.Errorf("query link %q has no datasource", link.Name)
	}
	model["datasource"] = ds
	model["refId"] = q.RefID
	model["intervalMs"] = q.Interval.Milliseconds()
	model["maxDataPoints"] = q.MaxDataPoints
	return model, nil
}

// CheckHealth checks that the plugin can authenticate to Grafana's API and read the
// datasource each query link targets
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	orgID := req.PluginContext.OrgID
	fail := func(msg string) (*backend.CheckHealthResult, error) {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: msg}, nil
	}

	status, raw, err := d.grafana.do(ctx, orgID, http.MethodGet, "/api/datasources", nil)
	if err != nil {
		return fail(err.Error())
	}
	if status != http.StatusOK {
		return fail(apiError(status, raw).Error())
	}

	for _, link := range d.settings.QueryLinks {
		ds := link.Datasource()
		if ds.UID == "" {
			return fail(fmt.Sprintf("query link %q has no datasource", link.Name))
		}
		status, raw, err := d.grafana.do(ctx, orgID, http.MethodGet, "/api/datasources/uid/"+url.PathEscape(ds.UID), nil)
		if err != nil {
			return fail(err.Error())
		}
		if status == http.StatusNotFound {
			return fail(fmt.Sprintf("query link %q targets a datasource that doesn't exist (uid %s)", link.Name, ds.UID))
		}
		if status != http.StatusOK {
			return fail(fmt.Sprintf("query link %q: %v", link.Name, apiError(status, raw)))
		}
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Connected to Grafana's API; %d query link(s) can reach their datasources", len(d.settings.QueryLinks)),
	}, nil
}
