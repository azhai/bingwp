package db

import (
	"strings"

	xq "github.com/azhai/xgen/xquery"
)

// WallDailyForeign 关联查询字段
type WallDailyForeign struct {
	Notes    map[string]*WallNote  `json:"notes" form:"notes" xorm:"-"`
	Images   map[string]*WallImage `json:"images" form:"images" xorm:"-"`
	ImageUrl string                `json:"image_url" form:"image_url" xorm:"-"`
	ThumbUrl string                `json:"thumb_url" form:"thumb_url" xorm:"-"`
}

// ImageUrlMixin 图片URL
type ImageUrlMixin struct {
	FileName string `json:"file_name" form:"file_name" xorm:"notnull VARCHAR(100) comment('文件路径')"`
	ImgMd5   string `json:"img_md5" form:"img_md5" xorm:"notnull index CHAR(32) comment('图片MD5哈希')"`
}

// GetUrl 获取图片的URL地址
// 若图片的MD5码不为空，则取后8位作为版本号
func (m *ImageUrlMixin) GetUrl() string {
	url := m.FileName
	if len(m.ImgMd5) > 24 {
		url += "?v=" + m.ImgMd5[24:]
	}
	return url
}

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

// GetDailyImages 从数据库中加载每日图片的URL地址
func GetDailyImages(rows []*WallDaily) []*WallDaily {
	var ids []any
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	where := xq.WithRange("daily_id", ids...)
	var imgRows []*WallImage
	if err := Query(where).Find(&imgRows); err != nil {
		return rows
	}
	thumbs, images := make(map[int64]string), make(map[int64]string)
	for _, row := range imgRows {
		url := row.GetUrl()
		if strings.HasPrefix(row.FileName, "thumb") {
			thumbs[row.DailyId] = url
		} else {
			images[row.DailyId] = url
		}
	}
	for i, row := range rows {
		if url, ok := thumbs[row.Id]; ok {
			row.ThumbUrl = url
		}
		if url, ok := images[row.Id]; ok {
			row.ImageUrl = url
		}
		rows[i] = row
	}
	return rows
}

// GetDailyNotes 从数据库中加载每日图片的小知识
func GetDailyNotes(rows []*WallDaily) []*WallDaily {
	var ids []any
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	where := xq.WithRange("daily_id", ids...)
	var noteRows []*WallNote
	if err := Query(where).Find(&noteRows); err != nil {
		return rows
	}
	notes := make(map[int64]map[string]*WallNote)
	for _, row := range noteRows {
		if _, ok := notes[row.DailyId]; !ok {
			notes[row.DailyId] = map[string]*WallNote{
				"title": nil, "headline": nil, "description": nil,
			}
		}
		notes[row.DailyId][row.NoteType] = row
	}
	for i, row := range rows {
		if dict, ok := notes[row.Id]; ok {
			row.Notes = dict
		}
		rows[i] = row
	}
	return rows
}
