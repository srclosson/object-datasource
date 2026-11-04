package models

import (
	"reflect"
	"testing"
)

func TestMetadataForInheritsFromFolders(t *testing.T) {
	one, ten := 1.0, 10.0
	s := &Settings{Folders: map[string]Metadata{
		"plant": {
			Description: "The plant",
			Unit:        "celsius",
			Labels:      map[string]string{"site": "plant1", "owner": "ops"},
			Links:       []Reference{{Title: "Site runbook", URL: "https://example.com/site"}},
			Min:         &one,
		},
		"plant.line1": {
			Labels:       map[string]string{"owner": "line1-team"},
			Links:        []Reference{{Title: "Line manual", URL: "https://example.com/line1"}},
			LabelLookups: []LabelLookup{{UID: "assets"}},
		},
		"elsewhere": {Unit: "percent"},
	}}
	link := QueryLink{Name: "plant.line1.temp", Metadata: Metadata{
		Description: "Line 1 temperature",
		Max:         &ten,
		Links:       []Reference{{Title: "Sensor datasheet", URL: "https://example.com/sensor"}},
	}}

	got := s.MetadataFor(link)
	want := Metadata{
		Description: "Line 1 temperature",
		Unit:        "celsius",
		Labels:      map[string]string{"site": "plant1", "owner": "line1-team"},
		Links: []Reference{
			{Title: "Site runbook", URL: "https://example.com/site"},
			{Title: "Line manual", URL: "https://example.com/line1"},
			{Title: "Sensor datasheet", URL: "https://example.com/sensor"},
		},
		Min:          &one,
		Max:          &ten,
		LabelLookups: []LabelLookup{{UID: "assets"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	if s.Folders["plant"].Labels["owner"] != "ops" {
		t.Error("inheriting changed a folder's own labels")
	}
}

func TestMetadataForWithoutFolders(t *testing.T) {
	s := &Settings{}
	if md := s.MetadataFor(QueryLink{Name: "a.b"}); !md.IsZero() {
		t.Errorf("want no metadata, got %+v", md)
	}
}
