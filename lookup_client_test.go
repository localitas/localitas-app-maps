package maps

import (
	"testing"
	"time"
)

// The map lookups ask through the SDK's upstream client, which is pooled and
// pings an idle HTTP/2 connection a NAT may have dropped; a plain client of
// their own would keep a dead connection and stall every try on it.
func TestLookupsUseTheSDKClientAndSaneTimeouts(t *testing.T) {
	if upstream.Client != nil {
		t.Fatal("map lookups bring their own client, without the SDK's stale-connection check")
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
