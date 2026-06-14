CREATE TABLE IF NOT EXISTS poi_cache (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    osm_type TEXT NOT NULL DEFAULT '',
    osm_id INTEGER NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'nominatim',
    created_at INTEGER NOT NULL
);
