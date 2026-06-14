package maps

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleSwagger(t *testing.T) {
	w := httptest.NewRecorder()
	HandleSwagger(w, httptest.NewRequest("GET", "/swagger.json", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var spec APIDoc
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if spec.AppName != "Maps" {
		t.Errorf("expected app_name Maps, got %q", spec.AppName)
	}
	if len(spec.Endpoints) == 0 {
		t.Error("expected at least one endpoint")
	}
	hasGeocode := false
	hasDirections := false
	for _, ep := range spec.Endpoints {
		if strings.Contains(ep.Path, "/api/geocode") {
			hasGeocode = true
		}
		if strings.Contains(ep.Path, "/api/directions") {
			hasDirections = true
		}
	}
	if !hasGeocode {
		t.Error("expected /api/geocode endpoint")
	}
	if !hasDirections {
		t.Error("expected /api/directions endpoint")
	}
}

func TestHaversine(t *testing.T) {
	// SF to LA is roughly 559 km
	dist := haversine(37.7749, -122.4194, 34.0522, -118.2437)
	km := dist / 1000
	if km < 500 || km > 600 {
		t.Errorf("SF to LA distance %.0f km, expected ~559 km", km)
	}
}

func TestFmtDistance(t *testing.T) {
	if fmtDistance(100) != "328 ft" {
		t.Errorf("expected feet for short distance, got %s", fmtDistance(100))
	}
	if fmtDistance(5000) != "3.1 mi" {
		t.Errorf("expected miles, got %s", fmtDistance(5000))
	}
}

func TestFmtDuration(t *testing.T) {
	if fmtDuration(300) != "5 min" {
		t.Errorf("expected '5 min', got %s", fmtDuration(300))
	}
	if fmtDuration(3900) != "1 hr 5 min" {
		t.Errorf("expected '1 hr 5 min', got %s", fmtDuration(3900))
	}
}

func TestFmtManeuver(t *testing.T) {
	tests := []struct {
		typ, mod, street, want string
	}{
		{"depart", "", "Main St", "Start on Main St"},
		{"turn", "left", "Oak Ave", "Turn left onto Oak Ave"},
		{"turn", "right", "Broadway", "Turn right onto Broadway"},
		{"new name", "", "Highway 101", "Continue onto Highway 101"},
		{"merge", "slight right", "I-280", "Merge slight right onto I-280"},
		{"fork", "left", "Exit 5", "Keep left onto Exit 5"},
		{"roundabout", "", "Ring Rd", "Enter roundabout, exit onto Ring Rd"},
		{"arrive", "", "", ""},
		{"notification", "", "", ""},
		{"depart", "", "", "Start on the road"},
	}
	for _, tt := range tests {
		got := fmtManeuver(tt.typ, tt.mod, tt.street)
		if got != tt.want {
			t.Errorf("fmtManeuver(%q, %q, %q) = %q, want %q", tt.typ, tt.mod, tt.street, got, tt.want)
		}
	}
}

func TestHandleDirections_MissingParams(t *testing.T) {
	h := &handler{}
	req := httptest.NewRequest("GET", "/api/directions", nil)
	w := httptest.NewRecorder()
	h.handleDirections(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing params, got %d", w.Code)
	}
}

func TestHandleGeocode_MissingParam(t *testing.T) {
	h := &handler{}
	req := httptest.NewRequest("GET", "/api/geocode", nil)
	w := httptest.NewRecorder()
	h.handleGeocode(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing q, got %d", w.Code)
	}
}

func TestModeLabel(t *testing.T) {
	if modeLabel("car") != "Driving" {
		t.Error("car should be Driving")
	}
	if modeLabel("bike") != "Cycling" {
		t.Error("bike should be Cycling")
	}
	if modeLabel("foot") != "Walking" {
		t.Error("foot should be Walking")
	}
}
