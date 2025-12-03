package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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
	args.MergeConfigs(env)
	if args.Verbose {
		fmt.Println("Cert dir is", args.CertDir)
		fmt.Println("Image dir is", args.ImageDir)
	}
	handlers.SetImageSaveDir(args.ImageDir)
}

func main() {
	var err error
	if err = log.OpenService(env); err != nil {
		panic(err)
	}
	defer log.CloseService()
	if err = db.OpenService(env); err != nil {
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
	if args.CertDir != "" {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		certPath := filepath.Join(args.CertDir, "cert.pem")
		keyPath := filepath.Join(args.CertDir, "key.pem")
		err = server.ListenAndServeTLS(certPath, keyPath)
	} else {
		err = server.ListenAndServe()
	}
	// if err := SeamLessListen(server, timeout); err != nil {
	// 	logutil.Error(err)
	// }
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		logutil.Error(err)
		os.Exit(1)
	}
}
