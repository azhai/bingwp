package main

import (
	"embed"
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

func setup(cfg *services.Config) error {
	services.InitDirs(cfg)
	_, err := models.OpenDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	return err
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
