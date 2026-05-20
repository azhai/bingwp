package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
	"github.com/azhai/bingwp/services"
)

var args struct {
	Update    *UpdateCmd  `arg:"subcommand:update" help:"更新壁纸数据"`
	Serve     *ServeCmd   `arg:"subcommand:serve" help:"启动Web服务"`
	ServerOpts             `arg:"embed"`
}

type ServeCmd struct{}

func (c *ServeCmd) Run() {
	dbPath := args.DBPath
	if dbPath == "" {
		dbPath = "./bingwp.db"
	}

	imageDir := args.ImageDir
	if imageDir == "" {
		imageDir = "./images"
	}

	db, err := services.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	server := NewServer(ServerOpts{
		Port:     args.Port,
		DBPath:   dbPath,
		ImageDir: imageDir,
	}, db)

	addr := server.Addr
	fmt.Printf("🚀 Starting server at http://localhost%s\n", addr)
	fmt.Printf("📁 Image directory: %s\n", imageDir)
	fmt.Printf("💾 Database: %s\n", dbPath)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func main() {
	arg.MustParse(&args)

	switch {
	case args.Update != nil:
		args.Update.Run()
	case args.Serve != nil:
		args.Serve.Run()
	default:
		args.Serve = &ServeCmd{}
		args.Serve.Run()
	}
}
