package handler

import (
	"net/http"
	"path/filepath"
)

// WebHandler serves the frontend files.
type WebHandler struct {
	templateDir string
	staticDir   string
}

// NewWebHandler creates a new WebHandler.
func NewWebHandler(webDir string) *WebHandler {
	return &WebHandler{
		templateDir: filepath.Join(webDir, "templates"),
		staticDir:   filepath.Join(webDir, "static"),
	}
}

// Index serves the main HTML page. Never cached — always fresh.
func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, filepath.Join(h.templateDir, "index.html"))
}

// StaticHandler returns a file server for static assets.
// Revalidates on every request so JS/CSS changes are picked up immediately.
func (h *WebHandler) StaticHandler() http.Handler {
	fs := http.FileServer(http.Dir(h.staticDir))
	return http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		fs.ServeHTTP(w, r)
	}))
}
