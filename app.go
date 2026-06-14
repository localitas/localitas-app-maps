package maps

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/localitas/localitas-go"
)

type App struct {
	Store    *Store
	BasePath string
	client   *client.Client
}

func New(c *client.Client, basePath string) *App {
	if basePath == "" {
		basePath = "/"
	}
	return &App{BasePath: basePath, client: c}
}

func (a *App) InitStore(coreURL, dbID, token string) error {
	store, err := OpenStore(coreURL, dbID, token)
	if err != nil {
		return err
	}
	a.Store = store
	return nil
}

func (a *App) Install(ctx context.Context) (string, error) {
	for attempt := 1; ; attempt++ {
		db, err := a.client.CreateSystemDatabase(ctx, DatabaseName)
		if err != nil {
			log.Printf("install: attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := applyEmbeddedMigrations(ctx, a.client, db.ID); err != nil {
			log.Printf("install: migrations attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		return db.ID, nil
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/index.html")
	if err != nil {
		log.Printf("maps index template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	tmpl.ExecuteTemplate(w, "index.html", map[string]string{"BasePath": a.BasePath})
}

func (a *App) RegisterRoutes(mux *http.ServeMux) {
	h := &handler{app: a}
	mux.HandleFunc("GET /{$}", a.handleIndex)
	mux.HandleFunc("GET /swagger.json", HandleSwagger)
	mux.HandleFunc("GET /help.md", handleHelpMarkdown)
	mux.HandleFunc("GET /api/geocode", h.handleGeocode)
	mux.HandleFunc("GET /api/directions", h.handleDirections)
	mux.HandleFunc("GET /api/poi/search", h.handlePOISearch)
	mux.HandleFunc("GET /api/poi/autocomplete", h.handlePOIAutocomplete)
	mux.HandleFunc("POST /api/poi/import", h.handlePOIImport)
}
