package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/srclosson/object-datasource/pkg/models"
)

func TestInterpolate(t *testing.T) {
	vars := map[string]models.VarValue{
		"line": {"1"},
		"host": {"a.example", "b.example"},
	}
	for name, tc := range map[string]struct {
		dsType, in, want string
	}{
		"single value":             {"prometheus", `temp{line="$line"}`, `temp{line="1"}`},
		"braces":                   {"prometheus", `temp{line="${line}"}`, `temp{line="1"}`},
		"brackets":                 {"prometheus", `temp{line="[[line]]"}`, `temp{line="1"}`},
		"multi value as regex":     {"prometheus", `up{host=~"$host"}`, `up{host=~"(a\.example|b\.example)"}`},
		"multi value as glob":      {"influxdb", `host = $host`, `host = {a.example,b.example}`},
		"explicit format":          {"prometheus", `${host:csv}`, `a.example,b.example`},
		"quoted":                   {"mysql", `host IN (${host:singlequote})`, `host IN ('a.example','b.example')`},
		"unknown variables kept":   {"prometheus", `rate(x[$__rate_interval]) by ($missing)`, `rate(x[$__rate_interval]) by ($missing)`},
		"name is a whole word":     {"prometheus", `$lines`, `$lines`},
		"several in one string":    {"prometheus", `$line-$line`, `1-1`},
		"json format, single item": {"prometheus", `${line:json}`, `["1"]`},
	} {
		t.Run(name, func(t *testing.T) {
			model := map[string]any{"expr": tc.in}
			interpolate(model, vars, tc.dsType)
			if model["expr"] != tc.want {
				t.Errorf("got %q, want %q", model["expr"], tc.want)
			}
		})
	}
}

func TestInterpolateWalksTheWholeModelButNotItsIdentity(t *testing.T) {
	model := map[string]any{
		"refId":      "$line",
		"datasource": map[string]any{"uid": "$line"},
		"targets":    []any{map[string]any{"filter": "line=$line"}, 7.0},
	}
	interpolate(model, map[string]models.VarValue{"line": {"1"}}, "")
	want := map[string]any{
		"refId":      "$line",
		"datasource": map[string]any{"uid": "$line"},
		"targets":    []any{map[string]any{"filter": "line=1"}, 7.0},
	}
	if !reflect.DeepEqual(model, want) {
		t.Errorf("got %v, want %v", model, want)
	}
}

func TestQueryAppliesVariables(t *testing.T) {
	g := newFakeGrafana(t)
	ds := newTestDatasource(t, g.URL, testToken, `{"queryLinks": [
		{"id": "l1", "name": "a", "query": {"expr": "up{job=~\"$job\"}", "datasource": {"uid": "prom", "type": "prometheus"}}}
	]}`)

	res := runQuery(t, ds, "A", `{"linkId": "l1", "variables": {"job": ["api", "web"]}}`)
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	expr := g.requests[0]["queries"].([]any)[0].(map[string]any)["expr"].(string)
	if !strings.Contains(expr, `job=~"(api|web)"`) {
		t.Errorf("variables not applied: %s", expr)
	}
}
