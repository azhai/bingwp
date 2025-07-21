package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/logutil"
	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/services/db"
	"github.com/azhai/bingwp/services/log"
)

var (
	env *config.Environ

	appName    = "Bing Wallpaper"
	appVersion = "0.0.0"
)

func GetAppName() string {
	appName = env.GetStr("APP_NAME", appName)
	appVersion = env.GetStr("APP_VERSION", appVersion)
	if appVersion != "" && appVersion != "0.0.0" {
		return fmt.Sprintf("%s v%s", appName, appVersion)
	}
	return appName
}

var args struct {
	Update  *UpdateCmd `arg:"subcommand:up" help:"更新数据"`
	Verbose bool       `arg:"-v,--verbose" help:"输出详细信息"`
	ServerOpts
}

func init() {
	env = config.Prepare(256)
	arg.MustParse(&args)
	if args.ImageDir == "" {
		args.ImageDir = env.Get("IMAGE_DIR")
	} else {
		_ = os.Setenv("IMAGE_DIR", args.ImageDir)
	}
	if args.Verbose {
		fmt.Println("Image dir is", args.ImageDir)
	}
	handlers.SetImageSaveDir(args.ImageDir)
}

func main() {
	if err := log.OpenService(env); err != nil {
		panic(err)
	}
	defer log.CloseService()
	if err := db.OpenService(env); err != nil {
		panic(err)
	}
	defer db.CloseService()

	if args.Update != nil {
		args.Update.Run()
		return
	}

	name, addr := GetAppName(), args.GetServerAddr()
	greeting := fmt.Sprintf("Server %s start at %s ...", name, addr)
	logutil.Info(greeting)
	app := NewApp(args.ImageDir)
	server := &http.Server{Addr: addr, Handler: app}
	if err := server.ListenAndServe(); err != nil {
		logutil.Error(err)
	}
	// if err := SeamLessListen(server, timeout); err != nil {
	// 	logutil.Error(err)
	// }
}
