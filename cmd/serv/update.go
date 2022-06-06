package main

// https://bing.wilii.cn/ymd.asp?ismobile=0&y=2022&m=5
// https://bing.wilii.cn/OneDrive/bingimages/2022/05/01/VanBlooms_ZH-CN6370306779_400x240.jpg
// https://bing.wilii.cn/photo.html?id=5674
// https://bing.wilii.cn/download.asp?filename=/OneDrive/bingimages/2022/05/01/VanBlooms_ZH-CN6370306779_UHD.jpg&name=%E7%9B%9B%E5%BC%80%E7%9A%84%E9%87%91%E9%93%BE%E8%8A%B1%E6%A0%91%E5%92%8C%E7%B4%AB%E8%89%B2%E8%91%B1%E5%B1%9E%E6%A4%8D%E7%89%A9%EF%BC%8C%E5%8A%A0%E6%8B%BF%E5%A4%A7%E6%B8%A9%E5%93%A5%E5%8D%8E%E8%8C%83%E5%BA%A6%E6%A3%AE%E6%A4%8D%E7%89%A9%E5%9B%AD_UHD.jpg&id=5674

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/azhai/bingwp/cmd"
	"github.com/azhai/bingwp/utils"
	"golang.org/x/net/html"

	hq "github.com/antchfx/htmlquery"
	"github.com/parnurzeal/gorequest"
)

const (
	UserAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edg/90.0.818.66"
	WallPaperURL       = "https://bing.wilii.cn/ymd.asp?ismobile=0"
	CurrentPathDir     = "images/"
	dateFirst, dateUHD = "2009-07-13", "2020-09-01"
)

type WallPaper struct {
	Image, Title, Date string
}

// FetchWallPapers 获取Bing背景图列表
func FetchWallPapers() (papers []WallPaper, err error) {
	date := time.Now().Format("2006-01-02")
	_, listUrl := GetMonthDirAndUrl(date)
	resp, body, errs := CreateSpider().Get(listUrl).End()
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	fmt.Println(body)
	var htmlDoc *html.Node
	if htmlDoc, err = hq.Parse(resp.Body); err != nil {
		return
	}

	divCards, err := hq.QueryAll(htmlDoc, "//div[@class=\"card progressive\"]")
	for _, card := range divCards {
		imgSrc := hq.SelectAttr(hq.FindOne(card, "//img/@src"), "src")
		wp := WallPaper{Image: imgSrc}
		desc := hq.FindOne(card, "//div[@class=\"description\"]")
		wp.Title = hq.InnerText(hq.FindOne(desc, "//h3"))
		cal := hq.FindOne(desc, "//p[@class=\"calendar\"]")
		wp.Date = hq.InnerText(hq.FindOne(cal, "//em"))
		papers = append(papers, wp)
	}

	return
}

func GetMonthDirAndUrl(date string) (saveDir, listUrl string) {
	year, month := date[:4], date[5:7]
	saveDir = filepath.Join(CurrentPathDir, year+month)
	month = strings.TrimLeft(month, "0")
	listUrl = WallPaperURL + fmt.Sprintf("&y=%s&m=%s", year, month)
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
