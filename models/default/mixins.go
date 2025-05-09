package db

import (
	xq "github.com/azhai/xgen/xquery"
)

// LoadDailyByDates 加载指定日期的行
func LoadDailyByDates(dates ...any) (map[string]*WallDaily, error) {
	var rows []*WallDaily
	data := make(map[string]*WallDaily)
	where := xq.WithRange("bing_date", dates...)
	err := Query(where).Asc("bing_date").Find(&rows)
	if err == nil {
		for _, row := range rows {
			dt := row.BingDate.Format("2006-01-02")
			data[dt] = row
		}
	}
	return data, err
}

// LoadNoteByDailyId 加载对应行的描述
func LoadNoteByDailyId(dailyId int64) (map[string]*WallNote, error) {
	var rows []*WallNote
	data := make(map[string]*WallNote)
	where := xq.WithWhere("daily_id = ?", dailyId)
	err := Query(where).Find(&rows)
	if err == nil {
		for _, row := range rows {
			data[row.NoteType] = row
		}
	}
	return data, err
}
