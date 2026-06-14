package maps

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

const DatabaseName = "maps"

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func OpenStore(coreURL, dbID, token string) (*Store, error) {
	dsn := fmt.Sprintf("%s?database_id=%s&token=%s", coreURL, dbID, token)
	db, err := sql.Open("localitas", dsn)
	if err != nil {
		return nil, err
	}
	return NewStore(db), nil
}

func (s *Store) Close() error { return s.db.Close() }

type POI struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Category    string  `json:"category"`
	OSMType     string  `json:"osm_type,omitempty"`
	OSMID       int64   `json:"osm_id,omitempty"`
	Source      string  `json:"source"`
}

func (s *Store) SearchPOI(ctx context.Context, query string, limit int) ([]*POI, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.display_name, p.lat, p.lon, p.category, p.osm_type, p.osm_id, p.source
		FROM poi_cache p
		JOIN poi_fts ON p.rowid = poi_fts.rowid
		WHERE poi_fts MATCH ?
		ORDER BY rank
		LIMIT ?`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pois []*POI
	for rows.Next() {
		var p POI
		rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Lat, &p.Lon, &p.Category, &p.OSMType, &p.OSMID, &p.Source)
		pois = append(pois, &p)
	}
	return pois, nil
}

func (s *Store) CachePOI(ctx context.Context, name, displayName string, lat, lon float64, category, osmType string, osmID int64, source string) error {
	id := newPOIID()
	now := time.Now().UTC().Unix()
	result, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO poi_cache (id, name, display_name, lat, lon, category, osm_type, osm_id, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, name, displayName, lat, lon, category, osmType, osmID, source, now)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows > 0 {
		s.db.ExecContext(ctx, "INSERT INTO poi_fts(rowid, name, display_name, category) SELECT rowid, name, display_name, category FROM poi_cache WHERE id = ?", id)
	}
	return nil
}

func (s *Store) BulkInsertPOIs(ctx context.Context, pois []POI) (int, error) {
	count := 0
	for _, p := range pois {
		err := s.CachePOI(ctx, p.Name, p.DisplayName, p.Lat, p.Lon, p.Category, p.OSMType, p.OSMID, p.Source)
		if err == nil {
			count++
		}
	}
	return count, nil
}

func (s *Store) GetPOICount(ctx context.Context) int {
	var count int
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM poi_cache").Scan(&count)
	return count
}

func newPOIID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
