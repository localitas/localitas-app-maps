package maps

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const poiImportAutomationName = "Maps: POI Import"

func RegisterPOIAutomation(coreURL, token, appURL string) {
	if automationExists(coreURL, token, poiImportAutomationName) {
		log.Printf("✅ Maps POI automation already registered")
		return
	}

	body := map[string]interface{}{
		"name":        poiImportAutomationName,
		"description": "Imports points of interest from OpenStreetMap Overpass API for cached local search",
		"dag_config": map[string]interface{}{
			"dag_id":      "maps_poi_import",
			"name":        "Maps: POI Import",
			"description": "Bulk imports POIs for configured areas",
			"nodes": []map[string]interface{}{
				{
					"node_id":            "import_amenities",
					"node_type":          "http-api",
					"execution_strategy": "raft-leader",
					"metadata": map[string]interface{}{
						"url":    appURL + "/api/poi/import",
						"method": "POST",
						"body": map[string]interface{}{
							"lat":      37.3349,
							"lon":      -122.0090,
							"radius":   10000,
							"category": "amenity",
						},
						"timeout_ms":      60000,
						"max_retries":     1,
						"expected_status": 200,
					},
				},
			},
		},
		"trigger_type": "periodic",
		"trigger_config": map[string]interface{}{
			"periodic": map[string]interface{}{
				"schedule":    "0 3 * * 0",
				"timezone":    "Local",
				"max_retries": 1,
			},
		},
		"is_enabled": true,
	}

	b, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", coreURL+"/apps/automation/api/automations", bytes.NewReader(b))
	if err != nil {
		log.Printf("⚠️  Failed to create POI automation request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️  Failed to register POI automation: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		log.Printf("✅ Registered POI import automation (weekly Sunday 3am)")
	} else {
		log.Printf("⚠️  POI automation registration returned %d", resp.StatusCode)
	}
}

func automationExists(coreURL, token, name string) bool {
	req, err := http.NewRequest("GET", coreURL+"/apps/automation/api/automations", nil)
	if err != nil {
		return false
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var result struct {
		Automations []struct {
			Name string `json:"name"`
		} `json:"automations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}
	for _, a := range result.Automations {
		if a.Name == name {
			return true
		}
	}
	return false
}
