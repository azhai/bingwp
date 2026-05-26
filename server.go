package main

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/services"
)

func NewServer(conf *services.Config) *http.Server {
	mux := http.NewServeMux()

	// Static files (embedded)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Thumbnail images
	mux.Handle("/thumbs/", http.StripPrefix("/thumbs/", http.FileServer(http.Dir(conf.ThumbDir))))

	// Image files
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(conf.ImageDir))))

	// API endpoint
	mux.HandleFunc("/api/wallpapers", handlers.APIWallpapersHandler)

	// SPA: all other routes serve the embedded index page
	indexHTML, _ := fs.ReadFile(viewsFS, "index.html")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
