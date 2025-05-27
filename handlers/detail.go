package handlers

import (
	"github.com/azhai/bingwp/services/db"
)

func GetNotExistNotes(wp *db.WallDaily, data *DetailDict) (
	notes []*db.WallNote, err error) {
	if wp.Notes == nil {
		wp.Notes = make(map[string]*db.WallNote)
	}
	// if _, ok := wp.Notes["longitude"]; !ok && data.Longitude != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "longitude"}
	// 	note.NoteChinese = db.NewNullString(data.Longitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := wp.Notes["latitude"]; !ok && data.Latitude != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "latitude"}
	// 	note.NoteChinese = db.NewNullString(data.Latitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := wp.Notes["keyword"]; !ok && data.Keyword != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "keyword"}
	// 	note.NoteChinese = db.NewNullString(data.Keyword)
	// 	notes = append(notes, note)
	// }
	if _, ok := wp.Notes["title"]; !ok && data.Title != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "title"}
		note.NoteChinese = db.NewNullString(data.Title)
		if data.TitleEn != "" {
			note.NoteEnglish = db.NewNullString(data.TitleEn)
		}
		notes = append(notes, note)
	}
	if _, ok := wp.Notes["headline"]; !ok && data.Headline != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "headline"}
		note.NoteChinese = db.NewNullString(data.Headline)
		if data.HeadlineEn != "" {
			note.NoteEnglish = db.NewNullString(data.HeadlineEn)
		}
		notes = append(notes, note)
	}
	if _, ok := wp.Notes["description"]; !ok && data.Description != "" {
		note := &db.WallNote{DailyId: wp.Id, NoteType: "description"}
		note.NoteChinese = db.NewNullString(data.Description)
		if data.DescriptionEn != "" {
			note.NoteEnglish = db.NewNullString(data.DescriptionEn)
		}
		notes = append(notes, note)
	}
	// if _, ok := wp.Notes["quick_fact"]; !ok && data.QuickFact != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "quick_fact"}
	// 	note.NoteChinese = db.NewNullString(data.QuickFact)
	// 	if data.QuickFactEn != "" {
	// 		note.NoteEnglish = db.NewNullString(data.QuickFactEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	// if _, ok := wp.Notes["caption"]; !ok && data.Caption != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "caption"}
	// 	note.NoteChinese = db.NewNullString(data.Caption)
	// 	if data.CaptionEn != "" {
	// 		note.NoteEnglish = db.NewNullString(data.CaptionEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	return
}
