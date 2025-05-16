package handlers

import (
	"fmt"
	"os"
	"time"

	"github.com/azhai/bingwp/services/database"
	"github.com/goccy/go-json"
)

const (
	SaveDetailFileName = "./tmp/wp-%s.json"
)

// SaveListPages 保存列表页面
func SaveListPages(pageCount int, pageSize int, getDetail bool) (err error) {
	var result *ListResult
	i, crawler := 1, NewCrawler()
	if result, err = crawler.CrawlList(i, pageSize); err != nil {
		return
	}
	if pageCount < 0 {
		pageCount = result.Response.PageCount
	}
	var dailyRows []*database.WallDaily
	for i = 1; i <= pageCount; i++ {
		url := fmt.Sprintf(ListUrl, i, pageSize)
		var body []byte
		if body, err = crawler.Crawl(url); err != nil {
			continue
		}
		if err = json.Unmarshal(body, &result); err != nil {
			fmt.Println(err)
			continue
		}
		if !getDetail {
			continue
		}
		for _, card := range result.Response.Data {
			if wp := CreateWallDaily(card); wp != nil {
				err = UpdateDailyDetail(wp, true)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	_, err = database.InsertBatch(dailyRows)
	return
}

func SaveSomeDetails(limit, start int) (err error) {
	dailyRows := database.GetLatestDailyRows(limit, start)
	for i := len(dailyRows) - 1; i >= 0; i-- {
		err = UpdateDailyDetail(dailyRows[i], true)
		if err != nil {
			fmt.Println(err)
		}
	}
	return
}

func UpdateDailyDetail(wp *database.WallDaily, override bool) (err error) {
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
	} else if err = json.Unmarshal(body, &result); err != nil {
		data = crawler.CrawlDetail(wp.Guid)
		time.Sleep(5 * time.Millisecond)
	} else {
		data = result.Response
	}
	err = WriteDetail(wp, data)
	return
}
