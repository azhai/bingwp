package handlers

import (
	"fmt"

	xutils "github.com/azhai/xgen/utils"
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
)

type ArchiveDict struct {
	Date     string `json:"enddate"`
	FilePath string `json:"urlbase"`
	Title    string `json:"copyright"`
	Headline string `json:"title"`
}

type ArchiveResult struct {
	Images []ArchiveDict `json:"images"`
}

func (r *ArchiveResult) ToDailyListData() (data []DailyDict) {
	for _, item := range r.Images {
		data = append(data, DailyDict{
			Date:     item.Date,
			FilePath: item.FilePath,
			Title:    item.Title,
			Headline: item.Headline,
		})
	}
	return
}

type DailyDict struct {
	Date     string `json:"date"`
	FilePath string `json:"filepath"`
	Title    string `json:"title"`
	Headline string `json:"headline"`
	OrigId   int    `json:"id,omitempty"`
}

type ListData struct {
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Data     []DailyDict `json:"data"`
}

type ListResult struct {
	Status   int      `json:"status"`
	Success  bool     `json:"success"`
	Response ListData `json:"response"`
}

type DetailDict struct {
	DailyDict     `json:",inline"`
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
	Status   int         `json:"status"`
	Success  bool        `json:"success"`
	Response *DetailDict `json:"response"`
}

// Crawler 网络爬虫
type Crawler struct {
	client *gorequest.SuperAgent
	err    error
}

// NewCrawler 创建cURL客户端
func NewCrawler() *Crawler {
	return &Crawler{
		client: gorequest.New().Set("User-Agent", UserAgent),
	}
}

// Error 最后一次的错误
func (c *Crawler) Error() error {
	return c.err
}

// Crawl 爬去页面内容
func (c *Crawler) Crawl(url string) (string, error) {
	_, body, errs := c.client.Get(url).End()
	if len(errs) > 0 {
		body = ""
		c.err = errs[0]
	} else {
		c.err = nil
	}
	return body, c.err
}

// CrawlArchive 爬取归档页面
func (c *Crawler) CrawlArchive() (int, error) {
	body, err := c.Crawl(ArchiveUrl)
	if err != nil {
		return 0, err
	}
	data := new(ArchiveResult)
	_, c.err = xutils.UnmarshalJSON([]byte(body), &data)
	return InsertDailyRows(data.ToDailyListData())
}

// CrawlList 爬取列表页面
func (c *Crawler) CrawlList(page int) (int, error) {
	url := fmt.Sprintf(ListUrl, page)
	body, err := c.Crawl(url)
	if err != nil {
		return 0, err
	}
	data := new(ListResult)
	_, c.err = xutils.UnmarshalJSON([]byte(body), &data)
	return InsertDailyRows(data.Response.Data)
}

// CrawlDetail 爬取详情页面
func (c *Crawler) CrawlDetail(origId int) *DetailDict {
	url := fmt.Sprintf(DetailUrl, origId)
	body, err := c.Crawl(url)
	if err != nil {
		return nil
	}
	data := new(DetailResult)
	_, c.err = xutils.UnmarshalJSON([]byte(body), &data)
	return data.Response
}
