package tests

import (
	"testing"

	"github.com/azhai/bingwp/handlers"
	"github.com/stretchr/testify/assert"
)

func init() {
	handlers.SetImageSaveDir("../data")
}

// TestSaveArchive 从微软Bing下载最新的图像，以及标题
func TestSaveArchive(t *testing.T) {
	crawler := handlers.NewCrawler()
	_, err := crawler.SaveArchive(0, "")
	assert.NoError(t, err)
}

// TestSaveListPages 从wilii.cn读取guid等信息
func TestSaveListPages(t *testing.T) {
	err := handlers.SaveListPages(1, 2, true)
	assert.NoError(t, err)
}

// TestSaveSomeDetails 从详情中读取正文等内容
func TestSaveSomeDetails(t *testing.T) {
	err := handlers.SaveSomeDetails(5, 1)
	assert.NoError(t, err)
}
