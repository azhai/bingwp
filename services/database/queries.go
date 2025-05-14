package database

import (
	"strconv"
	"strings"
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

// GetMonthDailyRows 读取指定月份的每日壁纸
func GetMonthDailyRows(monthBegin, nextBegin time.Time) []WallDaily {
	sql := "SELECT * FROM t_wall_daily WHERE bing_date >= $1 AND bing_date < $2 ORDER BY id"
	rows, err := GetDB().Query(sql, monthBegin.Format("2006-01-02"), nextBegin.Format("2006-01-02"))
	CheckErr(err)
	defer rows.Close()
	var dailyRows []WallDaily
	for rows.Next() {
		var row WallDaily
		err = rows.Scan(&row.Id, &row.Guid, &row.BingDate, &row.BingSku, &row.Title, &row.Headline, &row.Color, &row.MaxDpi)
		CheckErr(err)
		dailyRows = append(dailyRows, row)
	}
	return dailyRows
}

// GetDailyImages 从数据库中加载每日图片的URL地址
func GetDailyImages(dailyRows []WallDaily) []WallDaily {
	var ids []string
	for _, row := range dailyRows {
		ids = append(ids, strconv.FormatInt(row.Id, 10))
	}
	sql := "SELECT * FROM t_wall_image WHERE daily_id IN (" + strings.Join(ids, ",") + ") ORDER BY daily_id"
	rows, err := GetDB().Query(sql)
	CheckErr(err)
	defer rows.Close()
	thumbs, images := make(map[int64]*WallImage), make(map[int64]*WallImage)
	for rows.Next() {
		var img = new(WallImage)
		err = rows.Scan(&img.Id, &img.DailyId, &img.FileName, &img.ImgMd5, &img.ImgSize, &img.ImgOffset, &img.ImgWidth, &img.ImgHeight)
		CheckErr(err)
		if strings.HasPrefix(img.FileName, "thumb") {
			thumbs[img.DailyId] = img
		} else {
			images[img.DailyId] = img
		}
	}
	for i, row := range dailyRows {
		if img, ok := thumbs[row.Id]; ok {
			row.Thumb = img
			row.ThumbUrl = img.GetUrl()
		}
		if img, ok := images[row.Id]; ok {
			row.Image = img
			row.ImageUrl = img.GetUrl()
		}
		dailyRows[i] = row
	}
	return dailyRows
}

// GetDailyNotes 从数据库中加载每日图片的小知识
func GetDailyNotes(dailyRows []WallDaily) []WallDaily {
	var ids []string
	for _, row := range dailyRows {
		ids = append(ids, strconv.FormatInt(row.Id, 10))
	}
	sql := "SELECT * FROM t_wall_note WHERE daily_id IN (" + strings.Join(ids, ",") + ") ORDER BY daily_id"
	rows, err := GetDB().Query(sql)
	CheckErr(err)
	defer rows.Close()
	notes := make(map[int64]map[string]*WallNote)
	for rows.Next() {
		var note = new(WallNote)
		err = rows.Scan(&note.Id, &note.DailyId, &note.NoteType, &note.NoteChinese, &note.NoteEnglish)
		CheckErr(err)
		if _, ok := notes[note.DailyId]; !ok {
			notes[note.DailyId] = make(map[string]*WallNote)
		}
		notes[note.DailyId][note.NoteType] = note
	}
	for i, row := range dailyRows {
		if dict, ok := notes[row.Id]; ok {
			row.Notes = dict
		}
		dailyRows[i] = row
	}
	return dailyRows
}
