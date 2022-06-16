package main

// https://bing.wilii.cn/ymd.asp?ismobile=0&y=2022&m=5
// https://bing.wilii.cn/OneDrive/bingimages/2022/05/01/VanBlooms_ZH-CN6370306779_400x240.jpg
// https://api.wilii.cn/bing/binglist.ashx?DictType=Detail&id=5674
// https://bing.wilii.cn/download.asp?filename=/OneDrive/bingimages/2022/05/01/VanBlooms_ZH-CN6370306779_UHD.jpg&name=%E7%9B%9B%E5%BC%80%E7%9A%84%E9%87%91%E9%93%BE%E8%8A%B1%E6%A0%91%E5%92%8C%E7%B4%AB%E8%89%B2%E8%91%B1%E5%B1%9E%E6%A4%8D%E7%89%A9%EF%BC%8C%E5%8A%A0%E6%8B%BF%E5%A4%A7%E6%B8%A9%E5%93%A5%E5%8D%8E%E8%8C%83%E5%BA%A6%E6%A3%AE%E6%A4%8D%E7%89%A9%E5%9B%AD_UHD.jpg&id=5674

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/azhai/bingwp/cmd"
	db "github.com/azhai/bingwp/models/default"
	"github.com/azhai/bingwp/utils"

	"github.com/PuerkitoBio/goquery"
	xutils "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
	"github.com/parnurzeal/gorequest"
)

const (
	UserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edg/90.0.818.66"
	SiteListUrl  = "https://bing.wilii.cn/ymd.asp?ismobile=0"
	SiteNoteUrl  = "https://api.wilii.cn/bing/binglist.ashx?DictType=Detail"
	SiteThumbUrl = "https://bing.wilii.cn/OneDrive/bingimages"
	BingThumbUrl = "https://s.cn.bing.net/th?id=OHR."
	IdLinkPrefix = "photo.html?id="
	ImageDataDir = "/data/bingwp/"
	dateFirst    = "2009-07-13"
	dateZero     = "2008-12-31"
)

type WallDict struct {
	Longitude string           `json:"map_longitude"`
	Latitude  string           `json:"map_latitude"`
	Keyword   string           `json:"keyword"`
	Headline  string           `json:"Headline"`
	QuickFact string           `json:"quickFact"`
	Title     string           `json:"map_title"`
	TitleE    string           `json:"map_titleE"`
	Para1     string           `json:"map_para1"`
	Para1E    string           `json:"map_para1E"`
	Tags      []map[string]any `json:"tags"`
}

// FetchWallPapers 获取Bing背景图列表
func FetchWallPapers() (err error) {
	stopId, stopDate := 0, MustParseDate(dateFirst)
	row := new(db.WallDaily)
	var ok bool
	ok, err = row.Load(xq.WithOrderBy("bing_date", true))
	if ok && err == nil {
		stopId, stopDate = row.OrigId, row.BingDate
	}
	stopUnix := GetMonthBegin(stopDate).Unix()
	tempDate := GetMonthBegin(time.Now())
	for tempDate.Unix() >= stopUnix {
		date := tempDate.Format("2006-01-02")
		_, listUrl := GetMonthDirAndUrl(date)
		fmt.Println(listUrl)
		err = FetchList(listUrl, stopId)
		if err != nil {
			panic(err)
		}
		tempDate = tempDate.AddDate(0, -1, 0)
		time.Sleep(200 * time.Millisecond)
	}
	return
}

func FetchList(url string, stopId int) (err error) {
	resp, _, errs := CreateSpider().Get(url).End()
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil || doc == nil {
		return
	}
	sect := doc.Find("section.bg-section-secondary").First()
	if sect == nil {
		return
	}

	zeroUnix := MustParseDate(dateZero).Unix()
	linkSize, rows := len(IdLinkPrefix), []any{}
	var cards []*goquery.Selection
	sect.Find("div.card").Each(func(i int, card *goquery.Selection) {
		cards = append(cards, card)
	})
	for _, card := range cards {
		wp := &db.WallDaily{MaxDpi: "400x240"}
		imgSrc := card.Find("img").First().AttrOr("src", "")
		sku := filepath.Base(imgSrc)
		if strings.HasSuffix(sku, ".jpg") {
			sku = sku[:len(sku)-4]
		}
		if strings.HasSuffix(sku, "_400x240") {
			sku = sku[:len(sku)-8]
		}
		wp.BingSku = sku
		dataDiv := card.Find("div.title").First()
		wp.Brief = dataDiv.Find("div.name").First().Text()
		date := dataDiv.Find("div.date").First().Text()
		wp.BingDate = MustParseDate(date)
		wp.Id = int((wp.BingDate.Unix() - zeroUnix) / 86400)
		link := dataDiv.Find("a").First().AttrOr("href", "")
		if strings.HasPrefix(link, IdLinkPrefix) {
			wp.OrigId, _ = strconv.Atoi(link[linkSize:])
		}
		if wp.OrigId <= stopId {
			break
		}
		rows = append(rows, wp)
		err = FetchNotes(wp)
		err = FetchImage(wp)
	}
	if len(rows) > 0 {
		table := (db.WallDaily{}).TableName()
		err = db.InsertBatch(table, rows...)
	}
	return
}

func FetchNotes(row *db.WallDaily) (err error) {
	url := fmt.Sprintf("%s&id=%d", SiteNoteUrl, row.OrigId)
	_, body, errs := CreateSpider().Get(url).End()
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	data := []WallDict{}
	_, err = xutils.UnmarshalJSON([]byte(body), &data)
	if err != nil || len(data) == 0 {
		return
	}
	dict, notes := data[0], make([]any, 0)
	if dict.Longitude != "" {
		loc := &db.WallLocation{DailyId: row.Id}
		loc.Longitude, _ = strconv.ParseFloat(dict.Longitude, 64)
		if dict.Latitude != "" {
			loc.Latitude, _ = strconv.ParseFloat(dict.Latitude, 64)
		}
		table := (db.WallLocation{}).TableName()
		err = db.InsertBatch(table, loc)
	}
	if len(dict.Tags) > 0 {
		tags := make([]any, 0)
		for _, oneTag := range dict.Tags {
			if word, ok := oneTag["word"]; ok {
				tag := &db.WallTag{DailyId: row.Id, TagName: word.(string)}
				tags = append(tags, tag)
			}
		}
		table := (db.WallTag{}).TableName()
		err = db.InsertBatch(table, tags...)
	}
	if dict.Keyword != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "keyword"}
		note.NoteChinese = xq.NewNullString(dict.Keyword)
		notes = append(notes, note)
	}
	if dict.Headline != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "headline"}
		note.NoteChinese = xq.NewNullString(dict.Headline)
		notes = append(notes, note)
	}
	if dict.QuickFact != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "quick_fact"}
		note.NoteChinese = xq.NewNullString(dict.QuickFact)
		notes = append(notes, note)
	}
	if dict.Title != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "title"}
		note.NoteChinese = xq.NewNullString(dict.Title)
		if dict.TitleE != "" {
			note.NoteEnglish = xq.NewNullString(dict.TitleE)
		}
		notes = append(notes, note)
	}
	if dict.Para1 != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "paragraph"}
		note.NoteChinese = xq.NewNullString(dict.Para1)
		if dict.Para1E != "" {
			note.NoteEnglish = xq.NewNullString(dict.Para1E)
		}
		notes = append(notes, note)
	}
	if len(notes) > 0 {
		table := (db.WallNote{}).TableName()
		err = db.InsertBatch(table, notes...)
	}
	return
}

func GetMonthDirAndUrl(date string) (saveDir, listUrl string) {
	year, month := date[:4], date[5:7]
	saveDir = filepath.Join(ImageDataDir, year+month)
	month = strings.TrimLeft(month, "0")
	listUrl = fmt.Sprintf("%s&y=%s&m=%s", SiteListUrl, year, month)
	return
}

// CreateSpider 创建cURL客户端
func CreateSpider() *gorequest.SuperAgent {
	client := gorequest.New().Set("User-Agent", UserAgent)
	if logger := cmd.GetDefaultLogger(); logger != nil {
		curlLogger := &utils.CurlLogger{Logger: logger}
		client = client.SetDebug(true).SetLogger(curlLogger)
	}
	return client.Clone()
}

func MustParseDate(date string) time.Time {
	obj, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return obj
}

func GetMonthBegin(obj time.Time) time.Time {
	return obj.AddDate(0, 0, 1-obj.Day())
}
