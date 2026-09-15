package maps

import (
	"testing"
	"time"
)

// The geocode/directions/POI lookups used to allocate a fresh http.Client per
// request. They now share one pooled client and bound each call with a context
// deadline. Assert the shared client exists and the per-operation timeouts are
// sane positive bounds.
func TestSharedHTTPClientAndTimeouts(t *testing.T) {
	if httpClient == nil {
		t.Fatal("map lookups must share a pooled http client")
	}
	for name, d := range map[string]time.Duration{
		"geocode":    geocodeTimeout,
		"directions": directionsTimeout,
		"osmPOI":     osmPOITimeout,
	} {
		if d <= 0 || d > time.Minute {
			t.Errorf("%s timeout = %v, want a small positive bound", name, d)
		}
	}
}
