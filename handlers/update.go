package handlers

import (
	"fmt"
	"time"

	db "github.com/azhai/bingwp/models/default"
	xutils "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
	"github.com/parnurzeal/gorequest"
)

func FetchRecent() (err error) {
	var stopDate time.Time
	wp, ok := new(db.WallDaily), false
	ok, err = wp.Load(xq.WithOrderBy("bing_date", true))
	if ok && err == nil && wp.Id > 0 {
		stopDate = wp.BingDate
		_, err = UpdateDailyImages(wp) // 看最后一天的图片是否需要重新下载
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
	items, rows := data.Images, make([]any, 0)
	// pp.Println(items)

	var dims string
	for _, card := range items {
		wp := &db.WallDaily{MaxDpi: "400x240"}
		wp.BingDate = MustParseDate(card.Date)
		if wp.BingDate.Format("20060102") <= stopYmd {
			break
		}
		wp.Id = GetOffsetDay(wp.BingDate)
		wp.BingSku = GetSkuFromBaseUrl(card.FilePath)
		wp.Title = ParseDailyTitle(card.Title)
		dims, err = UpdateDailyImages(wp)
		if dims != "" {
			wp.MaxDpi = dims
		}
		rows = append(rows, wp)
	}
	if len(rows) > 0 {
		table := new(db.WallDaily).TableName()
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
	items := data.Response.Data
	// pp.Println(items)

	var exData map[string]*db.WallDaily
	dates := make([]any, len(items))
	for i, card := range items {
		dates[i] = card.Date
	}
	if exData, err = db.LoadDailyByDates(dates...); err != nil {
		return
	}

	var wp *db.WallDaily
	for _, card := range items {
		ok, dims := false, ""
		changes := make(map[string]any)
		if wp, ok = exData[card.Date]; !ok {
			wp = &db.WallDaily{MaxDpi: "400x240"}
			wp.BingDate = MustParseDate(card.Date)
			wp.Id = GetOffsetDay(wp.BingDate)
			wp.BingSku = GetSkuFromFullUrl(card.FilePath)
			wp.Title = ParseDailyTitle(card.Title)
			wp.OrigId = card.OrigId
		} else {
			wp.OrigId = card.OrigId
			changes["orig_id"] = card.OrigId
		}
		if err = wp.Save(changes); err != nil {
			continue
		}

		err = ReadDetail(wp)
		dims, err = UpdateDailyImages(wp)
		if dims != "" && dims != wp.MaxDpi {
			changes["max_dpi"] = dims
			err = wp.Save(changes)
		}
	}
	return
}

func ReadDetail(row *db.WallDaily) (err error) {
	if row.OrigId <= 0 {
		return
	}
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
	var exData map[string]*db.WallNote
	if exData, err = db.LoadNoteByDailyId(row.Id); err != nil {
		return
	}

	dict, notes := data.Response, make([]any, 0)
	if _, ok := exData["keyword"]; !ok && dict.Keyword != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "keyword"}
		note.NoteChinese = xq.NewNullString(dict.Keyword)
		notes = append(notes, note)
	}
	if _, ok := exData["headline"]; !ok && dict.Headline != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "headline"}
		note.NoteChinese = xq.NewNullString(dict.Headline)
		if dict.HeadlineEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.HeadlineEn)
		}
		notes = append(notes, note)
	}
	if _, ok := exData["quick_fact"]; !ok && dict.QuickFact != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "quick_fact"}
		note.NoteChinese = xq.NewNullString(dict.QuickFact)
		if dict.QuickFactEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.QuickFactEn)
		}
		notes = append(notes, note)
	}
	if _, ok := exData["title"]; !ok && dict.Title != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "title"}
		note.NoteChinese = xq.NewNullString(dict.Title)
		if dict.TitleEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.TitleEn)
		}
		notes = append(notes, note)
	}
	if _, ok := exData["paragraph"]; !ok && dict.Description != "" {
		note := &db.WallNote{DailyId: row.Id, NoteType: "paragraph"}
		note.NoteChinese = xq.NewNullString(dict.Description)
		if dict.DescriptionEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.DescriptionEn)
		}
		notes = append(notes, note)
	}
	if len(notes) > 0 {
		table := new(db.WallNote).TableName()
		err = db.InsertBatch(table, notes...)
	}
	return
}

// CreateSpider 创建cURL客户端
func CreateSpider() *gorequest.SuperAgent {
	client := gorequest.New().Set("User-Agent", UserAgent)
	// if logger := config.GetDefaultLogger(); logger != nil {
	// 	curlLogger := &utils.CurlLogger{Logger: logger}
	// 	client = client.SetDebug(true).SetLogger(curlLogger)
	// }
	return client
}
