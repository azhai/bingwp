package handlers

import (
	"strings"
	"time"

	db "github.com/azhai/bingwp/models/default"
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
	if strings.HasPrefix(card.FilePath, FullUrlPrefix) {
		wp.BingSku = GetSkuFromFullUrl(card.FilePath)
	} else {
		wp.BingSku = GetSkuFromBaseUrl(card.FilePath)
	}
	wp.Title = ParseDailyTitle(card.Title)
	return wp
}

// InsertDailyRows 写入多行 Daily
func InsertDailyRows(items []DailyDict) (num int, err error) {
	var wp *db.WallDaily
	for _, card := range items {
		if wp = CreateDailyModel(card); wp == nil {
			continue
		}
		if err = wp.Save(nil); err == nil {
			num += 1
		}
	}
	return
}

// InsertBatchDailyRows 写入多行 Daily
func InsertBatchDailyRows(items []DailyDict) (num int, err error) {
	var rows []any
	for _, card := range items {
		if wp := CreateDailyModel(card); wp != nil {
			rows = append(rows, wp)
		}
	}
	if num = len(rows); num > 0 {
		err = db.InsertBatch(dailyTable, rows...)
	}
	return
}
