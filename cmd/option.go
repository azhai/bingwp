package cmd

import (
	"flag"
	"fmt"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/gozzo/config"
)

var theOptions *OptionConfig

type OptionConfig struct {
	Host       string `hcl:"host,optional" json:"host,omitempty"`           // 运行IP
	Port       int    `hcl:"port,optional" json:"port,omitempty"`           // 运行端口
	ImageDir   string `hcl:"image_dir,optional" json:"image_dir,omitempty"` // 图片目录
	UpdateData bool   // 更新数据
}

func (c *OptionConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetAppOptions 返回命令行选项
func GetAppOptions() (string, *OptionConfig) {
	cfg := config.GetAppSettings()
	return cfg.Name, theOptions
}

func init() {
	theOptions = new(OptionConfig)
	flag.StringVar(&theOptions.Host, "s", "", "运行IP")
	flag.IntVar(&theOptions.Port, "p", 9870, "运行端口")
	flag.BoolVar(&theOptions.UpdateData, "u", false, "更新数据")
	flag.Parse()

	config.SetupEnv(theOptions)
	config.SetupLog()
	models.SetupDb()
}
