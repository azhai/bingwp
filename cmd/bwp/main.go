package main

import (
	"fmt"
	"runtime"

	"github.com/alexflint/go-arg"
	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/gozzo/config"
	"github.com/azhai/gozzo/logging"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
)

var args struct {
	Update   *UpdateCmd `arg:"subcommand:up" help:"更新数据"`
	Config   string     `arg:"-c,--config" default:"settings.hcl" help:"配置文件路径"`
	Verbose  bool       `arg:"-v,--verbose" help:"输出详细信息"`
	ImageDir string     `hcl:"image_dir,optional"` // 图片目录
	ServerOpts
}

type ServerOpts struct {
	Host string `arg:"-s,--host" default:"" help:"运行IP"`     // 运行IP
	Port int    `arg:"-p,--port" default:"9870" help:"运行端口"` // 运行端口
}

func (t ServerOpts) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

type UpdateCmd struct {
}

func (c *UpdateCmd) Run() {
	crawler := handlers.NewCrawler()
	if _, err := crawler.SavelArchive(0, ""); err != nil {
		logging.Error(err)
	}
	// handlers.SaveListPages(100)
	// handlers.UpdateDailyHeadline(100, 57)
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
	config.ReadConfigFile(args.Config, args.Verbose, &args)
	config.SetupLog()
}

func main() {
	var err error
	runtime.GOMAXPROCS(1)

	if args.ImageDir != "" {
		handlers.SetImageSaveDir(args.ImageDir)
	}
	if args.Update != nil {
		args.Update.Run()
		return
	}

	addr := args.GetServerAddr()
	greeting := fmt.Sprintf("Server is start at %s ...", addr)
	logging.Info(greeting)
	cfg := config.GetAppSettings()
	app := NewApp(cfg.Name, args.ImageDir)
	if err = app.Listen(addr); err != nil {
		panic(err)
	}
}
