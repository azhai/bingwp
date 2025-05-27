package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/azhai/allgo/logutil"
	"github.com/azhai/bingwp/handlers"
	"github.com/kataras/compress"
	"github.com/kavu/go_reuseport"
	"github.com/rs/cors"
	"github.com/rs/seamless"
)

var (
	pidFile = "/tmp/bingwp.pid"
	timeout = 60 * time.Second
)

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

// NewApp 创建http服务
func NewApp(imgDir string) http.Handler {
	mux := http.NewServeMux()

	// add routes
	mux.Handle("/favicon.ico", http.RedirectHandler(
		"/static/logo-small.svg", 301,
	))
	mux.Handle("/static/", http.FileServer(http.Dir("./")))
	mux.Handle("/wallpaper/", http.StripPrefix(
		"/wallpaper/", http.FileServer(http.Dir(imgDir)),
	))
	mux.HandleFunc("/{month}", handlers.PageHandler)
	mux.HandleFunc("/", handlers.PageHandler)

	// wrap middlewares
	app := cors.Default().Handler(mux)
	app = compress.Handler(mux)

	return app
}

func SeamLess(app http.Handler, addr string, timeout time.Duration) error {
	seamless.Init(pidFile)
	listener, err := reuseport.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: addr, Handler: app}

	var errChan = make(chan error, 1)
	// Implement the graceful shutdown that will be triggered once the new process
	// successfully rebound the socket.
	seamless.OnShutdown(func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err = server.Shutdown(ctx); err != nil {
			logutil.Info("Graceful shutdown timeout, force closing")
			errChan <- server.Close()
		}
	})

	go func() {
		// Give the server a second to start
		time.Sleep(time.Second)
		if err == nil {
			seamless.Started()
			errChan <- err
		}
	}()
	err = server.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
	}
	seamless.Wait()
	return <-errChan
}
