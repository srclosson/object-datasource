package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/srclosson/object-datasource/pkg/models"
)

// metaKey is where the tree path and metadata are kept in a frame's custom metadata
const metaKey = "objectDatasource"

// decorate applies a link's metadata to its results: field config so panels show the unit,
// name, description, thresholds and links; labels on every series, for alert routing; and
// the tree path and full metadata in the frame's custom metadata. A series keeps its own
// labels, then gets the label lookups', then the static labels. Lookups that failed are
// reported as warnings on the frames.
func decorate(frames data.Frames, path string, md models.Metadata, tables []labelTable, lookupErrs []error) {
	for _, f := range frames {
		if f == nil {
			continue
		}
		for _, field := range f.Fields {
			if field.Type().Time() {
				continue
			}
			for _, t := range tables {
				addLabels(field, t.labelsFor(field.Labels))
			}
			addLabels(field, md.Labels)
			field.SetConfig(withMetadata(field.Config, md))
			if md.DisplayName != "" {
				// After the labels are added, so the name can use them too
				field.Config.DisplayNameFromDS = displayName(md.DisplayName, field)
			}
		}
		setCustomMeta(f, path, md)
		for _, err := range lookupErrs {
			f.AppendNotices(data.Notice{Severity: data.NoticeSeverityWarning, Text: err.Error()})
		}
	}
}

// addLabels adds labels the series doesn't already have; its own labels identify it, so they win
func addLabels(field *data.Field, labels map[string]string) {
	if len(labels) == 0 {
		return
	}
	if field.Labels == nil {
		field.Labels = data.Labels{}
	}
	for k, v := range labels {
		if k == "" {
			continue
		}
		if _, ok := field.Labels[k]; !ok {
			field.Labels[k] = v
		}
	}
}

func withMetadata(cfg *data.FieldConfig, md models.Metadata) *data.FieldConfig {
	if cfg == nil {
		cfg = &data.FieldConfig{}
	}
	if md.Unit != "" {
		cfg.Unit = md.Unit
	}
	if md.Description != "" {
		cfg.Description = md.Description
	}
	if md.Min != nil {
		v := data.ConfFloat64(*md.Min)
		cfg.Min = &v
	}
	if md.Max != nil {
		v := data.ConfFloat64(*md.Max)
		cfg.Max = &v
	}
	if len(md.Thresholds) > 0 {
		// The base step covers everything below the first threshold
		steps := []data.Threshold{data.NewThreshold(math.Inf(-1), "green", "")}
		for _, t := range md.Thresholds {
			steps = append(steps, data.NewThreshold(t.Value, t.Color, ""))
		}
		cfg.Thresholds = &data.ThresholdsConfig{Mode: data.ThresholdsModeAbsolute, Steps: steps}
	}
	for _, l := range md.Links {
		cfg.Links = append(cfg.Links, data.DataLink{Title: l.Title, URL: l.URL, TargetBlank: true})
	}
	return cfg
}

// ${__field.labels.x}, ${__field.labels} and ${__field.name}
var fieldVariablePattern = regexp.MustCompile(`\$\{__field\.(name|labels)(?:\.([^}]+))?\}`)

// displayName fills in a display name's field variables. Grafana only does this for display
// names set on a panel, not for names that come with the data. A panel's own display name
// still wins over this one.
func displayName(template string, field *data.Field) string {
	return fieldVariablePattern.ReplaceAllStringFunc(template, func(match string) string {
		m := fieldVariablePattern.FindStringSubmatch(match)
		switch {
		case m[1] == "name":
			return field.Name
		case m[2] != "":
			return field.Labels[m[2]]
		default:
			return formatLabels(field.Labels)
		}
	})
}

// formatLabels writes labels the way Grafana does: {a="1", b="2"}
func formatLabels(labels data.Labels) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s=%q", k, labels[k])
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func setCustomMeta(f *data.Frame, path string, md models.Metadata) {
	if f.Meta == nil {
		f.Meta = &data.FrameMeta{}
	}
	custom, ok := f.Meta.Custom.(map[string]any)
	if !ok {
		// Keep whatever the target datasource put there
		custom = map[string]any{}
		if f.Meta.Custom != nil {
			custom["source"] = f.Meta.Custom
		}
	}
	custom[metaKey] = map[string]any{"path": path, "metadata": md}
	f.Meta.Custom = custom
}
