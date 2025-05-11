package db

import (
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
