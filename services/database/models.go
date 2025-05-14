package database

import (
	"database/sql"
	"time"
)

// WallDaily 每日壁纸
type WallDaily struct {
	Id               int64     `json:"id" form:"id" db:"pk;type:int"`
	Guid             string    `json:"guid" form:"guid" db:"type:char(10)"`
	BingDate         time.Time `json:"bing_date" form:"bing_date" db:"unique;type:date"`
	BingSku          string    `json:"bing_sku" form:"bing_sku" db:"index;type:varchar(100)"`
	Title            string    `json:"title" form:"title" db:"index;type:varchar(255)"`
	Headline         string    `json:"headline" form:"headline" db:"type:varchar(255)"`
	Color            string    `json:"color" form:"color" db:"type:varchar(15)"`
	MaxDpi           string    `json:"max_dpi" form:"max_dpi" db:"type:varchar(15)"`
	WallDailyForeign `json:",inline" db:"-"`
}

// TableName WallDaily的表名
func (*WallDaily) TableName() string {
	return "t_wall_daily"
}

// TableComment WallDaily的备注
func (*WallDaily) TableComment() string {
	return "每日壁纸"
}

// WallImage 壁纸图片
type WallImage struct {
	Id            int64 `json:"id" form:"id" db:"pk;type:int"`
	DailyId       int64 `json:"daily_id" form:"daily_id" db:"index;type:int"`
	ImageUrlMixin `json:",inline" db:"inline"`
	ImgSize       int64 `json:"img_size" form:"img_size" db:"index;type:int"`
	ImgOffset     int64 `json:"img_offset" form:"img_offset" db:"type:int"`
	ImgWidth      int   `json:"img_width" form:"img_width" db:"type:int"`
	ImgHeight     int   `json:"img_height" form:"img_height" db:"type:int"`
}

// TableName WallImage的表名
func (*WallImage) TableName() string {
	return "t_wall_image"
}

// TableComment WallImage的备注
func (*WallImage) TableComment() string {
	return "壁纸图片"
}

// WallNote 壁纸小知识
type WallNote struct {
	Id          int64          `json:"id" form:"id" db:"pk;serial;type:int"`
	DailyId     int64          `json:"daily_id" form:"daily_id" db:"index;type:int"`
	NoteType    string         `json:"note_type" form:"note_type" db:"type:varchar(50)"`
	NoteChinese sql.NullString `json:"note_chinese" form:"note_chinese" db:"type:text"`
	NoteEnglish sql.NullString `json:"note_english" form:"note_english" db:"type:text"`
}

// TableName WallNote的表名
func (*WallNote) TableName() string {
	return "t_wall_note"
}

// TableComment WallNote的备注
func (*WallNote) TableComment() string {
	return "壁纸小知识"
}
