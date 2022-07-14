package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/azhai/bingwp/cmd"
	db "github.com/azhai/bingwp/models/default"
	"github.com/azhai/bingwp/utils"

	xutils "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
	"github.com/k0kubun/pp"
	"github.com/parnurzeal/gorequest"
)

const (
	UserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edg/90.0.818.66"
	ListUrl      = "https://cn.bing.com/HPImageArchive.aspx?&format=js&idx=0&n=8&mkt=zh-CN"
	DetailUrl    = "https://api.wilii.cn/api/Bing/%d"
	SiteThumbUrl = "https://bing.wilii.cn/OneDrive/bingimages"
	BingThumbUrl = "https://s.cn.bing.net/th?id=OHR."
	ImageDataDir = "/data/bingwp/"
	dateFirst    = "20090713"
	dateZero     = "20081231"
)

type ItemDict struct {
	Date     string `json:"enddate"`
	Title    string `json:"copyright"`
	FilePath string `json:"urlbase"`
}

type ListResult struct {
	Images []ItemDict `json:"images"`
}

type DetailDict struct {
	Date          string           `json:"date"`
	FilePath      string           `json:"filepath"`
	Title         string           `json:"title"`
	Headline      string           `json:"headline"`
	TitleEn       string           `json:"titleEn"`
	HeadlineEn    string           `json:"headlineEn"`
	Description   string           `json:"description"`
	DescriptionEn string           `json:"descriptionEn"`
	QuickFact     string           `json:"quickFact"`
	QuickFactEn   string           `json:"quickFactEn"`
	Keyword       string           `json:"keyword"`
	Longitude     string           `json:"longitude"`
	Latitude      string           `json:"latitude"`
	Tags          []map[string]any `json:"tags"`
}

type DetailResult struct {
	Status   int        `json:"status"`
	Success  bool       `json:"success"`
	Response DetailDict `json:"response"`
}

func FetchRecent() (err error) {
	var stopDate time.Time
	row, ok := new(db.WallDaily), false
	ok, err = row.Load(xq.WithOrderBy("bing_date", true))
	if ok && err == nil {
		stopDate = row.BingDate
	}
	err = ReadList(stopDate.Format("20060102"))
	return
}

func ReadList(stopYmd string) (err error) {
	_, body, errs := CreateSpider().Get(ListUrl).End()
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
	items, rows := data.Images, make([]any, 0)
	pp.Println(items)
	var dims string
	for _, card := range items {
		wp := &db.WallDaily{MaxDpi: "400x240"}
		wp.BingDate = MustParseDate(card.Date)
		if wp.BingDate.Format("20060102") <= stopYmd {
			break
		}
		wp.BingSku = card.FilePath
		prelen := len("/th?id=OHR.")
		if strings.HasPrefix(wp.BingSku, "/th?id=OHR.") {
			wp.BingSku = strings.TrimSpace(wp.BingSku[prelen:])
		}
		wp.Brief = card.Title
		idx := strings.LastIndex(wp.Brief, "(")
		if idx > 0 && strings.HasSuffix(wp.Brief, ")") {
			wp.Brief = wp.Brief[:idx]
		}
		wp.Id = int((wp.BingDate.Unix() - zeroUnix) / 86400)
		// err = ReadDetail(wp)
		dims, err = FetchImage(wp)
		if dims != "" {
			wp.MaxDpi = dims
		}
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

func MustParseDate(date string) time.Time {
	obj, err := time.Parse("20060102", date)
	if err != nil {
		panic(err)
	}
	return obj
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
