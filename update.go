package main

import (
	"github.com/azhai/allgo/logutil"
	"github.com/azhai/bingwp/handlers"
)

// UpdateCmd 更新数据
type UpdateCmd struct {
}

// Run 下载图像并记录到数据库
func (c *UpdateCmd) Run() {
	// 从微软Bing下载最新的图像，以及标题
	var err error
	num, crawler := 0, handlers.NewCrawler()
	if num, err = crawler.SaveArchive(0, ""); err != nil {
		logutil.Error(err)
	}

	// 从wilii.cn读取guid等信息
	if num < 2 {
		num = 2
	}
	if err = handlers.SaveListPages(1, num, false); err != nil {
		logutil.Error(err)
	}

	// 从详情中读取正文等内容
	if err = handlers.SaveSomeDetails(5, 1); err != nil {
		logutil.Error(err)
	}
}
