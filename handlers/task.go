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
		for _, card := range result.Response.Data {
			dailyRows = append(dailyRows, CreateWallDaily(card))
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(dailyRows) == 0 {
		return
	}
	_, err = database.InsertDailyRows(dailyRows, nil)
	if err != nil || !getDetail {
		return
	}
	err = UpdateDailyDetails(dailyRows)
	return
}

func SaveSomeDetails(limit, start int) (err error) {
	dailyRows := database.GetLatestDailyRows(limit, start)
	if len(dailyRows) > 0 {
		err = UpdateDailyDetails(dailyRows)
	}
	return
}

func UpdateDailyDetails(dailyRows []*database.WallDaily) (err error) {
	dailyRows = database.GetDailyNotes(dailyRows)
	var (
		data            *DetailDict
		notes, noteRows []*database.WallNote
	)
	for i := len(dailyRows) - 1; i >= 0; i-- {
		wp := dailyRows[i]
		if data, err = GetDailyDetailDict(wp, true); err != nil {
			fmt.Println(err)
			continue
		}
		if notes, err = GetNotExistNotes(wp, data); err != nil {
			fmt.Println(err)
			continue
		}
		if len(notes) > 0 {
			noteRows = append(noteRows, notes...)
		}
	}
	if len(noteRows) > 0 {
		_, err = database.InsertBatch(noteRows)
	}
	return
}

func GetDailyDetailDict(wp *database.WallDaily, override bool) (
	data *DetailDict, err error) {
	if wp.Guid == "" {
		return
	}
	var (
		result *DetailResult
		body   []byte
	)
	crawler := NewCrawler()
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
	return
}
