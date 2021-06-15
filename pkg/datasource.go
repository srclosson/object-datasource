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

// ObjectDatasource backend
// new datasource plugins with an backend.
type ObjectDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager

	settings *models.Settings
}

func (d *ObjectDatasource) getInstance(ctx context.Context, pluginCtx backend.PluginContext) (*instanceSettings, error) {
	backend.Logger.Debug("New datasource instance comming right up!!!", "pluginCtx", pluginCtx.DataSourceInstanceSettings)

	s, err := d.im.Get(pluginCtx)
	if err != nil {
		backend.Logger.Debug("Apparently we have an error here?", "error", err)
		return nil, err
	}
	instance := s.(*instanceSettings)

	backend.Logger.Debug("Appaarently we have some settings here?", "settings", s)

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
	if err != nil || instance == nil {
		return nil, err
	}

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		//res := d.query(ctx, instance, q)

		var objectQuery models.ObjectDataQuery
		err := json.Unmarshal(q.JSON, &objectQuery)
		backend.Logger.Debug("objectQuery", "query", objectQuery)
		// save the response in a hashmap
		// based on with RefID as identifier
		frames := make([]*data.Frame, 0)
		for _, frameIn := range objectQuery.Response.Data {
			fields := make([]*data.Field, 0)
			for _, fieldIn := range frameIn.Fields {
				backend.Logger.Debug("The labels are", "labels", fieldIn.Labels, "type", fieldIn.TypeInfo.Frame)
				var field *data.Field
				switch fieldIn.TypeInfo.Frame {
				case "time.Time":
					field = data.NewField(fieldIn.Name, fieldIn.Labels, make([]time.Time, 0))
				default:
					field = data.NewField(fieldIn.Name, fieldIn.Labels, make([]float64, 0))
				}

				//field.Config = fieldIn.State
				for _, value := range fieldIn.Values {
					switch fieldIn.TypeInfo.Frame {
					case "time.Time":
						v := int64(value.(float64))
						backend.Logger.Debug("timestamp", "orig", v, "secs", v/1000, "msecs", v%1000)
						field.Append(time.Unix(v/1000, (v%1000)*1000000))
					default:
						field.Append(value.(float64))
					}
				}
				fields = append(fields, field)
			}
			frames = append(frames, &data.Frame{
				Name:   frameIn.Name,
				Fields: fields,
			})
		}
		response.Responses[q.RefID] = backend.DataResponse{
			Frames: frames,
			Error:  err,
		}
	}
	// proxiedRequestJSON, err := json.Marshal(*req)
	// if err != nil {
	// 	return nil, err
	// }

	// backend.Logger.Debug("!!! Got a query !!!", "proxiedRequestJSON", string(proxiedRequestJSON))

	// proxiedRequest, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/ds/query", instance.settings.URL), bytes.NewBuffer([]byte(proxiedRequestJSON)))
	// if err != nil {
	// 	return nil, err
	// }

	// proxiedRequest.Header.Set("Accept", "application/json")
	// proxiedRequest.Header.Set("Content-Type", "application/json")
	// proxiedRequest.Header.Set("Authorization", "Bearer eyJrIjoiNE1KZXB1bGVWRDRpdU9lVWNITVNZTFpXSzExQmppZG0iLCJuIjoiT2JqZWN0RGF0YXNvdXJjZSIsImlkIjoxfQ==")
	// proxiedResponse, err := instance.httpClient.Do(proxiedRequest)
	// if err != nil {
	// 	return nil, err
	// }

	// defer proxiedResponse.Body.Close() //nolint
	// proxiedResponseJSON, err := ioutil.ReadAll(proxiedResponse.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// var response backend.QueryDataResponse
	// err = json.Unmarshal(proxiedResponseJSON, &response)

	// backend.Logger.Debug("proxied response", "json", len(response.Responses))

	// if err != nil {
	// 	return nil, err
	// }

	return response, nil
}

func (td *ObjectDatasource) query(ctx context.Context, instance *instanceSettings, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var innerQuery models.ObjectDataQuery

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &innerQuery)

	if response.Error != nil {
		return response
	}

	innerQuery.Config.Query["intervalMs"] = query.Interval.Milliseconds()
	innerQuery.Config.Query["maxDataPoints"] = query.MaxDataPoints
	innerQuery.Config.Query["refId"] = innerQuery.Name
	innerQuery.Config.Query["limit"] = query.MaxDataPoints

	rawInnerQuery, err := json.Marshal(innerQuery.Config.Query)

	proxiedRequest := models.ProxiedDataRequest{
		From:    fmt.Sprintf("%d", query.TimeRange.From.UnixNano()/int64(time.Millisecond)),
		To:      fmt.Sprintf("%d", query.TimeRange.To.UnixNano()/int64(time.Millisecond)),
		Queries: []json.RawMessage{rawInnerQuery},
	}

	proxiedRequestJSON, err := json.Marshal(proxiedRequest)
	if err != nil {
		response.Error = err
		return response
	}

	backend.Logger.Debug("!!! Got a query !!!", "proxiedRequestJSON", string(proxiedRequestJSON))

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/ds/query", instance.settings.URL), bytes.NewBuffer([]byte(proxiedRequestJSON)))
	if err != nil {
		response.Error = err
		return response
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJrIjoiNE1KZXB1bGVWRDRpdU9lVWNITVNZTFpXSzExQmppZG0iLCJuIjoiT2JqZWN0RGF0YXNvdXJjZSIsImlkIjoxfQ==")
	resp, err := instance.httpClient.Do(req)
	if err != nil {
		response.Error = err
		return response
	}

	defer resp.Body.Close() //nolint
	proxiedResponseJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.Error = err
		return response
	}

	var proxiedResponse models.ProxiedResponse
	err = json.Unmarshal(proxiedResponseJSON, &proxiedResponse)

	backend.Logger.Debug("proxied response", "json", len(proxiedResponse.Responses))

	if err != nil {
		response.Error = err
		return response
	}

	for _, dataResponse := range proxiedResponse.Responses {
		response.Frames = append(response.Frames, dataResponse.Frames...)
	}

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
	hreq.Header.Set("Authorization", "Bearer eyJrIjoiNE1KZXB1bGVWRDRpdU9lVWNITVNZTFpXSzExQmppZG0iLCJuIjoiT2JqZWN0RGF0YXNvdXJjZSIsImlkIjoxfQ==")
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
	settings, err := models.LoadSettings(rawSettings)
	return &instanceSettings{
		httpClient: &http.Client{},
		settings:   settings,
	}, err
}

func (s *instanceSettings) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
