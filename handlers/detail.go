package handlers

import (
	"github.com/azhai/bingwp/services/database"
)

func WriteDetail(wp *database.WallDaily, data *DetailDict) (err error) {
	dailyRows := []*database.WallDaily{wp}
	dailyRows = database.GetDailyNotes(dailyRows)
	dict := dailyRows[0].Notes
	if dict == nil {
		dict = make(map[string]*database.WallNote)
	}
	var notes []*database.WallNote
	// if _, ok := dict["longitude"]; !ok && data.Longitude != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "longitude"}
	// 	note.NoteChinese = database.NewNullString(data.Longitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := dict["latitude"]; !ok && data.Latitude != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "latitude"}
	// 	note.NoteChinese = database.NewNullString(data.Latitude)
	// 	notes = append(notes, note)
	// }
	// if _, ok := dict["keyword"]; !ok && data.Keyword != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "keyword"}
	// 	note.NoteChinese = database.NewNullString(data.Keyword)
	// 	notes = append(notes, note)
	// }
	if _, ok := dict["title"]; !ok && data.Title != "" {
		note := &database.WallNote{DailyId: wp.Id, NoteType: "title"}
		note.NoteChinese = database.NewNullString(data.Title)
		if data.TitleEn != "" {
			note.NoteEnglish = database.NewNullString(data.TitleEn)
		}
		notes = append(notes, note)
	}
	if _, ok := dict["headline"]; !ok && data.Headline != "" {
		note := &database.WallNote{DailyId: wp.Id, NoteType: "headline"}
		note.NoteChinese = database.NewNullString(data.Headline)
		if data.HeadlineEn != "" {
			note.NoteEnglish = database.NewNullString(data.HeadlineEn)
		}
		notes = append(notes, note)
	}
	if _, ok := dict["description"]; !ok && data.Description != "" {
		note := &database.WallNote{DailyId: wp.Id, NoteType: "description"}
		note.NoteChinese = database.NewNullString(data.Description)
		if data.DescriptionEn != "" {
			note.NoteEnglish = database.NewNullString(data.DescriptionEn)
		}
		notes = append(notes, note)
	}
	// if _, ok := dict["quick_fact"]; !ok && data.QuickFact != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "quick_fact"}
	// 	note.NoteChinese = database.NewNullString(data.QuickFact)
	// 	if data.QuickFactEn != "" {
	// 		note.NoteEnglish = database.NewNullString(data.QuickFactEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	// if _, ok := dict["caption"]; !ok && data.Caption != "" {
	// 	note := &database.WallNote{DailyId: wp.Id, NoteType: "caption"}
	// 	note.NoteChinese = database.NewNullString(data.Caption)
	// 	if data.CaptionEn != "" {
	// 		note.NoteEnglish = database.NewNullString(data.CaptionEn)
	// 	}
	// 	notes = append(notes, note)
	// }
	if len(notes) > 0 {
		_, err = database.InsertBatch(notes)
	}
	return
}
