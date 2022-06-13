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
	xq "github.com/azhai/xgen/xquery"
	"github.com/parnurzeal/gorequest"
)

const (
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edg/90.0.818.66"
	SiteHomeUrl    = "https://bing.wilii.cn"
	SiteListPath   = "/ymd.asp?ismobile=0"
	IdLinkPrefix   = "photo.html?id="
	CurrentPathDir = "images/"
	dateFirst      = "2009-07-01"
	dateZero       = "2008-12-31"
)

// FetchWallPapers 获取Bing背景图列表
func FetchWallPapers() (err error) {
	monthBegin := time.Now().AddDate(0, 0, 1-time.Now().Day())
	date := monthBegin.Format("2006-01-02")
	for date >= dateFirst {
		_, listUrl := GetMonthDirAndUrl(date)
		fmt.Println(listUrl)
		err = FetchList(CreateSpider(), listUrl)
		if err != nil {
			panic(err)
		}
		monthBegin = monthBegin.AddDate(0, -1, 0)
		date = monthBegin.Format("2006-01-02")
		time.Sleep(200 * time.Millisecond)
	}
	return
}

func FetchList(spider *gorequest.SuperAgent, url string) (err error) {
	resp, _, errs := spider.Get(url).End()
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

	zeroUnix := MustDate(dateZero).Unix()
	linkSize, rows := len(IdLinkPrefix), []any{}
	sect.Find("div.card").Each(func(i int, card *goquery.Selection) {
		wp := &db.WallDaily{MaxDpi: "400x240"}
		imgSrc := card.Find("img").First().AttrOr("src", "")
		wp.OrigUrl = xq.NewNullString(imgSrc)
		dataDiv := card.Find("div.title").First()
		wp.Title = dataDiv.Find("div.name").First().Text()
		date := dataDiv.Find("div.date").First().Text()
		wp.BingDate = MustDate(date)
		wp.Id = int((wp.BingDate.Unix() - zeroUnix) / 86400)
		link := dataDiv.Find("a").First().AttrOr("href", "")
		if strings.HasPrefix(link, IdLinkPrefix) {
			wp.OrigId, _ = strconv.Atoi(link[linkSize:])
		}
		rows = append(rows, wp)
	})
	if len(rows) > 0 {
		table := (db.WallDaily{}).TableName()
		err = db.InsertBatch(table, rows)
	}
	return
}

func FetchImage(row *db.WallDaily) (err error) {
	return
}

func FetchNotes(row *db.WallDaily) (err error) {
	return
}

func GetMonthDirAndUrl(date string) (saveDir, listUrl string) {
	year, month := date[:4], date[5:7]
	saveDir = filepath.Join(CurrentPathDir, year+month)
	month = strings.TrimLeft(month, "0")
	listUrl = fmt.Sprintf("%s%s&y=%s&m=%s", SiteHomeUrl, SiteListPath, year, month)
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

func MustDate(date string) time.Time {
	obj, err := time.Parse("2006-01-02", date)
	if err == nil {
		panic(err)
	}
	return obj
}
