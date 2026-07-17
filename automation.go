package maps

import (
	"context"

	client "github.com/localitas/localitas-go"
)

const poiImportAutomationName = "Maps: POI Import"

func RegisterPOIAutomation(ctx context.Context, c *client.Client, appURL string) {
	if automationExists(ctx, c, poiImportAutomationName) {
		logger.Info("POI automation already registered")
		return
	}

	req := client.CreateAutomationRequest{
		Name:        poiImportAutomationName,
		Description: "Imports points of interest from OpenStreetMap Overpass API for cached local search",
		DAGConfig: client.DAGConfig{
			DAGID:       "maps_poi_import",
			Name:        "Maps: POI Import",
			Description: "Bulk imports POIs for configured areas",
			Nodes: []client.DAGNode{
				{
					NodeID:            "import_amenities",
					NodeType:          "http-api",
					ExecutionStrategy: "raft-leader",
					Metadata: map[string]any{
						"url":    appURL + "/api/poi/import",
						"method": "POST",
						"body": map[string]any{
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
		TriggerType: "periodic",
		TriggerConfig: client.TriggerConfig{
			Periodic: &client.PeriodicTrigger{
				Schedule:   "0 3 * * 0",
				Timezone:   "Local",
				MaxRetries: 1,
			},
		},
		IsEnabled: true,
	}

	if _, err := c.Automation().Create(ctx, req); err != nil {
		logger.Error("failed to register POI automation", "error", err)
		return
	}
	logger.Info("registered POI import automation", "schedule", "weekly Sunday 3am")
}

func automationExists(ctx context.Context, c *client.Client, name string) bool {
	automations, err := c.Automation().List(ctx)
	if err != nil {
		return false
	}
	for _, a := range automations {
		if a.Name == name {
			return true
		}
	}
	return false
}
