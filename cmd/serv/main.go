package main

import (
	"fmt"
	"runtime"

	"github.com/azhai/bingwp/cmd"
	"github.com/azhai/bingwp/geohash"
	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/models"
	db "github.com/azhai/bingwp/models/default"

	"gitee.com/azhai/fiber-u8l/v2"
	"gitee.com/azhai/fiber-u8l/v2/middleware/compress"
)

func main() {
	var err error
	runtime.GOMAXPROCS(1)
	options, _ := cmd.GetOptions()
	models.Setup()
	if options.UpdateData {
		err = FetchWallPapers()
		if err != nil {
			panic(err)
		}
		err = ArrangeImages()

		coord := geohash.NewCoordinate(0)
		var rows []*db.WallLocation
		db.Query().Where("geohash = '' AND latitude <> 0").Find(&rows)
		for _, row := range rows {
			hash := coord.Encode(row.Latitude, row.Longitude)
			row.Save(map[string]any{"geohash": hash})
		}
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
