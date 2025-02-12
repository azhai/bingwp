package handlers

import (
	"fmt"
	"os"
	"time"

	db "github.com/azhai/bingwp/models/default"
	xutils "github.com/azhai/xgen/utils"
)

const (
	SaveDetailFileName = "./tmp/wp-%s.json"
)

// SaveListPages 保存列表页面
func SaveListPages(pageCount int, pageSize int) (err error) {
	var (
		result *ListResult
		body   []byte
	)
	i, crawler := 1, NewCrawler()
	if result, err = crawler.CrawlList(i, pageSize); err != nil {
		return
	}
	if pageCount < 0 {
		pageCount = result.Response.PageCount
	}
	for i = 1; i <= pageCount; i++ {
		url := fmt.Sprintf(ListUrl, i, pageSize)
		if body, err = crawler.Crawl(url); err != nil {
			continue
		}

		if _, err = xutils.UnmarshalJSON(body, &result); err != nil {
			fmt.Println(err)
			continue
		}
		for _, card := range result.Response.Data {
			if wp := CreateDailyModel(card); wp != nil {
				changes := map[string]any{
					"guid":     wp.Guid,
					"headline": wp.Headline,
					"color":    wp.Color,
				}
				if err = wp.Save(changes); err == nil {
					err = UpdateDailyDetail(wp)
				}
			}
		}

		time.Sleep(10 * time.Millisecond)
	}
	return
}

func UpdateDailyDetail(wp *db.WallDaily) (err error) {
	var (
		result  *DetailResult
		data    *DetailDict
		body    []byte
		crawler = NewCrawler()
	)
	if wp.Guid == "" {
		return
	}
	path := fmt.Sprintf(SaveDetailFileName, wp.Guid)
	if body, err = os.ReadFile(path); err != nil || body == nil {
		data = crawler.CrawlDetail(wp.Guid)
		time.Sleep(5 * time.Millisecond)
	} else if _, err = xutils.UnmarshalJSON(body, &result); err != nil {
		data = crawler.CrawlDetail(wp.Guid)
		time.Sleep(5 * time.Millisecond)
	} else {
		data = result.Response
	}
	err = WriteDetail(wp, data)
	return
}
