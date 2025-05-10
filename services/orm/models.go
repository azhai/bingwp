package orm

import (
	"time"
)

// WallDaily 每日壁纸
type WallDaily struct {
	Id       int64       `json:"id" form:"id" goe:"pk;type:int"`
	Guid     string      `json:"guid" form:"guid" goe:"type:char(10)"`
	BingDate time.Time   `json:"bing_date" form:"bing_date" goe:"unique;type:date"`
	BingSku  string      `json:"bing_sku" form:"bing_sku" goe:"index;type:varchar(100)"`
	Title    string      `json:"title" form:"title" goe:"index;type:varchar(255)"`
	Headline string      `json:"headline" form:"headline" goe:"type:varchar(255)"`
	Color    string      `json:"color" form:"color" goe:"type:varchar(15)"`
	MaxDpi   string      `json:"max_dpi" form:"max_dpi" goe:"type:varchar(15)"`
	Images   []WallImage `json:"images" form:"images" goe:"-"`
	Notes    []WallNote  `json:"notes" form:"notes" goe:"-"`
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
	Id        int64  `json:"id" form:"id" goe:"pk;type:int"`
	DailyId   int64  `json:"daily_id" form:"daily_id" goe:"index;type:int"`
	FileName  string `json:"file_name" form:"file_name" goe:"type:varchar(100)"`
	ImgMd5    string `json:"img_md5" form:"img_md5" goe:"index;type:char(32)"`
	ImgSize   int64  `json:"img_size" form:"img_size" goe:"index;type:int"`
	ImgOffset int64  `json:"img_offset" form:"img_offset" goe:"type:int"`
	ImgWidth  int    `json:"img_width" form:"img_width" goe:"type:int"`
	ImgHeight int    `json:"img_height" form:"img_height" goe:"type:int"`
}

// TableName WallImage的表名
func (*WallImage) TableName() string {
	return "t_wall_image"
}

// TableComment WallImage的备注
func (*WallImage) TableComment() string {
	return "壁纸图片"
}

// WallNote 墙纸小知识
type WallNote struct {
	Id          int64   `json:"id" form:"id" goe:"pk;serial;type:int"`
	DailyId     int64   `json:"daily_id" form:"daily_id" goe:"index;type:int"`
	NoteType    string  `json:"note_type" form:"note_type" goe:"type:varchar(50)"`
	NoteChinese *string `json:"note_chinese" form:"note_chinese" goe:"type:text"`
	NoteEnglish *string `json:"note_english" form:"note_english" goe:"type:text"`
}

// TableName WallNote的表名
func (*WallNote) TableName() string {
	return "t_wall_note"
}

// TableComment WallNote的备注
func (*WallNote) TableComment() string {
	return "墙纸小知识"
}
