package handlers

import (
	db "github.com/azhai/bingwp/models/default"
	xq "github.com/azhai/xgen/xquery"
)

func ReadDetail(wp *db.WallDaily) (err error) {
	if wp.OrigId <= 0 {
		return
	}
	crawler := NewCrawler()
	dict := crawler.CrawlDetail(wp.OrigId)
	var exData map[string]*db.WallNote
	if exData, err = db.LoadNoteByDailyId(wp.Id); err != nil {
		return
	}

	var notes []any
	if _, ok := exData["keyword"]; !ok && dict.Keyword != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "keyword"}
		note.NoteChinese = xq.NewNullString(dict.Keyword)
		notes = append(notes, note)
	}
	if _, ok := exData["headline"]; !ok && dict.Headline != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "headline"}
		note.NoteChinese = xq.NewNullString(dict.Headline)
		if dict.HeadlineEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.HeadlineEn)
		}
		notes = append(notes, note)
	}
	if _, ok := exData["quick_fact"]; !ok && dict.QuickFact != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "quick_fact"}
		note.NoteChinese = xq.NewNullString(dict.QuickFact)
		if dict.QuickFactEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.QuickFactEn)
		}
		notes = append(notes, note)
	}
	if _, ok := exData["title"]; !ok && dict.Title != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "title"}
		note.NoteChinese = xq.NewNullString(dict.Title)
		if dict.TitleEn != "" {
			note.NoteEnglish = xq.NewNullString(dict.TitleEn)
		}
		notes = append(notes, note)
	}
	if _, ok := exData["paragraph"]; !ok && dict.Description != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "paragraph"}
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
