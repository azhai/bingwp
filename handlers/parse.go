package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/azhai/bingwp/services/database"
)

const (
	dateFirst = "20090713"
	dateZero  = "20081231"
)

var zeroUnix = MustParseDate(dateZero).Unix()

// GetOffsetDay 以2009年元旦为第一天，计算当前是多少天
func GetOffsetDay(dt time.Time) int64 {
	return (dt.Unix() - zeroUnix) / 86400
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

// CreateWallDaily 创建一行 Daily
func CreateWallDaily(card DailyDict) *database.WallDaily {
	wp := &database.WallDaily{MaxDpi: "0x0"}
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
	var dates string
	dailyRows := make([]*database.WallDaily, 0)
	for _, card := range items {
		if wp := CreateWallDaily(card); wp != nil {
			bingDate := wp.BingDate.Format("2006-01-02")
			dates += "'" + bingDate + "', "
			dailyRows = append(dailyRows, wp)
		}
	}
	if strings.HasSuffix(dates, ", ") {
		dates = dates[:len(dates)-2]
	}
	num, err = database.InsertDailyRows(dailyRows, dates)
	if err == nil && withImages {
		err = UpdateDailyMaxDPI(dailyRows, true)
	}
	return
}

// UpdateDailyMaxDPI 更新壁纸的图片信息并回写最大分辨率
func UpdateDailyMaxDPI(dailyRows []*database.WallDaily, saveImage bool) (err error) {
	dict := make(map[string][]string)
	for _, wp := range dailyRows {
		var dims, id string
		if saveImage {
			dims, err = SaveDailyImages(wp)
		} else if wp.Image != nil {
			dims = wp.Image.GeDims()
		}
		if err == nil && dims != "" && dims != "0x0" {
			id, wp.MaxDpi = strconv.FormatInt(wp.Id, 10), dims
			dict[dims] = append(dict[dims], id)
		}
	}
	table := new(database.WallDaily).TableName()
	for dims, ids := range dict {
		where := "id IN (" + strings.Join(ids, ", ") + ")"
		changes := map[string]any{"max_dpi": dims}
		_, err = database.ExecUpdate(table, where, nil, changes)
	}
	return
}
