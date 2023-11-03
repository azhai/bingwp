package handlers

import (
	"fmt"
	"strings"
	"time"

	db "github.com/azhai/bingwp/models/default"
	xutils "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
	"github.com/k0kubun/pp"
	"github.com/parnurzeal/gorequest"
)

const (
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edg/90.0.818.66"
	ArchiveUrl    = "https://cn.bing.com/HPImageArchive.aspx?&format=js&mkt=zh-CN&idx=0&n=8&uhd=1&uhdwidth=3840&uhdheight=2160"
	ListUrl       = "https://api.wilii.cn/api/bing?page=%d&pageSize=16"
	DetailUrl     = "https://api.wilii.cn/api/Bing/%d"
	FullUrlPrefix = "https://bing.wilii.cn/OneDrive/bingimages/"
	BaseUrlPrefix = "/th?id=OHR."
	BingThumbUrl  = "https://s.cn.bing.net/th?id=OHR."
	dateFirst     = "20090713"
	dateZero      = "20081231"
)

type ArchiveDict struct {
	Date     string `json:"enddate"`
	Title    string `json:"copyright"`
	FilePath string `json:"urlbase"`
}

type ArchiveResult struct {
	Images []ArchiveDict `json:"images"`
}

type ListDict struct {
	Id       int    `json:"id"`
	Date     string `json:"date"`
	FilePath string `json:"filepath"`
	Title    string `json:"title"`
	Headline string `json:"headline"`
}

type ListData struct {
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Data     []ListDict `json:"data"`
}

type ListResult struct {
	Status   int      `json:"status"`
	Success  bool     `json:"success"`
	Response ListData `json:"response"`
}

type DetailDict struct {
	Date          string `json:"date"`
	FilePath      string `json:"filepath"`
	Title         string `json:"title"`
	Headline      string `json:"headline"`
	TitleEn       string `json:"titleEn"`
	HeadlineEn    string `json:"headlineEn"`
	Description   string `json:"description"`
	DescriptionEn string `json:"descriptionEn"`
	QuickFact     string `json:"quickFact"`
	QuickFactEn   string `json:"quickFactEn"`
	Keyword       string `json:"keyword"`
	Longitude     string `json:"longitude"`
	Latitude      string `json:"latitude"`
}

type DetailResult struct {
	Status   int        `json:"status"`
	Success  bool       `json:"success"`
	Response DetailDict `json:"response"`
}

func FetchRecent() (err error) {
	var stopDate time.Time
	wp, ok := new(db.WallDaily), false
	ok, err = wp.Load(xq.WithOrderBy("bing_date", true))
	if ok && err == nil && wp.Id > 0 {
		stopDate = wp.BingDate
		_, err = UpdateDailyImages(wp) // 看最后一天的图片是否需要重新下载
		err = ReadDetail(wp)
	}
	err = ReadArchive(stopDate.Format("20060102"))
	return
}

func ReadArchive(stopYmd string) (err error) {
	_, body, errs := CreateSpider().Get(ArchiveUrl).End()
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	data := new(ArchiveResult)
	_, err = xutils.UnmarshalJSON([]byte(body), &data)
	if err != nil {
		return
	}
	zeroUnix := MustParseDate(dateZero).Unix()
	items, rows := data.Images, make([]any, 0)
	pp.Println(items)
	var dims string
	for _, card := range items {
		wp := &db.WallDaily{MaxDpi: "400x240"}
		wp.BingDate = MustParseDate(card.Date)
		if wp.BingDate.Format("20060102") <= stopYmd {
			break
		}
		wp.BingSku = GetSkuFromBaseUrl(card.FilePath)
		wp.Title = ParseDailyTitle(card.Title)
		wp.Id = int((wp.BingDate.Unix() - zeroUnix) / 86400)
		dims, err = UpdateDailyImages(wp)
		if dims != "" {
			wp.MaxDpi = dims
		}
		err = ReadDetail(wp)
		rows = append(rows, wp)
	}
	if len(rows) > 0 {
		table := (db.WallDaily{}).TableName()
		err = db.InsertBatch(table, rows...)
	}
	return
}

func ReadList(page int) (err error) {
	url := fmt.Sprintf(ListUrl, page)
	_, body, errs := CreateSpider().Get(url).End()
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	data := new(ListResult)
	_, err = xutils.UnmarshalJSON([]byte(body), &data)
	if err != nil {
		return
	}
	zeroUnix := MustParseDate(dateZero).Unix()
	items, rows := data.Response.Data, make([]any, 0)
	pp.Println(items)

	var dims string
	for _, card := range items {
		wp := &db.WallDaily{OrigId: card.Id, MaxDpi: "400x240"}
		date := strings.ReplaceAll(card.Date, "-", "")
		wp.BingDate = MustParseDate(date)
		wp.BingSku = GetSkuFromFullUrl(card.FilePath)
		wp.Title = ParseDailyTitle(card.Title)
		wp.Id = int((wp.BingDate.Unix() - zeroUnix) / 86400)
		dims, err = UpdateDailyImages(wp)
		if dims != "" {
			wp.MaxDpi = dims
		}
		err = ReadDetail(wp)
		rows = append(rows, wp)
	}
	if len(rows) > 0 {
		table := (db.WallDaily{}).TableName()
		err = db.InsertBatch(table, rows...)
	}
	return
}

func FetchDetails() (err error) {
	note, maxId := new(db.WallNote), 0
	_, err = note.Load(xq.WithOrderBy("daily_id", true))
	if note.Id > 0 {
		maxId = note.DailyId
	}
	where := xq.WithWhere("id > ?", maxId)
	var rows []*db.WallDaily
	err = db.Query(where).Desc("id").Find(&rows)
	for _, wp := range rows {
		err = ReadDetail(wp)
	}
	return
}

func ReadDetail(row *db.WallDaily) (err error) {
	url := fmt.Sprintf(DetailUrl, row.OrigId)
	_, body, errs := CreateSpider().Get(url).End()
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	data := new(DetailResult)
	_, err = xutils.UnmarshalJSON([]byte(body), &data)
	if err != nil {
		return
	}
	dict, notes := data.Response, make([]any, 0)
	if dict.Keyword != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "keyword"}
		note.NoteChinese = xq.NewNullString(dict.Keyword)
		notes = append(notes, note)
	}
	if dict.Headline != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "headline"}
		note.NoteChinese = xq.NewNullString(dict.Headline)
		if dict.HeadlineEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.HeadlineEn)
		}
		notes = append(notes, note)
	}
	if dict.QuickFact != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "quick_fact"}
		note.NoteChinese = xq.NewNullString(dict.QuickFact)
		if dict.QuickFactEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.QuickFactEn)
		}
		notes = append(notes, note)
	}
	if dict.Title != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "title"}
		note.NoteChinese = xq.NewNullString(dict.Title)
		if dict.TitleEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.TitleEn)
		}
		notes = append(notes, note)
	}
	if dict.Description != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "paragraph"}
		note.NoteChinese = xq.NewNullString(dict.Description)
		if dict.DescriptionEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.DescriptionEn)
		}
		notes = append(notes, note)
	}
	if len(notes) > 0 {
		table := (db.WallNote{}).TableName()
		err = db.InsertBatch(table, notes...)
	}
	return
}

// CreateSpider 创建cURL客户端
func CreateSpider() *gorequest.SuperAgent {
	client := gorequest.New().Set("User-Agent", UserAgent)
	// if logger := cmd.GetDefaultLogger(); logger != nil {
	// 	curlLogger := &utils.CurlLogger{Logger: logger}
	// 	client = client.SetDebug(true).SetLogger(curlLogger)
	// }
	return client.Clone()
}

// MustParseDate 解析8位数字表示的日期
func MustParseDate(date string) time.Time {
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
