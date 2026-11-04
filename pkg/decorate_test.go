package main

import (
	"math"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const decoratedLinks = `{
	"folders": {
		"plant": {
			"description": "The plant",
			"labels": {"site": "plant1", "host": "static-loses"},
			"links": [{"title": "Site runbook", "url": "https://example.com/runbook"}],
			"labelLookups": [{
				"uid": "assets",
				"query": {"table": {"host": ["a", "b"], "maker": ["Acme", "Globex"], "room": ["1", "2"]}},
				"matchLabel": "host"
			}]
		}
	},
	"queryLinks": [{
		"id": "l1", "name": "plant.temp", "uid": "prom",
		"query": {"series": [{"host": "a"}, {"host": "b"}, {"host": "c"}], "datasource": {"uid": "prom", "type": "prometheus"}},
		"metadata": {
			"unit": "celsius", "displayName": "Temperature ${__field.labels.host}", "min": 0, "max": 120,
			"thresholds": [{"value": 80, "color": "orange"}, {"value": 100, "color": "red"}]
		}
	}]
}`

func TestQueryDecoratesResults(t *testing.T) {
	g := newFakeGrafana(t)
	ds := newTestDatasource(t, g.URL, testToken, decoratedLinks)

	res := runQuery(t, ds, "A", `{"linkId": "l1"}`)
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	if len(res.Frames) != 3 {
		t.Fatalf("want 3 series, got %d", len(res.Frames))
	}

	byHost := map[string]map[string]string{}
	for _, f := range res.Frames {
		v := f.Fields[1]
		byHost[v.Labels["host"]] = v.Labels

		cfg := v.Config
		if cfg.Unit != "celsius" || cfg.Description != "The plant" || cfg.DisplayNameFromDS != "Temperature "+v.Labels["host"] {
			t.Errorf("field config not applied: %+v", cfg)
		}
		if cfg.Min == nil || *cfg.Min != 0 || cfg.Max == nil || *cfg.Max != 120 {
			t.Errorf("range not applied: min %v max %v", cfg.Min, cfg.Max)
		}
		steps := cfg.Thresholds.Steps
		if len(steps) != 3 || !math.IsInf(float64(steps[0].Value), -1) || steps[1].Value != 80 || steps[2].Color != "red" {
			t.Errorf("thresholds not applied: %+v", steps)
		}
		if len(cfg.Links) != 1 || cfg.Links[0].URL != "https://example.com/runbook" || !cfg.Links[0].TargetBlank {
			t.Errorf("links not applied: %+v", cfg.Links)
		}
		if f.Fields[0].Config != nil && f.Fields[0].Config.Unit != "" {
			t.Error("the time field was decorated")
		}

		custom := f.Meta.Custom.(map[string]any)[metaKey].(map[string]any)
		if custom["path"] != "plant.temp" {
			t.Errorf("tree path not in frame metadata: %v", custom)
		}
	}

	// The lookup joins on host; the series' own host label beats the static one
	if l := byHost["a"]; l["maker"] != "Acme" || l["room"] != "1" || l["site"] != "plant1" {
		t.Errorf("host a labels: %v", l)
	}
	if l := byHost["b"]; l["maker"] != "Globex" || l["room"] != "2" {
		t.Errorf("host b labels: %v", l)
	}
	if l := byHost["c"]; l["maker"] != "" || l["site"] != "plant1" {
		t.Errorf("host c has no row in the lookup, but got %v", l)
	}
}

func TestDisplayName(t *testing.T) {
	field := data.NewField("value", data.Labels{"host": "a", "site": "s1"}, []float64{1})
	for template, want := range map[string]string{
		"Temperature ${__field.labels.host}": "Temperature a",
		"${__field.name} ${__field.labels}":  `value {host="a", site="s1"}`,
		"${__field.labels.missing}x":         "x",
		"plain":                              "plain",
	} {
		if got := displayName(template, field); got != want {
			t.Errorf("%q: got %q, want %q", template, got, want)
		}
	}
}

func TestLabelLookupsAreCachedAndFailuresAreWarnings(t *testing.T) {
	g := newFakeGrafana(t)
	// With the query cache off, so the second query reaches the source and only the lookup is cached
	noQueryCache := strings.Replace(decoratedLinks, `"folders"`, `"cache": {"defaultTtl": -1}, "folders"`, 1)
	ds := newTestDatasource(t, g.URL, testToken, noQueryCache)

	runQuery(t, ds, "A", `{"linkId": "l1"}`)
	runQuery(t, ds, "A", `{"linkId": "l1"}`)
	if n := len(g.requests); n != 3 {
		t.Errorf("want 2 link queries and 1 lookup, got %d requests", n)
	}

	broken := strings.Replace(decoratedLinks, `"uid": "assets",`, `"uid": "",`, 1)
	ds = newTestDatasource(t, g.URL, testToken, broken)
	res := runQuery(t, ds, "A", `{"linkId": "l1"}`)
	if res.Error != nil {
		t.Fatalf("a failed lookup shouldn't fail the query: %v", res.Error)
	}
	notices := res.Frames[0].Meta.Notices
	if len(notices) != 1 || !strings.Contains(notices[0].Text, "label lookup 1 has no datasource") {
		t.Errorf("want a warning about the lookup, got %+v", notices)
	}
}
