package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	db "github.com/azhai/bingwp/models/default"
	xutils "github.com/azhai/xgen/utils"
)

const (
	SaveDetailFileName = "./tmp/wp-%s.json"
)

// SaveListPages 保存列表页面
func SaveListPages(pageCount int, pageSize int, getDetail bool) (err error) {
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
					"guid":     strings.TrimSpace(wp.Guid),
					"headline": strings.TrimSpace(wp.Headline),
					"color":    strings.TrimSpace(wp.Color),
				}
				err = wp.Save(changes)
				if err == nil && getDetail {
					err = UpdateDailyDetail(wp, true)
				}
			}
		}

		time.Sleep(10 * time.Millisecond)
	}
	return
}

func SaveSomeDetails(limit, start int) (err error) {
	var rows []*db.WallDaily
	qr := db.Query().Limit(limit, start).Desc("id")
	if err = qr.Find(&rows); err != nil {
		fmt.Println(err)
	}
	for i := len(rows) - 1; i >= 0; i-- {
		err = UpdateDailyDetail(rows[i], true)
		if err != nil {
			fmt.Println(err)
		}
	}
	return
}

func UpdateDailyDetail(wp *db.WallDaily, override bool) (err error) {
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
	if override {
		data = crawler.CrawlDetail(wp.Guid)
		time.Sleep(5 * time.Millisecond)
	} else if body, err = os.ReadFile(path); err != nil || body == nil {
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
