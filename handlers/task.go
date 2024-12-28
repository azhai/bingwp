package handlers

import (
	"fmt"
	"os"
	"time"

	xutils "github.com/azhai/xgen/utils"
)

const (
	SaveFileNameFormat = "./tmp/list-s%d-p%03d.json"
)

// SaveListPages 保存列表页面
func SaveListPages(size int) (err error) {
	var (
		data *ListResult
		body []byte
	)
	i, crawler := 1, NewCrawler()
	if data, err = crawler.CrawlList(i, size); err != nil {
		return
	}
	maxPage := data.Response.PageCount
	for i = 1; i <= maxPage; i++ {
		url := fmt.Sprintf(ListUrl, i, size)
		if body, err = crawler.Crawl(url); err != nil {
			continue
		}
		path := fmt.Sprintf(SaveFileNameFormat, size, i)
		_ = os.WriteFile(path, body, 0644)
		time.Sleep(10 * time.Millisecond)
	}
	return
}

func UpdateDailyHeadline(size, maxPage int) (err error) {
	var (
		data *ListResult
		body []byte
	)
	for i := 1; i <= maxPage; i++ {
		path := fmt.Sprintf(SaveFileNameFormat, size, i)
		if body, err = os.ReadFile(path); err != nil || body == nil {
			continue
		}
		if _, err = xutils.UnmarshalJSON(body, &data); err != nil {
			continue
		}
		for _, card := range data.Response.Data {
			if wp := CreateDailyModel(card); wp != nil {
				err = wp.Save(map[string]any{"headline": wp.Headline})
			}
		}
	}
	return
}
