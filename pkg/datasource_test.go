package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const testToken = "test-token"

// fakeGrafana stands in for Grafana's API. It answers /api/ds/query with one frame per query,
// and /api/datasources/uid/<uid> for the uids in known.
type fakeGrafana struct {
	*httptest.Server
	status   int           // when set, every request gets this status
	delay    time.Duration // how long /api/ds/query takes
	known    map[string]bool
	mu       sync.Mutex
	requests []map[string]any
	headers  []http.Header
}

// queryCount is how many queries reached /api/ds/query
func (f *fakeGrafana) queryCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

func newFakeGrafana(t *testing.T) *fakeGrafana {
	f := &fakeGrafana{known: map[string]bool{"prom": true}}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.headers = append(f.headers, r.Header.Clone())
		f.mu.Unlock()
		if f.status != 0 {
			w.WriteHeader(f.status)
			_, _ = w.Write([]byte(`{"message":"nope"}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+testToken {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"invalid API key"}`))
			return
		}
		switch {
		case r.URL.Path == "/api/ds/query":
			raw, _ := io.ReadAll(r.Body)
			var body map[string]any
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Errorf("bad request body: %v", err)
			}
			f.mu.Lock()
			f.requests = append(f.requests, body)
			f.mu.Unlock()
			time.Sleep(f.delay)
			res := backend.NewQueryDataResponse()
			for _, q := range body["queries"].([]any) {
				refID := q.(map[string]any)["refId"].(string)
				// A query with a "table" answers with that table, as a SQL datasource would;
				// one with "series" answers with a series per label set, as Prometheus would
				if table, ok := q.(map[string]any)["table"].(map[string]any); ok {
					f := data.NewFrame("table")
					for _, col := range sortedKeys(table) {
						var vals []string
						for _, v := range table[col].([]any) {
							vals = append(vals, v.(string))
						}
						f.Fields = append(f.Fields, data.NewField(col, nil, vals))
					}
					res.Responses[refID] = backend.DataResponse{Frames: data.Frames{f}}
					continue
				}
				var frames data.Frames
				series, _ := q.(map[string]any)["series"].([]any)
				if series == nil {
					series = []any{map[string]any{}}
				}
				for _, l := range series {
					labels := data.Labels{}
					for k, v := range l.(map[string]any) {
						labels[k] = v.(string)
					}
					frames = append(frames, data.NewFrame("up",
						data.NewField("time", nil, []time.Time{time.UnixMilli(1000)}),
						data.NewField("value", labels, []float64{1})))
				}
				res.Responses[refID] = backend.DataResponse{Frames: frames}
			}
			_ = json.NewEncoder(w).Encode(res)
		case r.URL.Path == "/api/datasources":
			_, _ = w.Write([]byte(`[]`))
		case strings.HasPrefix(r.URL.Path, "/api/datasources/uid/"):
			if !f.known[strings.TrimPrefix(r.URL.Path, "/api/datasources/uid/")] {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"Data source not found"}`))
				return
			}
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

const testLinks = `{"queryLinks": [
	{"id": "l1", "name": "plant.line1.temp", "uid": "prom",
	 "query": {"refId": "x", "expr": "temp{line=\"1\"}", "datasource": {"uid": "prom", "type": "prometheus"}}},
	{"name": "plant.line2.temp", "uid": "prom", "query": {"expr": "temp{line=\"2\"}"}},
	{"id": "gone", "name": "plant.gone", "uid": "missing", "query": {}}
]}`

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func newTestDatasource(t *testing.T, url, token, jsonData string) *Datasource {
	t.Helper()
	inst, err := NewDatasource(context.Background(), backend.DataSourceInstanceSettings{
		URL:                     url,
		JSONData:                []byte(jsonData),
		DecryptedSecureJSONData: map[string]string{"serviceAccountToken": token},
	})
	if err != nil {
		t.Fatal(err)
	}
	return inst.(*Datasource)
}

func runQuery(t *testing.T, ds *Datasource, refID, queryJSON string) backend.DataResponse {
	t.Helper()
	res, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		PluginContext: backend.PluginContext{OrgID: 2},
		Queries: []backend.DataQuery{{
			RefID:         refID,
			JSON:          []byte(queryJSON),
			Interval:      30 * time.Second,
			MaxDataPoints: 500,
			TimeRange:     backend.TimeRange{From: time.UnixMilli(1000), To: time.UnixMilli(2000)},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return res.Responses[refID]
}

func TestQueryForwardsTheLinkQuery(t *testing.T) {
	g := newFakeGrafana(t)
	ds := newTestDatasource(t, g.URL+"/", testToken, testLinks)

	res := runQuery(t, ds, "A", `{"linkId": "l1", "name": "plant.line1.temp"}`)
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if len(res.Frames) != 1 || res.Frames[0].RefID != "A" {
		t.Fatalf("want one frame with refId A, got %+v", res.Frames)
	}

	if len(g.requests) != 1 {
		t.Fatalf("want 1 request to /api/ds/query, got %d", len(g.requests))
	}
	body := g.requests[0]
	if body["from"] != "1000" || body["to"] != "2000" {
		t.Errorf("time range not forwarded: from=%v to=%v", body["from"], body["to"])
	}
	q := body["queries"].([]any)[0].(map[string]any)
	if q["expr"] != `temp{line="1"}` || q["refId"] != "A" {
		t.Errorf("link query not forwarded with the outer refId: %v", q)
	}
	if q["intervalMs"] != float64(30000) || q["maxDataPoints"] != float64(500) {
		t.Errorf("interval and max data points not forwarded: %v", q)
	}
	if ref := q["datasource"].(map[string]any); ref["uid"] != "prom" || ref["type"] != "prometheus" {
		t.Errorf("wrong target datasource: %v", ref)
	}
	if got := g.headers[0].Get("X-Grafana-Org-Id"); got != "2" {
		t.Errorf("org header = %q, want 2", got)
	}
}

func TestQueryFindsLinksWithoutIDsByName(t *testing.T) {
	g := newFakeGrafana(t)
	ds := newTestDatasource(t, g.URL, testToken, testLinks)

	res := runQuery(t, ds, "B", `{"name": "plant.line2.temp"}`)
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	q := g.requests[0]["queries"].([]any)[0].(map[string]any)
	if q["expr"] != `temp{line="2"}` || q["datasource"].(map[string]any)["uid"] != "prom" {
		t.Errorf("wrong query forwarded: %v", q)
	}
}

func TestQueryErrors(t *testing.T) {
	g := newFakeGrafana(t)
	for name, tc := range map[string]struct {
		token, query, want string
	}{
		"deleted link": {testToken, `{"linkId": "nope", "name": "plant.old"}`, `"plant.old" no longer exists`},
		"no link":      {testToken, `{}`, "no query link selected"},
		"bad token":    {"wrong", `{"linkId": "l1"}`, "rejected the service account token"},
		"no token":     {"", `{"linkId": "l1"}`, "no Grafana service account"},
	} {
		t.Run(name, func(t *testing.T) {
			ds := newTestDatasource(t, g.URL, tc.token, testLinks)
			res := runQuery(t, ds, "A", tc.query)
			if res.Error == nil || !strings.Contains(res.Error.Error(), tc.want) {
				t.Errorf("error = %v, want it to contain %q", res.Error, tc.want)
			}
		})
	}
}

func TestCheckHealth(t *testing.T) {
	g := newFakeGrafana(t)
	check := func(token, links string) *backend.CheckHealthResult {
		ds := newTestDatasource(t, g.URL, token, links)
		res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	if res := check(testToken, `{"queryLinks": [{"name": "a", "uid": "prom"}]}`); res.Status != backend.HealthStatusOk {
		t.Errorf("want OK, got %s", res.Message)
	}
	if res := check(testToken, testLinks); res.Status != backend.HealthStatusError || !strings.Contains(res.Message, `"plant.gone"`) {
		t.Errorf("want an error naming the broken link, got %s", res.Message)
	}
	if res := check("wrong", `{}`); res.Status != backend.HealthStatusError || !strings.Contains(res.Message, "401") {
		t.Errorf("want a 401 error, got %s", res.Message)
	}
	if res := check("", `{}`); res.Status != backend.HealthStatusError {
		t.Errorf("want an error without a token, got %s", res.Message)
	}
}
