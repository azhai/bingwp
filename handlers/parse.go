package handlers

import (
	"strings"
	"time"

	db "github.com/azhai/bingwp/models/default"
	xq "github.com/azhai/xgen/xquery"
)

const (
	dateFirst = "20090713"
	dateZero  = "20081231"
)

var (
	zeroUnix   = MustParseDate(dateZero).Unix()
	dailyTable = new(db.WallDaily).TableName()
)

// GetOffsetDay 以2009年元旦为第一天，计算当前是多少天
func GetOffsetDay(dt time.Time) int {
	return int((dt.Unix() - zeroUnix) / 86400)
}

// MustParseDate 解析8位数字表示的日期
func MustParseDate(date string) time.Time {
	date = strings.ReplaceAll(date, "-", "")
	obj, err := time.Parse("20060102", date)
	if err != nil {
		panic(err)
	}
	return obj
}

// ParseDailyTitle 解析标题，去掉最后括号中的版权内容
func ParseDailyTitle(title string) string {
	idx := strings.LastIndex(title, "(")
	if idx > 0 && strings.HasSuffix(title, ")") {
		return title[:idx]
	}
	return title
}

// GetSkuFromBaseUrl 从基础网址中提取壁纸的SKU
func GetSkuFromBaseUrl(url string) string {
	size := len(BaseUrlPrefix)
	if strings.HasPrefix(url, BaseUrlPrefix) {
		url = url[size:]
	}
	return strings.TrimSpace(url)
}

// GetSkuFromFullUrl 从完整网址中提取壁纸的SKU
func GetSkuFromFullUrl(url string) string {
	size := len(FullUrlPrefix) + len("2006/01/02/")
	if strings.HasPrefix(url, FullUrlPrefix) {
		url = url[size:]
	}
	if pieces := strings.Split(url, "_"); len(pieces) > 1 {
		if strings.HasPrefix(pieces[1], "ZH-CN") {
			url = pieces[0] + "_" + pieces[1]
		}
	}
	return strings.TrimSpace(url)
}

// CreateDailyModel 创建一行 Daily
func CreateDailyModel(card DailyDict) *db.WallDaily {
	wp := &db.WallDaily{MaxDpi: "400x240"}
	wp.BingDate = MustParseDate(card.Date)
	wp.Id = GetOffsetDay(wp.BingDate)
	wp.Guid, wp.Color = card.Guid, card.Color
	if strings.HasPrefix(card.FilePath, FullUrlPrefix) {
		wp.BingSku = GetSkuFromFullUrl(card.FilePath)
	} else {
		wp.BingSku = GetSkuFromBaseUrl(card.FilePath)
	}
	wp.Title = ParseDailyTitle(card.Title)
	wp.Headline = strings.TrimSpace(card.Headline)
	return wp
}

// InsertNotExistDailyRows 写入多行 Daily ，但先要排除掉已存在的行
func InsertNotExistDailyRows(items []DailyDict, withImages bool) (num int, err error) {
	var (
		dailyRows, existRows []*db.WallDaily
		dates, rows          []any
	)
	dict := make(map[string]int)
	for _, card := range items {
		if wp := CreateDailyModel(card); wp != nil {
			bingDate := wp.BingDate.Format("2006-01-02")
			dates = append(dates, bingDate)
			dict[bingDate] = 0
			dailyRows = append(dailyRows, wp)
		}
	}
	where := xq.WithRange("bing_date", dates...)
	err = db.Query(where).Asc("bing_date").Find(&existRows)
	if err == nil {
		for _, wp := range existRows {
			bingDate := wp.BingDate.Format("2006-01-02")
			dict[bingDate] = wp.Id
		}
	}
	var dims string
	for _, wp := range dailyRows {
		bingDate := wp.BingDate.Format("2006-01-02")
		if id, ok := dict[bingDate]; !ok || id == 0 {
			if withImages {
				if dims, err = UpdateDailyImages(wp); dims != "" {
					wp.MaxDpi = dims
				}
			}
			rows = append(rows, wp)
		}
	}
	if num = len(rows); num > 0 {
		err = db.InsertBatch(dailyTable, rows...)
	}
	return
}
