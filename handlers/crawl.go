package handlers

import (
	"fmt"
	"os"

	xutils "github.com/azhai/xgen/utils"
	"github.com/parnurzeal/gorequest"
)

const (
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edg/90.0.818.66"
	ArchiveUrl    = "https://cn.bing.com/HPImageArchive.aspx?&format=js&mkt=zh-CN&idx=%d&n=8&uhd=1&uhdwidth=3840&uhdheight=2160"
	ListUrl       = "https://api.wilii.cn/api/bing?page=%d&pageSize=%d"
	DetailUrl     = "https://api.wilii.cn/api/Bing/%s"
	FullUrlPrefix = "https://bing.wilii.cn/OneDrive/bingimages/"
	BaseUrlPrefix = "/th?id=OHR."
	BingThumbUrl  = "https://s.cn.bing.net/th?id=OHR."
)

type ArchiveDict struct {
	Date     string `json:"enddate"` // 格式20060102
	FilePath string `json:"urlbase"`
	Title    string `json:"copyright"`
	Headline string `json:"title"`
}

type ArchiveResult struct {
	Images []ArchiveDict `json:"images"`
}

func (r *ArchiveResult) ToDailyListData(stopYmd string) (data []DailyDict) {
	for _, item := range r.Images {
		if item.Date <= stopYmd {
			break
		}
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
	Guid     string `json:"guid,omitempty"`
	Date     string `json:"date"` // 格式2006-01-02
	FilePath string `json:"filepath"`
	Title    string `json:"title"`
	Headline string `json:"headline"`
	Color    string `json:"color,omitempty"`
}

type ListData struct {
	Page      int         `json:"page"`
	PageSize  int         `json:"pageSize"`
	PageCount int         `json:"pageCount"`
	DataCount int         `json:"dataCount"`
	Data      []DailyDict `json:"data"`
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
	Caption       string `json:"caption"`
	CaptionEn     string `json:"captionEn"`
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
func (c *Crawler) Crawl(url string) ([]byte, error) {
	_, body, errs := c.client.Get(url).EndBytes()
	if len(errs) > 0 {
		c.err = errs[0]
	} else {
		c.err = nil
	}
	return body, c.err
}

// CrawlArchive 爬取归档页面
func (c *Crawler) CrawlArchive(offset int, stopYmd string) (*ArchiveResult, error) {
	url := fmt.Sprintf(ArchiveUrl, offset)
	body, err := c.Crawl(url)
	if err != nil {
		return nil, err
	}
	data := new(ArchiveResult)
	_, c.err = xutils.UnmarshalJSON(body, &data)
	return data, c.err
}

// SavelArchive 保存归档到数据库
func (c *Crawler) SavelArchive(offset int, stopYmd string) (int, error) {
	data, err := c.CrawlArchive(offset, stopYmd)
	rows := data.ToDailyListData(stopYmd)
	if err != nil {
		return 0, err
	}
	return InsertNotExistDailyRows(rows, true)
}

// CrawlList 爬取列表页面
func (c *Crawler) CrawlList(page, size int) (*ListResult, error) {
	url := fmt.Sprintf(ListUrl, page, size)
	body, err := c.Crawl(url)
	if err != nil {
		return nil, err
	}
	data := new(ListResult)
	_, c.err = xutils.UnmarshalJSON(body, &data)
	return data, c.err
}

// SaveList 保存列表到数据库
func (c *Crawler) SaveList(page, size int) (int, error) {
	data, err := c.CrawlList(page, size)
	if err != nil || !data.Success {
		return 0, err
	}
	rows := data.Response.Data
	return InsertNotExistDailyRows(rows, false)
}

// CrawlDetail 爬取详情页面
func (c *Crawler) CrawlDetail(guid string) *DetailDict {
	url := fmt.Sprintf(DetailUrl, guid)
	body, err := c.Crawl(url)
	if err != nil {
		return nil
	}
	path := fmt.Sprintf(SaveDetailFileName, guid)
	_ = os.WriteFile(path, body, 0644)
	data := new(DetailResult)
	_, c.err = xutils.UnmarshalJSON(body, &data)
	return data.Response
}
