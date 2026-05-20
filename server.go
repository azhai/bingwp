package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/azhai/bingwp/handlers"
)

type ServerOpts struct {
	Port     int    `arg:"-p,--port" default:"8080" help:"服务端口"`
	DBPath   string `arg:"--db-path" help:"数据库路径"`
	ImageDir string `arg:"--image-dir" help:"图片目录"`
}

func NewServer(opts ServerOpts, db *sql.DB) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.FileServer(http.Dir("./")))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(opts.ImageDir))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.PageHandler(w, r, db)
	})

	addr := fmt.Sprintf(":%d", opts.Port)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
