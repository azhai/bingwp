package main

import (
	"fmt"
	"runtime"

	"github.com/azhai/bingwp/cmd"
	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/models"

	"gitee.com/azhai/fiber-u8l/v2"
	"gitee.com/azhai/fiber-u8l/v2/middleware/compress"
)

func main() {
	var err error
	runtime.GOMAXPROCS(1)
	options, _ := cmd.GetOptions()
	models.Setup()
	if options.UpdateData {
		err = handlers.FetchRecent()
		if err != nil {
			panic(err)
		}
		handlers.RepairImage()
		handlers.UpdateGeo()
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
	app.Static("/wallpaper", "/data/bingwp")
	app.Get("/", handlers.HomeHandler)
	app.Get("/:month", handlers.HomeHandler)
	app.Post("/data/", handlers.DataHandler)
	return app
}
