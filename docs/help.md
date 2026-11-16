---
title: Maps
description: Geocoding, directions, and point-of-interest search
---

# Maps

Geocode addresses, get directions, and search for points of interest using OpenStreetMap data.

Full API spec: swagger.json

## API Endpoints

Method | Path | Description
GET | /api/geocode?q={query} | Geocode address to lat/lon
GET | /api/directions?from={addr}&to={addr}&mode={mode} | Turn-by-turn directions
GET | /api/poi/search?q={query}&category={cat} | Search nearby POIs
GET | /api/poi/autocomplete?q={prefix} | Autocomplete POI names
POST | /api/poi/import | Import POI data from OpenStreetMap

## Geocoding

Convert an address or place name to geographic coordinates. Results are cached locally.

GET /api/geocode?q=Golden+Gate+Bridge

## Directions

Get turn-by-turn directions between two locations. Modes: auto, car, bike, foot.

GET /api/directions?from=San+Francisco&to=Oakland&mode=car

## Points of Interest

Search for nearby businesses, landmarks, and other points of interest. Results come from Nominatim and can be imported into local storage.

## Data Sources

The app uses Nominatim (OpenStreetMap) for geocoding and POI search. Location bias is applied based on your configured region to prioritize local results.

## Caching

Geocoding results are cached locally in SQLite. Repeated lookups for the same address return cached coordinates without hitting external APIs.

## Web Interface

The browser UI displays an interactive map for searching locations, viewing geocoded results, and exploring nearby points of interest.

## Build & Deploy

Run these in the app's repository, `~/localitas-app-maps`.

### Version

```bash
bin/maps-server --version
```

### Build from source

```bash
# Native binary
go build -o bin/maps-server ./cmd/maps-server

# Native binary plus the Linux amd64 binary the Docker image is built from (maps-server-linux-amd64)
make build-linux
```

`make lint` runs gofmt and go vet; `make test` runs them and then the tests. `make build` runs only the lint step.

### Run natively

```bash
make start   # writes the launchd service com.localitas.app.maps for bin/maps-server and loads it on port 9207
make stop
make logs    # follows ~/.localitas/logs/maps/stdout.log
```

The app announces itself to core when it starts, and core serves it at `/apps/ext/maps/`.

### Docker

```bash
make build-docker   # builds the Linux binary and the maps:latest image
make start-docker   # runs container "maps" on port 9207, labelled localitas.app=true
make stop-docker
make logs-docker
```

Core registers containers labelled `localitas.app=true` once Docker reports them healthy.

### Publish

```bash
# Run the tests, build the image, push ghcr.io/localitas/localitas-app-maps:latest and :<VERSION>
make docker-push

# Build dist/maps-server-darwin-arm64 and dist/maps-server-darwin-amd64
# and create GitHub release vX.Y.Z on this repository with them
make release VERSION=vX.Y.Z
```

`VERSION` defaults to `git describe --tags --always --dirty`.

### Download

Pre-built macOS binaries are on the [releases page](https://github.com/localitas/localitas-app-maps/releases):
- `maps-server-darwin-arm64` (Apple Silicon)
- `maps-server-darwin-amd64` (Intel)

Every push to `main` publishes a release: `.github/workflows/release.yml` runs the tests and releases both binaries as `v0.1.<commit count>.<UTC timestamp>`. `make release` publishes one by hand.

Download with the GitHub CLI:

    gh release download --repo localitas/localitas-app-maps --pattern 'maps-server-*'
