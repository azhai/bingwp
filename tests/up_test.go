package tests

import (
	"testing"
	"time"

	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/services/db"
	"github.com/stretchr/testify/assert"
)

var (
	resetSqls = []string{
		"DELETE FROM t_wall_daily WHERE id >= $1",
		"DELETE FROM t_wall_image WHERE daily_id >= $1",
		"DELETE FROM t_wall_note WHERE daily_id >= $1",
		"SELECT setval('t_wall_note_id_seq', (SELECT max(id) FROM t_wall_note WHERE daily_id < $1))",
	}
)

// ServerOpts 服务配置
type ServerOpts struct {
	Host     string `arg:"-s,--host" default:"" help:"运行IP"`              // 运行IP
	Port     int    `arg:"-p,--port" default:"9870" help:"运行端口"`          // 运行端口
	ImageDir string `arg:"-d,--dir" help:"图片目录" hcl:"image_dir,optional"` // 图片目录
}

// removeTwoDays 删除最近两天的数据
func removeTwoDays() {
	db := db.New()
	id := handlers.GetOffsetDay(time.Now()) - 1
	for _, query := range resetSqls {
		db.Exec(query, id)
	}
}

// TestSaveArchive 从微软Bing下载最新的图像，以及标题
func TestSaveArchive(t *testing.T) {
	crawler := handlers.NewCrawler()
	_, err := crawler.SaveArchive(0, "")
	assert.NoError(t, err)
}

// TestSaveListPages 从wilii.cn读取guid等信息
func TestSaveListPages(t *testing.T) {
	removeTwoDays()
	err := handlers.SaveListPages(1, 2, true)
	assert.NoError(t, err)
}

// TestSaveSomeDetails 从详情中读取正文等内容
func TestSaveSomeDetails(t *testing.T) {
	err := handlers.SaveSomeDetails(5, 1)
	assert.NoError(t, err)
}
