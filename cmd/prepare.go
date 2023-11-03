package cmd

import (
	"flag"
	"fmt"

	"github.com/azhai/xgen/config"
	"github.com/k0kubun/pp"
)

var (
	serverHost string // 运行IP
	serverPort int    // 运行端口
	updateData bool   // 更新数据
)

type OptionConfig struct {
	Host       string `hcl:"host,optional" json:"host,omitempty"`
	Port       int    `hcl:"port" json:"port"`
	UpdateData bool
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
	fmt.Printf("err=%#v\n%#v\n\n", err, settings)
	if err != nil {
		// panic(err)
	}

	if serverHost != "" {
		options.Host = serverHost
	}
	if serverPort > 0 {
		options.Port = serverPort
	}
	options.UpdateData = updateData
	if config.Verbose() {
		pp.Println(options)
	}
	return options, settings
}
