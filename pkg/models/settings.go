package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type QueryLinkConfigBase struct {
	Name string  `json:"name,omitempty"`
	UID  string  `json:"uid,omitempty"`
	ID   float64 `json:"id,omitempty"`
}

type SettingsIn struct {
	QueryLinkConfigBase
	QueryLinks []map[string]interface{} `json:"queryLinks,omitempty"`
}

// Settings - all filled from
type Settings struct {
	backend.DataSourceInstanceSettings
	Request       ProxiedDataRequest
	Authorization string
}

// LoadSettings will read and validate Settings from the DataSourceConfg
func LoadSettings(config backend.DataSourceInstanceSettings) (*Settings, error) {
	settingsIn := &SettingsIn{}
	settings := &Settings{}

	backend.Logger.Debug("config", "config", config.JSONData)

	if err := json.Unmarshal(config.JSONData, &settingsIn); err != nil {
		return settings, fmt.Errorf("could not unmarshal DataSourceInfo json: %w", err)
	}

	settings.URL = config.URL
	settings.ID = config.ID
	settings.Name = config.Name
	settings.User = config.User
	settings.Database = config.Database
	settings.BasicAuthEnabled = config.BasicAuthEnabled
	settings.BasicAuthUser = config.BasicAuthUser

	settings.Request = ProxiedDataRequest{
		From:    fmt.Sprintf("%d", time.Now().Add(-1*time.Duration(10)*time.Minute).UnixNano()/int64(time.Millisecond)),
		To:      fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
		Queries: make([]json.RawMessage, 0),
	}

	for _, queryLink := range settingsIn.QueryLinks {
		//query := queryLink["query"].(map[string]interface{})
		json, err := json.Marshal(queryLink["query"])
		backend.Logger.Debug("Adding query", "json", string(json))
		if err != nil {
			return nil, err
		}

		settings.Request.Queries = append(settings.Request.Queries, json)
	}
	settings.Authorization = GetBasicAuthFromUsernameAndPassword(config.BasicAuthUser, config.DecryptedSecureJSONData["basicAuthPassword"])

	return settings, nil
}

func GetBasicAuthFromUsernameAndPassword(username string, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString(([]byte)(strings.TrimSpace(username)+":"+strings.TrimSpace(password)))
}
