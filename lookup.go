package maps

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/localitas/localitas-go"
)

func nominatimLocationBias() string {
	return client.LocationBias()
}

func Geocode(ctx context.Context, address string) (*Location, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	u := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1&accept-language=en%s", url.QueryEscape(address), nominatimLocationBias())
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("User-Agent", "Localitas Maps/1.0")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode failed: %w", err)
	}
	defer resp.Body.Close()
	var results []struct {
		DisplayName string `json:"display_name"`
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil || len(results) == 0 {
		return nil, fmt.Errorf("no location found for: %s", address)
	}
	var lat, lon float64
	fmt.Sscanf(results[0].Lat, "%f", &lat)
	fmt.Sscanf(results[0].Lon, "%f", &lon)
	return &Location{Name: results[0].DisplayName, Lat: lat, Lon: lon}, nil
}

func GeocodeMulti(ctx context.Context, address string, limit int) ([]Location, error) {
	if address == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=%d&accept-language=en%s", url.QueryEscape(address), limit, nominatimLocationBias())
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("User-Agent", "Localitas Maps/1.0")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var results []struct {
		DisplayName string `json:"display_name"`
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	var locs []Location
	for _, r := range results {
		var lat, lon float64
		fmt.Sscanf(r.Lat, "%f", &lat)
		fmt.Sscanf(r.Lon, "%f", &lon)
		locs = append(locs, Location{Name: r.DisplayName, Lat: lat, Lon: lon})
	}
	return locs, nil
}

func GeocodeWithCache(ctx context.Context, store *Store, query string) (*Location, error) {
	loc, err := Geocode(ctx, query)
	if err != nil && store != nil {
		cached, _ := store.SearchPOI(ctx, query, 1)
		if len(cached) > 0 {
			return &Location{Name: cached[0].DisplayName, Lat: cached[0].Lat, Lon: cached[0].Lon}, nil
		}
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if store != nil {
		store.CachePOI(ctx, query, loc.Name, loc.Lat, loc.Lon, "", "", 0, "nominatim")
	}
	return loc, nil
}

func GetDirections(ctx context.Context, store *Store, fromAddr, toAddr, mode string) (*DirectionsResult, error) {
	from, err := GeocodeWithCache(ctx, store, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("from: %w", err)
	}
	to, err := GeocodeWithCache(ctx, store, toAddr)
	if err != nil {
		return nil, fmt.Errorf("to: %w", err)
	}

	if mode == "" || mode == "auto" {
		dist := haversine(from.Lat, from.Lon, to.Lat, to.Lon)
		if dist < 1609.34 {
			mode = "foot"
		} else {
			mode = "car"
		}
	}

	// FOSSGIS OSRM servers support car/bike/foot profiles
	osrmServer := "https://routing.openstreetmap.de/routed-car"
	if mode == "bike" {
		osrmServer = "https://routing.openstreetmap.de/routed-bike"
	} else if mode == "foot" {
		osrmServer = "https://routing.openstreetmap.de/routed-foot"
	}

	u := fmt.Sprintf("%s/route/v1/driving/%.7f,%.7f;%.7f,%.7f?overview=full&geometries=geojson&steps=true",
		osrmServer, from.Lon, from.Lat, to.Lon, to.Lat)

	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("User-Agent", "Localitas Maps/1.0")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("routing failed: %w", err)
	}
	defer resp.Body.Close()

	var osrm struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"`
			} `json:"geometry"`
			Legs []OSRMLeg `json:"legs"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&osrm); err != nil {
		return nil, fmt.Errorf("parse route: %w", err)
	}
	if osrm.Code != "Ok" || len(osrm.Routes) == 0 {
		return nil, fmt.Errorf("no route found (osrm code: %s, from: %s [%.4f,%.4f], to: %s [%.4f,%.4f])", osrm.Code, from.Name, from.Lat, from.Lon, to.Name, to.Lat, to.Lon)
	}

	route := osrm.Routes[0]
	result := &DirectionsResult{
		From:     *from,
		To:       *to,
		Mode:     mode,
		Distance: fmtDistance(route.Distance),
		Duration: fmtDuration(route.Duration),
	}

	// Convert coords from [lon,lat] to [lat,lon] for Leaflet
	for _, c := range route.Geometry.Coordinates {
		if len(c) >= 2 {
			result.RouteCoords = append(result.RouteCoords, []float64{c[1], c[0]})
		}
	}

	for _, leg := range route.Legs {
		for _, step := range leg.Steps {
			instr := fmtManeuver(step.Maneuver.Type, step.Maneuver.Modifier, step.Name)
			if instr != "" {
				result.Steps = append(result.Steps, RouteStep{
					Instruction: instr,
					Distance:    fmtDistance(step.Distance),
				})
			}
		}
	}
	result.Steps = append(result.Steps, RouteStep{Instruction: "Arrive at destination", Distance: ""})

	return result, nil
}

func fmtDistance(meters float64) string {
	miles := meters / 1609.34
	if miles < 0.1 {
		return fmt.Sprintf("%.0f ft", meters*3.28084)
	}
	return fmt.Sprintf("%.1f mi", miles)
}

func fmtDuration(seconds float64) string {
	mins := int(seconds / 60)
	if mins < 60 {
		return fmt.Sprintf("%d min", mins)
	}
	h, m := mins/60, mins%60
	if m == 0 {
		return fmt.Sprintf("%d hr", h)
	}
	return fmt.Sprintf("%d hr %d min", h, m)
}

func fmtManeuver(typ, modifier, street string) string {
	if street == "" {
		street = "the road"
	}
	switch typ {
	case "depart":
		return fmt.Sprintf("Start on %s", street)
	case "turn":
		return fmt.Sprintf("Turn %s onto %s", modifier, street)
	case "new name":
		return fmt.Sprintf("Continue onto %s", street)
	case "merge":
		return fmt.Sprintf("Merge %s onto %s", modifier, street)
	case "on ramp":
		return fmt.Sprintf("Take the ramp onto %s", street)
	case "off ramp":
		return fmt.Sprintf("Take the exit onto %s", street)
	case "fork":
		return fmt.Sprintf("Keep %s onto %s", modifier, street)
	case "end of road":
		return fmt.Sprintf("Turn %s onto %s", modifier, street)
	case "continue":
		return fmt.Sprintf("Continue on %s", street)
	case "roundabout", "rotary":
		return fmt.Sprintf("Enter roundabout, exit onto %s", street)
	case "arrive", "notification":
		return ""
	default:
		if modifier != "" {
			return fmt.Sprintf("Go %s on %s", modifier, street)
		}
		return fmt.Sprintf("Continue on %s", street)
	}
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func modeLabel(mode string) string {
	switch mode {
	case "car":
		return "Driving"
	case "bike":
		return "Cycling"
	case "foot":
		return "Walking"
	default:
		return strings.Title(mode)
	}
}
