package maps

import (
	"context"
	"testing"
)

func TestFetchOSMPOIs(t *testing.T) {
	pois, err := FetchOSMPOIs(context.Background(), 37.3349, -122.0090, 2000, "amenity")
	if err != nil {
		t.Skipf("Overpass API may be unavailable: %v", err)
	}
	if len(pois) == 0 {
		t.Skip("no POIs returned (Overpass may be rate-limited)")
	}
	found := false
	for _, p := range pois {
		if p.Name != "" && p.Lat != 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one POI with name and coordinates")
	}
}

func TestJoinNonEmpty(t *testing.T) {
	tests := []struct {
		parts []string
		sep   string
		want  string
	}{
		{[]string{"Apple Park", "Cupertino"}, ", ", "Apple Park, Cupertino"},
		{[]string{"", "test", ""}, "-", "test"},
		{[]string{}, ", ", ""},
	}
	for _, tt := range tests {
		got := joinNonEmpty(tt.parts, tt.sep)
		if got != tt.want {
			t.Errorf("joinNonEmpty(%v, %q) = %q, want %q", tt.parts, tt.sep, got, tt.want)
		}
	}
}
