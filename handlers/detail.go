package handlers

import (
	"fmt"

	db "github.com/azhai/bingwp/models/default"
	xq "github.com/azhai/xgen/xquery"
)

func WriteDetail(wp *db.WallDaily, data *DetailDict) (err error) {
	var dict map[string]*db.WallNote
	if dict, err = db.LoadNoteByDailyId(wp.Id); err != nil {
		fmt.Println(err)
		return
	}
	var notes []any
	// if _, ok := dict["longitude"]; !ok && data.Longitude != "" {
	// 	note := &db.WallNote{DailyId: wp.Id, NoteType: "longitude"}
	// 	note.NoteChinese = xq.NewNullString(data.Longitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := dict["latitude"]; !ok && data.Latitude != "" {
	// 	note := &db.WallNote{DailyId: wp.Id, NoteType: "latitude"}
	// 	note.NoteChinese = xq.NewNullString(data.Latitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := dict["keyword"]; !ok && data.Keyword != "" {
	// 	note := &db.WallNote{DailyId: wp.Id, NoteType: "keyword"}
	// 	note.NoteChinese = xq.NewNullString(data.Keyword)
	// 	notes = append(notes, note)
	// }
	if _, ok := dict["title"]; !ok && data.Title != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "title"}
		note.NoteChinese = xq.NewNullString(data.Title)
		if data.TitleEn != "" {
			note.NoteEnglish = xq.NewNullString(data.TitleEn)
		}
		notes = append(notes, note)
	}
	if _, ok := dict["headline"]; !ok && data.Headline != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "headline"}
		note.NoteChinese = xq.NewNullString(data.Headline)
		if data.HeadlineEn != "" {
			note.NoteEnglish = xq.NewNullString(data.HeadlineEn)
		}
		notes = append(notes, note)
	}
	if _, ok := dict["description"]; !ok && data.Description != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "description"}
		note.NoteChinese = xq.NewNullString(data.Description)
		if data.DescriptionEn != "" {
			note.NoteEnglish = xq.NewNullString(data.DescriptionEn)
		}
		notes = append(notes, note)
	}
	// if _, ok := dict["quick_fact"]; !ok && data.QuickFact != "" {
	// 	note := &db.WallNote{DailyId: wp.Id, NoteType: "quick_fact"}
	// 	note.NoteChinese = xq.NewNullString(data.QuickFact)
	// 	if data.QuickFactEn != "" {
	// 		note.NoteEnglish = xq.NewNullString(data.QuickFactEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	// if _, ok := dict["caption"]; !ok && data.Caption != "" {
	// 	note := &db.WallNote{DailyId: wp.Id, NoteType: "caption"}
	// 	note.NoteChinese = xq.NewNullString(data.Caption)
	// 	if data.CaptionEn != "" {
	// 		note.NoteEnglish = xq.NewNullString(data.CaptionEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	if len(notes) > 0 {
		table := new(db.WallNote).TableName()
		err = db.InsertBatch(table, notes...)
	}
	return
}
