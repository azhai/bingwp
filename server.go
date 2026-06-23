package main

import (
	"crypto/tls"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/models"
	"github.com/azhai/bingwp/services"
)

type ServeCmd struct{}

func (c *ServeCmd) Run() {
	cfg := services.LoadConfig()
	if err := setup(cfg); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
	defer models.CloseDB()

	url, err := RunServer(cfg)
	fmt.Printf("🚀 Starting server at %s\n", url)
	fmt.Printf("📁 Image directory: %s\n", cfg.ImageDir)
	fmt.Printf("📁 Thumb directory: %s\n", cfg.ThumbDir)
	fmt.Printf("💾 Database: %s\n", cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func NewServeMux(conf *services.Config) *http.ServeMux {
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
	return mux
}

func RunServer(conf *services.Config) (url string, err error) {
	mux := NewServeMux(conf)
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	if certPath, keyPath, ok := conf.GetCertFile(); ok {
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		url = "https://" + addr
		err = srv.ListenAndServeTLS(certPath, keyPath)
	} else {
		url = "http://" + addr
		err = srv.ListenAndServe()
	}
	return url, err
}
