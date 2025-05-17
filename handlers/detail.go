package handlers

import (
	"github.com/azhai/bingwp/services/database"
)

func GetNotExistNotes(wp *database.WallDaily, data *DetailDict) (
	notes []*database.WallNote, err error) {
	if wp.Notes == nil {
		wp.Notes = make(map[string]*database.WallNote)
	}
	// if _, ok := wp.Notes["longitude"]; !ok && data.Longitude != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "longitude"}
	// 	note.NoteChinese = database.NewNullString(data.Longitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := wp.Notes["latitude"]; !ok && data.Latitude != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "latitude"}
	// 	note.NoteChinese = database.NewNullString(data.Latitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := wp.Notes["keyword"]; !ok && data.Keyword != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "keyword"}
	// 	note.NoteChinese = database.NewNullString(data.Keyword)
	// 	notes = append(notes, note)
	// }
	if _, ok := wp.Notes["title"]; !ok && data.Title != "" {
		note := &database.WallNote{DailyId: wp.Id, NoteType: "title"}
		note.NoteChinese = database.NewNullString(data.Title)
		if data.TitleEn != "" {
			note.NoteEnglish = database.NewNullString(data.TitleEn)
		}
		notes = append(notes, note)
	}
	if _, ok := wp.Notes["headline"]; !ok && data.Headline != "" {
		note := &database.WallNote{DailyId: wp.Id, NoteType: "headline"}
		note.NoteChinese = database.NewNullString(data.Headline)
		if data.HeadlineEn != "" {
			note.NoteEnglish = database.NewNullString(data.HeadlineEn)
		}
		notes = append(notes, note)
	}
	if _, ok := wp.Notes["description"]; !ok && data.Description != "" {
		note := &database.WallNote{DailyId: wp.Id, NoteType: "description"}
		note.NoteChinese = database.NewNullString(data.Description)
		if data.DescriptionEn != "" {
			note.NoteEnglish = database.NewNullString(data.DescriptionEn)
		}
		notes = append(notes, note)
	}
	// if _, ok := wp.Notes["quick_fact"]; !ok && data.QuickFact != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "quick_fact"}
	// 	note.NoteChinese = database.NewNullString(data.QuickFact)
	// 	if data.QuickFactEn != "" {
	// 		note.NoteEnglish = database.NewNullString(data.QuickFactEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	// if _, ok := wp.Notes["caption"]; !ok && data.Caption != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "caption"}
	// 	note.NoteChinese = database.NewNullString(data.Caption)
	// 	if data.CaptionEn != "" {
	// 		note.NoteEnglish = database.NewNullString(data.CaptionEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	return
}
