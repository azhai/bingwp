package main

import (
	"fmt"
	"runtime"

	"github.com/azhai/bingwp/cmd"
	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/gozzo/logging"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// NewApp 创建http服务
func NewApp(name, imgDir string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:     name,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})
	app.Use(compress.New())
	app.Use("/static", static.New("./static"))
	app.Get("/wallpaper/*", static.New(imgDir))
	app.Get("/", handlers.PageHandler)
	app.Get("/:month", handlers.PageHandler)
	return app
}

func main() {
	var err error
	runtime.GOMAXPROCS(1)
	appName, options := cmd.GetAppOptions()
	if options.ImageDir != "" {
		handlers.SetImageSaveDir(options.ImageDir)
	}
	if options.UpdateData {
		if err = handlers.FetchRecent(); err != nil {
			logging.Error(err)
		}
		if err = handlers.ReadList(1); err != nil {
			logging.Error(err)
		}
		return
	}

	addr := options.GetAddr()
	greeting := fmt.Sprintf("Server is start at %s ...", addr)
	fmt.Println(greeting)
	logging.Info(greeting)
	app := NewApp(appName, options.ImageDir)
	if err = app.Listen(addr); err != nil {
		panic(err)
	}
}
