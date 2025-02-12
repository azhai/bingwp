package main

import (
	"fmt"
	"runtime"

	"github.com/alexflint/go-arg"
	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/models"
	"github.com/azhai/gozzo/config"
	"github.com/azhai/gozzo/logging"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
	// db "github.com/azhai/bingwp/models/default"
)

var app *fiber.App

var args struct {
	Update  *UpdateCmd `arg:"subcommand:up" help:"更新数据"`
	Config  string     `arg:"-c,--config" default:"settings.hcl" help:"配置文件路径"`
	Verbose bool       `arg:"-v,--verbose" help:"输出详细信息"`
	ServerOpts
}

type ServerOpts struct {
	Host     string `arg:"-s,--host" default:"" help:"运行IP"`              // 运行IP
	Port     int    `arg:"-p,--port" default:"9870" help:"运行端口"`          // 运行端口
	ImageDir string `arg:"-d,--dir" help:"图片目录" hcl:"image_dir,optional"` // 图片目录
}

func (t ServerOpts) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

type UpdateCmd struct {
}

func (c *UpdateCmd) Run() {
	// 从微软Bing下载最新的图像，以及标题
	var err error
	num, crawler := 0, handlers.NewCrawler()
	if num, err = crawler.SavelArchive(0, ""); err != nil {
		logging.Error(err)
	}

	// 从wilii.cn读取guid等信息
	if num > 0 {
		handlers.SaveListPages(1, 8)
	}

	// 从详情中读取正文等内容
	// var rows []*db.WallDaily
	// qr := db.Query().Asc("id")
	// if err = qr.Find(&rows); err != nil {
	// 	fmt.Println(err)
	// }
	// for _, row := range rows {
	// 	err = handlers.UpdateDailyDetail(row)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }
}

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

func init() {
	arg.MustParse(&args)
	root, err := config.ReadConfigFile(args.Config, nil)
	if err != nil {
		panic(err)
	}
	if args.ImageDir == "" {
		root.ParseAppRemain(&args.ServerOpts)
	}
	handlers.SetImageSaveDir(args.ImageDir)
	models.PrepareConns(root)
	config.SetupLog(root.Log)
	app = NewApp(root.App.Name, args.ImageDir)
	if args.Verbose {
		fmt.Println("Config file is", args.Config)
		fmt.Println("Image dir is", args.ImageDir)
	}
}

func main() {
	var err error
	runtime.GOMAXPROCS(1)

	if args.Update != nil {
		args.Update.Run()
		return
	}

	addr := args.GetServerAddr()
	greeting := fmt.Sprintf("Server is start at %s ...", addr)
	logging.Info(greeting)
	if err = app.Listen(addr); err != nil {
		panic(err)
	}
}
