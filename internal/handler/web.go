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

// Index serves the main HTML page.
func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(h.templateDir, "index.html"))
}

// StaticHandler returns a file server for static assets.
func (h *WebHandler) StaticHandler() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.Dir(h.staticDir)))
}
