package maps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// mapServices stands in for Nominatim, OSRM and Overpass, which a test
// cannot make fail on demand. Nominatim knows "Home" and "Work"; routing
// fails its first routeFailures requests with routeStatus, then routes.
type mapServices struct {
	routeCalls, geocodeCalls atomic.Int32
}

func serveMaps(t *testing.T, routeFailures, routeStatus, geocodeFailures int) *mapServices {
	t.Helper()
	m := &mapServices{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		if int(m.geocodeCalls.Add(1)) <= geocodeFailures {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		switch r.URL.Query().Get("q") {
		case "Home":
			w.Write([]byte(`[{"display_name":"Home","lat":"37.3349","lon":"-122.0090"}]`))
		case "Work":
			w.Write([]byte(`[{"display_name":"Work","lat":"37.4220","lon":"-122.0841"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	})
	mux.HandleFunc("GET /routed-car/route/v1/driving/", func(w http.ResponseWriter, r *http.Request) {
		if int(m.routeCalls.Add(1)) <= routeFailures {
			http.Error(w, "down", routeStatus)
			return
		}
		w.Write([]byte(`{"code":"Ok","routes":[{"distance":12000,"duration":900,"geometry":{"coordinates":[[-122.009,37.3349],[-122.0841,37.422]]},"legs":[]}]}`))
	})
	mux.HandleFunc("GET /interpreter", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Query().Get("data"), `node["amenity"]`) {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"elements":[{"type":"node","id":1,"lat":37.335,"lon":-122.009,"tags":{"name":"Corner Cafe","amenity":"cafe"}}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	oldNominatim, oldOSRM, oldOverpass, oldUpstream := nominatimURL, osrmURL, overpassURL, upstream
	nominatimURL, osrmURL, overpassURL = srv.URL+"/search", srv.URL, srv.URL+"/interpreter"
	upstream.Delay = time.Millisecond
	t.Cleanup(func() { nominatimURL, osrmURL, overpassURL, upstream = oldNominatim, oldOSRM, oldOverpass, oldUpstream })
	return m
}

func get(h http.HandlerFunc, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest("GET", target, nil))
	return w
}

// Routing failing for a moment is asked again, and the person gets the
// directions.
func TestDirections_AMomentaryRoutingFailureIsTriedAgain(t *testing.T) {
	m := serveMaps(t, 1, http.StatusBadGateway, 0)
	h := &handler{app: &App{}}

	w := get(h.handleDirections, "/api/directions?from=Home&to=Work&mode=car")
	if w.Code != http.StatusOK {
		t.Fatalf("directions = %d %s, want 200", w.Code, w.Body)
	}
	if m.routeCalls.Load() != 2 {
		t.Errorf("routing asked %d times, want 2", m.routeCalls.Load())
	}
}

// Routing that keeps failing is a bad gateway.
func TestDirections_RoutingThatKeepsFailingIs502(t *testing.T) {
	serveMaps(t, 100, http.StatusServiceUnavailable, 0)
	h := &handler{app: &App{}}

	if w := get(h.handleDirections, "/api/directions?from=Home&to=Work&mode=car"); w.Code != http.StatusBadGateway {
		t.Errorf("directions = %d %s, want 502", w.Code, w.Body)
	}
}

// A place no one can find is not found; geocoding that keeps failing is a
// bad gateway.
func TestGeocode_NotFoundAndUpstreamFailureDiffer(t *testing.T) {
	serveMaps(t, 0, 0, 0)
	h := &handler{app: &App{}}
	if w := get(h.handleGeocode, "/api/geocode?q=Nowhere"); w.Code != http.StatusNotFound {
		t.Errorf("unknown place = %d %s, want 404", w.Code, w.Body)
	}
	if w := get(h.handleGeocode, "/api/geocode?q=Home"); w.Code != http.StatusOK {
		t.Errorf("known place = %d %s, want 200", w.Code, w.Body)
	}

	serveMaps(t, 0, 0, 100)
	if w := get(h.handleGeocode, "/api/geocode?q=Home"); w.Code != http.StatusBadGateway {
		t.Errorf("geocoding down = %d %s, want 502", w.Code, w.Body)
	}
}

// Overpass's places near a point come back as POIs.
func TestFetchOSMPOIs(t *testing.T) {
	serveMaps(t, 0, 0, 0)

	pois, err := FetchOSMPOIs(context.Background(), 37.3349, -122.0090, 2000, "amenity")
	if err != nil {
		t.Fatal(err)
	}
	if len(pois) != 1 || pois[0].Name != "Corner Cafe" || pois[0].Lat != 37.335 {
		t.Errorf("pois = %+v", pois)
	}
}
