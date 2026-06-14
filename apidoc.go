package maps

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type APIEndpoint struct {
	Method      string     `json:"method"`
	Path        string     `json:"path"`
	Summary     string     `json:"summary"`
	QueryParams []APIParam `json:"query_params,omitempty"`
	Response    *APIBody   `json:"response,omitempty"`
}

type APIParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIBody struct {
	ContentType string `json:"content_type"`
	Example     string `json:"example"`
}

type APIDoc struct {
	AppName     string        `json:"app_name"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Keywords    []string      `json:"keywords,omitempty"`
	Endpoints   []APIEndpoint `json:"endpoints"`
}

var MapsAPIDoc = APIDoc{
	AppName:     "Maps",
	Version:     "0.1.0",
	Description: "Geocoding and driving/walking/cycling directions using OpenStreetMap + OSRM",
	Keywords:    []string{"maps", "map", "directions", "route", "navigate", "navigation", "geocode", "address", "location", "driving", "walking", "cycling", "distance", "ETA"},
	Endpoints: []APIEndpoint{
		{
			Method:  "GET",
			Path:    "/api/geocode",
			Summary: "Geocode an address to coordinates",
			QueryParams: []APIParam{
				{Name: "q", Type: "string", Required: true, Description: "Address or place name"},
			},
			Response: &APIBody{ContentType: "application/json", Example: `{"name":"San Francisco, CA","lat":37.7749,"lon":-122.4194}`},
		},
		{
			Method:  "GET",
			Path:    "/api/directions",
			Summary: "Get driving/walking/cycling directions with route geometry",
			QueryParams: []APIParam{
				{Name: "from", Type: "string", Required: true, Description: "Starting address"},
				{Name: "to", Type: "string", Required: true, Description: "Destination address"},
				{Name: "mode", Type: "string", Description: "auto (default), car, bike, or foot"},
			},
			Response: &APIBody{ContentType: "application/json", Example: `{"from":{"name":"...","lat":37.77,"lon":-122.41},"to":{"name":"...","lat":37.33,"lon":-121.89},"mode":"car","distance":"48.2 mi","duration":"52 min","steps":[{"instruction":"Start on Market St","distance":"0.3 mi"}],"route_coords":[[37.77,-122.41],[37.33,-121.89]]}`},
		},
	},
}

func HandleSwagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MapsAPIDoc)
}

func RenderDocsHTML(doc APIDoc) template.HTML {
	var sb strings.Builder
	sb.WriteString(`<h3 style="font-size: 0.875rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-secondary); margin-bottom: 1rem;">API</h3><div class="accordion-list">`)
	for _, ep := range doc.Endpoints {
		title := fmt.Sprintf("%s %s — %s", ep.Method, ep.Path, ep.Summary)
		sb.WriteString(fmt.Sprintf(`<details class="glass-panel" style="border-radius: 0.5rem; margin-bottom: 0.5rem;"><summary style="padding: 0.75rem 1rem; cursor: pointer; font-weight: 500; color: var(--color-text-primary);">%s</summary><div style="padding: 0 1rem 0.75rem; font-size: 0.875rem; color: var(--color-text-secondary);">`, template.HTMLEscapeString(title)))
		if len(ep.QueryParams) > 0 {
			sb.WriteString("<p>")
			for i, p := range ep.QueryParams {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("<code>%s</code>", p.Name))
				if p.Required {
					sb.WriteString(" (required)")
				}
			}
			sb.WriteString("</p>")
		}
		if ep.Response != nil {
			sb.WriteString(fmt.Sprintf(`<pre style="background: var(--color-bg-base); padding: 0.75rem; border-radius: 0.375rem; overflow-x: auto; font-size: 0.8125rem;">%s</pre>`, template.HTMLEscapeString(prettyJSON(ep.Response.Example))))
		}
		sb.WriteString(`</div></details>`)
	}
	sb.WriteString(`</div>`)
	return template.HTML(sb.String())
}

func prettyJSON(s string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
