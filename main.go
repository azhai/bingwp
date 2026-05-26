package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"

	"github.com/alexflint/go-arg"
	"github.com/azhai/bingwp/models"
	"github.com/azhai/bingwp/services"
)

//go:embed views static
var efs embed.FS

func subFS(prefix string) fs.FS {
	fsys, err := fs.Sub(efs, prefix)
	if err != nil {
		panic(err)
	}
	return fsys
}

var (
	viewsFS  = subFS("views")
	staticFS = subFS("static")
)

var args struct {
	Update *UpdateCmd `arg:"subcommand:up" help:"更新壁纸数据（含下载缩略图）"`
	Serve  *ServeCmd  `arg:"subcommand:web" help:"启动Web服务"`
}

// initDB loads config and initializes the database connection
func initDB() error {
	conf := loadConfig()
	services.InitDirs(conf)
	_, err := models.InitDB(conf.DBDSN, conf.LogFile)
	return err
}

// loadConfig is a convenience wrapper to get the current config
func loadConfig() *services.Config {
	return services.LoadConfig()
}

type ServeCmd struct{}

func (c *ServeCmd) Run() {
	conf := loadConfig()
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer models.CloseDB()

	server := NewServer(conf)
	addr := server.Addr
	fmt.Printf("🚀 Starting server at http://%s\n", addr)
	fmt.Printf("📁 Image directory: %s\n", conf.ImageDir)
	fmt.Printf("📁 Thumb directory: %s\n", conf.ThumbDir)
	fmt.Printf("💾 Database: %s\n", conf.DBDSN)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func main() {
	arg.MustParse(&args)

	switch {
	case args.Update != nil:
		args.Update.Run()
	default:
		if args.Serve == nil {
			args.Serve = &ServeCmd{}
		}
		args.Serve.Run()
	}
}
