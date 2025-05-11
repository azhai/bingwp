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
)

var (
	appName    = "Bing Wallpaper"
	appVersion = "0.0.0"
)

func getAppName() string {
	if appVersion != "" && appVersion != "0.0.0" {
		return fmt.Sprintf("%s v%s", appName, appVersion)
	}
	return appName
}

var args struct {
	Update  *UpdateCmd `arg:"subcommand:up" help:"更新数据"`
	Config  string     `arg:"-c,--config" default:"settings.hcl" help:"配置文件路径"`
	Verbose bool       `arg:"-v,--verbose" help:"输出详细信息"`
	ServerOpts
}

// ServerOpts 服务配置
type ServerOpts struct {
	Host     string `arg:"-s,--host" default:"" help:"运行IP"`              // 运行IP
	Port     int    `arg:"-p,--port" default:"9870" help:"运行端口"`          // 运行端口
	ImageDir string `arg:"-d,--dir" help:"图片目录" hcl:"image_dir,optional"` // 图片目录
}

// GetServerAddr 获取服务地址
func (t ServerOpts) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

// UpdateCmd 更新数据
type UpdateCmd struct {
}

// Run 下载图像并记录到数据库
func (c *UpdateCmd) Run() {
	// 从微软Bing下载最新的图像，以及标题
	var err error
	num, crawler := 0, handlers.NewCrawler()
	if num, err = crawler.SavelArchive(0, ""); err != nil {
		logging.Error(err)
	}

	// 从wilii.cn读取guid等信息
	if num < 2 {
		num = 2
	}
	if err = handlers.SaveListPages(1, num, false); err != nil {
		logging.Error(err)
	}

	// 从详情中读取正文等内容
	if err = handlers.SaveSomeDetails(5, 1); err != nil {
		logging.Error(err)
	}
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
	appName, appVersion = root.App.Name, root.App.Version
	if args.ImageDir == "" {
		err = root.ParseAppRemain(&args.ServerOpts)
		if err != nil {
			panic(err)
		}
	}
	handlers.SetImageSaveDir(args.ImageDir)
	models.PrepareConns(root)
	config.SetupLog(root.Log)
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

	name, addr := getAppName(), args.GetServerAddr()
	greeting := fmt.Sprintf("Server %s is start at %s ...", name, addr)
	logging.Info(greeting)
	app := NewApp(name, args.ImageDir)
	if err = app.Listen(addr); err != nil {
		panic(err)
	}
}
