package database

import (
	"time"
)

// WallDailyForeign 关联查询字段
type WallDailyForeign struct {
	ImageUrl string      `json:"image_url" form:"image_url" db:"-"`
	ThumbUrl string      `json:"thumb_url" form:"thumb_url" db:"-"`
	Thumb    *WallImage  `json:"thumb" form:"thumb" db:"-"`
	Image    *WallImage  `json:"image" form:"image" db:"-"`
	Notes    []*WallNote `json:"notes" form:"notes" db:"-"`
}

// ImageUrlMixin 图片URL
type ImageUrlMixin struct {
	FileName string `json:"file_name" form:"file_name" db:"type:varchar(100)"`
	ImgMd5   string `json:"img_md5" form:"img_md5" db:"index;type:char(32)"`
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

// GetMonthDailyRows 读取指定月份的每日壁纸
func GetMonthDailyRows(monthBegin, nextBegin time.Time) []*WallDaily {
	table := new(WallDaily).TableName()
	date1, date2 := monthBegin.Format("2006-01-02"), nextBegin.Format("2006-01-02")
	sql := "SELECT * FROM " + table + " WHERE bing_date >= $1 AND bing_date < $2 ORDER BY id"
	rows, err := New().Query(sql, date1, date2)
	var dailyRows []*WallDaily
	if err == nil {
		err = ScanToList(&dailyRows, rows)
	}
	CheckErr(err)
	return dailyRows
}

// GetDailyImages 从数据库中加载每日图片的URL地址
func GetDailyImages(dailyRows []*WallDaily) []*WallDaily {
	table := new(WallImage).TableName()
	ids := WallDailyList(dailyRows).GetIds()
	sql := "SELECT * FROM " + table + " WHERE daily_id IN (" + ids + ")"
	rows, err := New().Query(sql + " AND file_name LIKE 'thumb/%' ORDER BY daily_id")
	var thumbRows = make(map[int64]*WallImage)
	if err == nil {
		err = ScanToUnique(thumbRows, rows)
	}
	CheckErr(err)
	rows, err = New().Query(sql + " AND file_name LIKE 'image/%' ORDER BY daily_id")
	var imageRows = make(map[int64]*WallImage)
	if err == nil {
		err = ScanToUnique(imageRows, rows)
	}
	CheckErr(err)
	for i, row := range dailyRows {
		if img, ok := thumbRows[row.Id]; ok {
			row.Thumb = img
			row.ThumbUrl = img.GetUrl()
		}
		if img, ok := imageRows[row.Id]; ok {
			row.Image = img
			row.ImageUrl = img.GetUrl()
		}
		dailyRows[i] = row
	}
	return dailyRows
}

// GetDailyNotes 从数据库中加载每日图片的小知识
func GetDailyNotes(dailyRows []*WallDaily) []*WallDaily {
	table := new(WallNote).TableName()
	ids := WallDailyList(dailyRows).GetIds()
	sql := "SELECT * FROM " + table + " WHERE daily_id IN (" + ids + ") ORDER BY daily_id"
	rows, err := New().Query(sql)
	var noteRows = make(map[int64][]*WallNote)
	if err == nil {
		err = ScanToIndex(noteRows, rows)
	}
	CheckErr(err)
	for i, row := range dailyRows {
		if notes, ok := noteRows[row.Id]; ok {
			row.Notes = notes
		}
		dailyRows[i] = row
	}
	return dailyRows
}

// InsertDailyRows 保存每日壁纸，排除已有日期的行
func InsertDailyRows(dailyRows []*WallDaily, dates string) (int, error) {
	model := new(WallDaily)
	table := model.TableName()
	sql := "SELECT * FROM " + table + " WHERE bing_date IN (" + dates + ")"
	rows, err := New().Query(sql)
	var existRows = make(map[string]*WallDaily)
	if err == nil {
		err = ScanToUnique(existRows, rows)
	}
	CheckErr(err)
	for i, row := range dailyRows {
		bingDate := row.BingDate.Format("2006-01-02")
		if _, ok := existRows[bingDate]; ok {
			dailyRows[i] = nil
		}
	}
	return InsertBatch(dailyRows)
}
