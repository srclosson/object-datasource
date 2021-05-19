package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/srclosson/object-datasource/pkg/models"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &ObjectDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// SampleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type ObjectDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im       instancemgmt.InstanceManager
	settings *models.Settings
}

func (d *ObjectDatasource) getInstance(ctx context.Context, pluginCtx backend.PluginContext) (*instanceSettings, error) {
	s, err := d.im.Get(pluginCtx)
	instance := s.(*instanceSettings)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *ObjectDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()
	instance, err := d.getInstance(ctx, req.PluginContext)
	if err != nil {
		return nil, err
	}

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, instance, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (td *ObjectDatasource) query(ctx context.Context, instance *instanceSettings, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var q models.ObjectDataQuery

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &q)
	if response.Error != nil {
		return response
	}

	backend.Logger.Debug("!!! Got a query !!!", "query", q.Config.Query, "instance", instance.settings)
	// create data frame response
	frame := data.NewFrame("response")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/ds/query?orgId=1", instance.settings.URL), bytes.NewBuffer(q.Config.Query))
	if err != nil {
		response.Error = err
		return response
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grafana-Org-Id", "1")
	req.Header.Set("Authorization", "Bearer eyJrIjoiVTJtQXZKd3c4ajZ4MjFUU08zbWVVMFFPaG9jZ1RXamUiLCJuIjoiTXlLZXkiLCJpZCI6MX0=")
	resp, err := instance.httpClient.Do(req)
	if err != nil {
		response.Error = err
		return response
	}

	defer resp.Body.Close() //nolint
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.Error = err
		return response
	}
	backend.Logger.Debug(">>>>>>>>>>> Query Body Response <<<<<<<<<<<<<<<", "body", body)

	// add the time dimension
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
	)

	// add values
	frame.Fields = append(frame.Fields,
		data.NewField("values", nil, []int64{10, 20}),
	)

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *ObjectDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	instance, err := d.getInstance(ctx, req.PluginContext)
	if err != nil {
		return nil, err
	}

	settings, err := models.LoadSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}

	instance.settings = settings

	backend.Logger.Debug(">>>>>>>>>>>>>>>>> We got a health check request", "req", req.PluginContext.DataSourceInstanceSettings, "settings", settings)

	queries, err := json.Marshal(settings.Request)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}
	backend.Logger.Debug(">>>>>>>> Request going out >>>>>>>>>>", "request", string(queries))
	hreq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/ds/query", instance.settings.URL), bytes.NewBuffer(queries))
	if err != nil {
		return nil, err
	}

	hreq.Header.Set("Accept", "application/json")
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Authorization", "Bearer eyJrIjoiVTJtQXZKd3c4ajZ4MjFUU08zbWVVMFFPaG9jZ1RXamUiLCJuIjoiTXlLZXkiLCJpZCI6MX0=")
	resp, err := instance.httpClient.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint
	data, err := ioutil.ReadAll(resp.Body)
	backend.Logger.Debug(">>>>>>>>>>> Query Body Response <<<<<<<<<<<<<<<", "data", string(data), "err", err)

	if err != nil && resp.StatusCode != 200 {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

type instanceSettings struct {
	httpClient *http.Client
	settings   *models.Settings
}

func newDataSourceInstance(rawSettings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &instanceSettings{
		httpClient: &http.Client{},
		settings:   nil,
	}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
