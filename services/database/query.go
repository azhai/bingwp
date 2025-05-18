package database

import (
	"fmt"
	"time"
)

// WallDailyForeign 关联查询字段
type WallDailyForeign struct {
	ImageUrl string               `json:"image_url" form:"image_url" db:"-"`
	ThumbUrl string               `json:"thumb_url" form:"thumb_url" db:"-"`
	Thumb    *WallImage           `json:"thumb" form:"thumb" db:"-"`
	Image    *WallImage           `json:"image" form:"image" db:"-"`
	Notes    map[string]*WallNote `json:"notes" form:"notes" db:"-"`
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

// ImageDimMixin 图片宽高
type ImageDimMixin struct {
	ImgWidth  int `json:"img_width" form:"img_width" db:"type:int"`
	ImgHeight int `json:"img_height" form:"img_height" db:"type:int"`
}

// GeDims 获取图片的尺寸
func (m *ImageDimMixin) GeDims() string {
	return fmt.Sprintf("%dx%d", m.ImgWidth, m.ImgHeight)
}

// GetLatestDailyRows 读取最后的一些每日壁纸
func GetLatestDailyRows(limit, start int) []*WallDaily {
	table := new(WallDaily).TableName()
	tpl := "SELECT * FROM %s ORDER BY id DESC LIMIT %d OFFSET %d"
	rows, err := New().Query(fmt.Sprintf(tpl, table, limit, start))
	var dailyRows []*WallDaily
	if err == nil {
		err = ScanToList(&dailyRows, rows)
	}
	CheckErr(err)
	return dailyRows
}

// GetMonthDailyRows 读取指定月份的每日壁纸
func GetMonthDailyRows(monthBegin, nextBegin time.Time) []*WallDaily {
	table := new(WallDaily).TableName()
	date1, date2 := monthBegin.Format("2006-01-02"), nextBegin.Format("2006-01-02")
	query := "SELECT * FROM " + table + " WHERE bing_date >= $1 AND bing_date < $2 ORDER BY id"
	rows, err := New().Query(query, date1, date2)
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
	tpl := "SELECT * FROM %s WHERE daily_id IN (%s) ORDER BY daily_id"
	rows, err := New().Query(fmt.Sprintf(tpl, table, ids))
	var imageRows = make(map[int64]map[string]*WallImage)
	if err == nil {
		err = ScanToSecondary(imageRows, rows)
	}
	CheckErr(err)
	var img *WallImage
	for i, row := range dailyRows {
		if dict, ok := imageRows[row.Id]; ok {
			if img, ok = dict["thumb"]; ok {
				row.Thumb = img
				row.ThumbUrl = img.GetUrl()
			}
			if img, ok = dict["image"]; ok {
				row.Image = img
				row.ImageUrl = img.GetUrl()
			}
		}
		dailyRows[i] = row
	}
	return dailyRows
}

// GetDailyNotes 从数据库中加载每日图片的小知识
func GetDailyNotes(dailyRows []*WallDaily) []*WallDaily {
	table := new(WallNote).TableName()
	ids := WallDailyList(dailyRows).GetIds()
	tpl := "SELECT * FROM %s WHERE daily_id IN (%s) ORDER BY daily_id"
	rows, err := New().Query(fmt.Sprintf(tpl, table, ids))
	var noteRows = make(map[int64]map[string]*WallNote)
	if err == nil {
		err = ScanToSecondary(noteRows, rows)
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
	if dates == "" && len(dailyRows) > 0 {
		dates = WallDailyList(dailyRows).GetDates()
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE bing_date IN (%s)", table, dates)
	rows, err := New().Query(query)
	var existRows = make(map[string]*WallDaily)
	if err == nil {
		err = ScanToUnique(existRows, rows)
	}
	CheckErr(err)
	var newbieRows []*WallDaily
	for _, row := range dailyRows {
		bingDate := row.BingDate.Format("2006-01-02")
		if _, ok := existRows[bingDate]; !ok {
			newbieRows = append(newbieRows, row)
		} else if row.Guid != "" || row.Color != "" {
			changes := map[string]any{"guid": row.Guid, "color": row.Color}
			_, err = UpdateRow(row, changes)
		}
	}
	return InsertBatch(newbieRows)
}
