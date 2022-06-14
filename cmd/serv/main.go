package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/azhai/bingwp/cmd"
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
		// err = FetchWallPapers()
		// if err != nil {
		// 	panic(err)
		// }
		var rows []*db.WallDaily
		err = db.Query().OrderBy("id").Where("id >= ?", 1).Find(&rows)
		for _, row := range rows {
			// err = FetchNotes(row)
			err = FetchImage(row)
			if err != nil {
				panic(err)
			}
			if row.Id%10 == 0 {
				fmt.Println(row.Id, row.OrigId)
				time.Sleep(50 * time.Millisecond)
			}
		}
		return
	}

	go handlers.SaveMsgData(options.MaxWriteSize)
	addr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	err = NewApp("BingWallPaper").Listen(addr)
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
