package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/srclosson/object-datasource/pkg/models"
)

// lookupTTL is how long a lookup's rows are reused. Labels such as a machine's manufacturer or
// location rarely change, and every panel refresh would otherwise query them again.
const lookupTTL = time.Minute

// labelTable is a label lookup's rows, each a set of labels
type labelTable struct {
	lookup models.LabelLookup
	rows   []map[string]string
}

// labelsFor returns the labels this table gives a series
func (t labelTable) labelsFor(series data.Labels) map[string]string {
	if t.lookup.MatchLabel == "" {
		if len(t.rows) == 0 {
			return nil
		}
		return t.rows[0]
	}
	key, ok := series[t.lookup.MatchLabel]
	if !ok {
		return nil
	}
	column := t.lookup.MatchColumn
	if column == "" {
		column = t.lookup.MatchLabel
	}
	for _, row := range t.rows {
		if row[column] == key {
			labels := make(map[string]string, len(row))
			for k, v := range row {
				if k != column {
					labels[k] = v
				}
			}
			return labels
		}
	}
	return nil
}

type lookupCache struct {
	mu      sync.Mutex
	entries map[string]lookupEntry
}

type lookupEntry struct {
	rows    []map[string]string
	expires time.Time
}

// labelTables runs the label lookups, reusing recent results. A lookup that fails is
// reported, and the query goes on without its labels.
func (d *Datasource) labelTables(ctx context.Context, orgID int64, lookups []models.LabelLookup, vars map[string]models.VarValue, tr backend.TimeRange) ([]labelTable, []error) {
	var tables []labelTable
	var errs []error
	for i, l := range lookups {
		model := map[string]any{}
		if len(l.Query) > 0 {
			if err := json.Unmarshal(l.Query, &model); err != nil {
				errs = append(errs, fmt.Errorf("label lookup %d: can't read its query: %w", i+1, err))
				continue
			}
		}
		ds := l.Datasource()
		if ds.UID == "" {
			errs = append(errs, fmt.Errorf("label lookup %d has no datasource", i+1))
			continue
		}
		model["datasource"] = ds
		model["refId"] = "lookup"
		interpolate(model, vars, ds.Type)

		key, _ := json.Marshal([]any{orgID, model})
		rows, err := d.lookups.get(string(key), func() ([]map[string]string, error) {
			res, err := d.grafana.query(ctx, orgID, model, tr, false)
			if err != nil {
				return nil, err
			}
			if res.Error != nil {
				return nil, res.Error
			}
			return labelRows(res.Frames), nil
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("label lookup %d: %w", i+1, err))
			continue
		}
		tables = append(tables, labelTable{lookup: l, rows: rows})
	}
	return tables, errs
}

func (c *lookupCache) get(key string, load func() ([]map[string]string, error)) ([]map[string]string, error) {
	c.mu.Lock()
	e, ok := c.entries[key]
	c.mu.Unlock()
	if ok && time.Now().Before(e.expires) {
		return e.rows, nil
	}
	rows, err := load()
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.entries == nil {
		c.entries = map[string]lookupEntry{}
	}
	c.entries[key] = lookupEntry{rows: rows, expires: time.Now().Add(lookupTTL)}
	c.mu.Unlock()
	return rows, nil
}

// labelRows turns query results into rows of labels. A table's columns become labels, as do
// a series' own labels; a series' values and any times are left out.
func labelRows(frames data.Frames) []map[string]string {
	var rows []map[string]string
	for _, f := range frames {
		if f == nil {
			continue
		}
		for i := 0; i < f.Rows(); i++ {
			row := map[string]string{}
			for _, field := range f.Fields {
				if field.Type().Time() {
					continue
				}
				if len(field.Labels) > 0 {
					for k, v := range field.Labels {
						row[k] = v
					}
					continue
				}
				if s, ok := labelValue(field, i); ok {
					row[field.Name] = s
				}
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}
		}
	}
	return rows
}

func labelValue(field *data.Field, i int) (string, bool) {
	v, ok := field.ConcreteAt(i)
	if !ok {
		return "", false
	}
	switch t := v.(type) {
	case string:
		return t, true
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(t), true
	default:
		return fmt.Sprint(t), true
	}
}
