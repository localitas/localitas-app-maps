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

### Version

```bash
./maps-server --version
```

### Build from source

```bash
# Development (native)
cd apps/maps && go build -o bin/maps-server ./cmd/maps-server

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o bin/maps-server-linux-amd64 ./cmd/maps-server
```

### Docker

Build a Docker image directly from the binary:

```bash
# Default base image (debian:12-slim)
./maps-server docker-build

# Custom base image
./maps-server docker-build --base ubuntu:24.04

# Custom Dockerfile
./maps-server docker-build --dockerfile ./my.Dockerfile

# Tag and push to registry
./maps-server docker-build --tag ghcr.io/localitas/maps:latest --push
```

The `docker-build` command requires a Linux amd64 binary in the same directory. Run `make deploy-build` from the project root first.

### Download

Pre-built binaries are available on the [GitHub releases page](https://github.com/localitas/localitas/releases).

Each release includes three builds per app:
- `maps-server-darwin-arm64` (macOS Apple Silicon)
- `maps-server-linux-amd64` (Linux x86_64)
- `maps-server-linux-arm64` (Linux ARM64)

Download with the GitHub CLI:

    gh release download --repo localitas/localitas --pattern 'maps-server-*'

### Release

All app binaries are published to GitHub releases as part of `make deploy-upload-image`.
