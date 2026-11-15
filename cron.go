package maps

import (
	"net/http"

	client "github.com/localitas/localitas-go"
)

var cronSpec = client.CronSpec{
	Jobs: []client.CronJob{
		{
			ID:          "cron:maps:poi-import",
			Path:        "/api/poi/import",
			Method:      "POST",
			Schedule:    "0 3 * * 0",
			Description: "Imports points of interest from OpenStreetMap Overpass API",
			Timeout:     "60s",
			Retry:       &client.CronRetry{MaxAttempts: 1},
		},
	},
}

func HandleCron(w http.ResponseWriter, r *http.Request) {
	cronSpec.ServeHTTP(w, r)
}
