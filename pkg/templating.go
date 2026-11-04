package main

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/srclosson/object-datasource/pkg/models"
)

// $name, ${name}, ${name:format}, [[name]] and [[name:format]], as Grafana writes them
var variablePattern = regexp.MustCompile(`\$(\w+)|\$\{(\w+)(?::(\w+))?\}|\[\[(\w+)(?::(\w+))?\]\]`)

// regexDatasources match multi-value variables with a regex, as their own templating does
var regexDatasources = map[string]bool{"prometheus": true, "loki": true}

// interpolate replaces the variables in every string of a query model. Variables it has no
// value for, such as Grafana's $__interval, are left for the target datasource.
func interpolate(model map[string]any, vars map[string]models.VarValue, dsType string) {
	if len(vars) == 0 {
		return
	}
	defaultFormat := "glob"
	if regexDatasources[dsType] {
		defaultFormat = "regex"
	}
	var walk func(v any) any
	walk = func(v any) any {
		switch t := v.(type) {
		case string:
			return variablePattern.ReplaceAllStringFunc(t, func(match string) string {
				m := variablePattern.FindStringSubmatch(match)
				name, format := m[1]+m[2]+m[4], m[3]+m[5]
				value, ok := vars[name]
				if !ok {
					return match
				}
				if format == "" {
					format = defaultFormat
				}
				return formatValue(value, format)
			})
		case map[string]any:
			for k, e := range t {
				t[k] = walk(e)
			}
		case []any:
			for i, e := range t {
				t[i] = walk(e)
			}
		}
		return v
	}
	for k, v := range model {
		// The target datasource and the fields Grafana's API owns aren't templated
		if k == "datasource" || k == "refId" {
			continue
		}
		model[k] = walk(v)
	}
}

// formatValue formats a value like Grafana's variable formats of the same names
func formatValue(v models.VarValue, format string) string {
	switch format {
	case "regex", "glob", "pipe", "csv", "raw":
		// A single value is used as is
		if len(v) == 1 {
			return v[0]
		}
	}
	switch format {
	case "regex":
		escaped := make([]string, len(v))
		for i, s := range v {
			escaped[i] = regexp.QuoteMeta(s)
		}
		return "(" + strings.Join(escaped, "|") + ")"
	case "pipe":
		return strings.Join(v, "|")
	case "csv", "raw":
		return strings.Join(v, ",")
	case "json":
		b, _ := json.Marshal([]string(v))
		return string(b)
	case "doublequote":
		return `"` + strings.Join(v, `","`) + `"`
	case "singlequote":
		return "'" + strings.Join(v, "','") + "'"
	default:
		return "{" + strings.Join(v, ",") + "}"
	}
}
