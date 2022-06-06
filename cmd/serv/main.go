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
	runtime.GOMAXPROCS(1)
	options, _ := cmd.GetOptions()
	models.Setup()
	go handlers.SaveMsgData(options.MaxWriteSize)
	if options.UpdateData {
		FetchWallPapers()
	}

	addr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	err := NewApp("BingWallPaper").Listen(addr)
	if err == nil {
		fmt.Printf("Server is start at %s ...\n", addr)
	} else {
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
	app.Get("*", handlers.MyGetHandler)
	app.Post("*", handlers.MyPostHandler)
	return app
}
