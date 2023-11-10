package main

import (
	"fmt"
	"runtime"

	"gitee.com/azhai/fiber-u8l/v2"
	"gitee.com/azhai/fiber-u8l/v2/middleware/compress"
	"github.com/azhai/bingwp/cmd"
	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/models"
	"github.com/azhai/gozzo/logging"
)

func main() {
	var err error
	runtime.GOMAXPROCS(1)
	options, settings := cmd.GetOptions()
	logging.SetLoggerDir(settings.App.LogDir)
	models.Setup()
	if options.UpdateData {
		if err = handlers.FetchRecent(); err != nil {
			logging.Error(err)
		}
		if err = handlers.ReadList(1); err != nil {
			logging.Error(err)
		}
		return
	}

	addr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	fmt.Printf("Server is start at %s ...\n", addr)
	err = NewApp("BingWallPaper").Listen(addr)
	if err != nil {
		panic(err)
	}
}

func NewApp(name string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               name,
		Prefork:               true,
		DisableStartupMessage: true,
	})
	app.Use(compress.New())
	app.Static("/static", "./static")
	app.Static("/wallpaper", handlers.ImgSaveDir)
	app.Get("/", handlers.PageHandler)
	app.Get("/:month", handlers.PageHandler)
	return app
}
