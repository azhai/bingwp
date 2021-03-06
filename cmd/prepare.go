package cmd

import (
	"flag"

	"github.com/azhai/xgen/config"
	xutils "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
	"github.com/k0kubun/pp"
)

var (
	serverHost string // 运行IP
	serverPort int    // 运行端口
	updateData bool   // 更新数据
	logger     *xutils.Logger
)

type OptionConfig struct {
	Host         string `hcl:"host,optional" json:"host,omitempty"`
	Port         int    `hcl:"port" json:"port"`
	MaxWriteSize int
	UpdateData   bool
}

func init() {
	flag.StringVar(&serverHost, "s", "", "运行IP")
	flag.IntVar(&serverPort, "p", 9870, "运行端口")
	flag.BoolVar(&updateData, "u", false, "更新数据")
	config.Setup()
}

func GetOptions() (*OptionConfig, *config.RootConfig) {
	options := new(OptionConfig)
	settings, err := config.ReadConfigFile(options)
	if err != nil {
		panic(err)
	}
	logger, err = config.GetConfigLogger()
	if err != nil {
		panic(err)
	}

	if serverHost != "" {
		options.Host = serverHost
	}
	if serverPort > 0 {
		options.Port = serverPort
	}
	options.UpdateData = updateData
	options.MaxWriteSize = xq.MaxWriteSize
	if config.Verbose() {
		pp.Println(options)
	}
	return options, settings
}

func GetDefaultLogger() *xutils.Logger {
	return logger
}
