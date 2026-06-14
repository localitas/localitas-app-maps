package maps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type handler struct {
	app *App
}

func (h *handler) handleGeocode(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeErr(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	loc, err := GeocodeWithCache(r.Context(), h.app.Store, q)
	if err != nil {
		writeErr(w, http.StatusNotFound, "%v", err)
		return
	}
	writeJSON(w, http.StatusOK, loc)
}

func (h *handler) handlePOIAutocomplete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeErr(w, http.StatusBadRequest, "q is required")
		return
	}

	var results []Location

	if h.app.Store != nil {
		cached, _ := h.app.Store.SearchPOI(r.Context(), q, 5)
		for _, p := range cached {
			results = append(results, Location{Name: p.DisplayName, Lat: p.Lat, Lon: p.Lon})
		}
	}

	if len(results) < 5 {
		nominatimResults, _ := GeocodeMulti(r.Context(), q, 5-len(results))
		for _, loc := range nominatimResults {
			results = append(results, loc)
			if h.app.Store != nil {
				h.app.Store.CachePOI(r.Context(), q, loc.Name, loc.Lat, loc.Lon, "", "", 0, "nominatim")
			}
		}
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *handler) handleDirections(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	mode := r.URL.Query().Get("mode")
	if from == "" || to == "" {
		writeErr(w, http.StatusBadRequest, "'from' and 'to' query parameters are required")
		return
	}
	result, err := GetDirections(r.Context(), h.app.Store, from, to, mode)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) handlePOISearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeErr(w, http.StatusBadRequest, "q is required")
		return
	}
	if h.app.Store == nil {
		writeJSON(w, http.StatusOK, []POI{})
		return
	}
	pois, err := h.app.Store.SearchPOI(r.Context(), q, 20)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	writeJSON(w, http.StatusOK, pois)
}

func (h *handler) handlePOIImport(w http.ResponseWriter, r *http.Request) {
	if h.app.Store == nil {
		writeErr(w, http.StatusInternalServerError, "no store configured")
		return
	}
	var req struct {
		Lat      float64 `json:"lat"`
		Lon      float64 `json:"lon"`
		Radius   int     `json:"radius"`
		Category string  `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Radius <= 0 {
		req.Radius = 5000
	}
	if req.Category == "" {
		req.Category = "amenity"
	}

	pois, err := FetchOSMPOIs(r.Context(), req.Lat, req.Lon, req.Radius, req.Category)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	count, _ := h.app.Store.BulkInsertPOIs(r.Context(), pois)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"imported": count,
		"total":    h.app.Store.GetPOICount(r.Context()),
	})
}

func FetchOSMPOIs(ctx context.Context, lat, lon float64, radius int, category string) ([]POI, error) {
	query := fmt.Sprintf(`[out:json][timeout:30];
(
  node["%s"](around:%d,%f,%f);
  way["%s"](around:%d,%f,%f);
);
out center tags 500;`, category, radius, lat, lon, category, radius, lat, lon)

	overpassURL := "https://overpass-api.de/api/interpreter?data=" + url.QueryEscape(query)
	req, _ := http.NewRequestWithContext(ctx, "GET", overpassURL, nil)
	req.Header.Set("User-Agent", "Localitas Maps/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("overpass query failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Elements []struct {
			Type   string  `json:"type"`
			ID     int64   `json:"id"`
			Lat    float64 `json:"lat"`
			Lon    float64 `json:"lon"`
			Center *struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"center"`
			Tags map[string]string `json:"tags"`
		} `json:"elements"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse overpass: %w", err)
	}

	pois := make([]POI, 0, len(result.Elements))
	for _, el := range result.Elements {
		name := el.Tags["name"]
		if name == "" {
			continue
		}
		lat, lon := el.Lat, el.Lon
		if el.Center != nil {
			lat, lon = el.Center.Lat, el.Center.Lon
		}
		cat := el.Tags[category]
		if cat == "" {
			cat = category
		}

		displayParts := []string{name}
		if street := el.Tags["addr:street"]; street != "" {
			if housenumber := el.Tags["addr:housenumber"]; housenumber != "" {
				displayParts = append(displayParts, housenumber+" "+street)
			} else {
				displayParts = append(displayParts, street)
			}
		}
		if city := el.Tags["addr:city"]; city != "" {
			displayParts = append(displayParts, city)
		}

		pois = append(pois, POI{
			Name:        name,
			DisplayName: joinNonEmpty(displayParts, ", "),
			Lat:         lat,
			Lon:         lon,
			Category:    cat,
			OSMType:     el.Type,
			OSMID:       el.ID,
			Source:      "overpass",
		})
	}
	return pois, nil
}

func joinNonEmpty(parts []string, sep string) string {
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	out := ""
	for i, p := range result {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf(format, args...)})
}
