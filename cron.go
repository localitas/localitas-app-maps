package maps

import (
	"encoding/json"
	"net/http"
)

func HandleCron(w http.ResponseWriter, r *http.Request) {
	spec := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"id":          "cron:maps:poi-import",
				"path":        "/api/poi/import",
				"method":      "POST",
				"schedule":    "0 3 * * 0",
				"description": "Imports points of interest from OpenStreetMap Overpass API",
				"timeout":     "60s",
				"retry": map[string]interface{}{
					"max_attempts": 1,
				},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}
