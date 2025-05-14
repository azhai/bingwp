package db

import (
	"time"

	xutils "github.com/azhai/xgen/utils"
)

// WallDaily 每日壁纸
type WallDaily struct {
	Id       int64     `json:"id" form:"id" xorm:"pk BIGINT"`
	Guid     string    `json:"guid" form:"guid" xorm:"notnull CHAR(10) comment('bing.wilii.cn原始ID')"`
	BingDate time.Time `json:"bing_date" form:"bing_date" xorm:"unique DATE comment('必应的发布日期')"`
	BingSku  string    `json:"bing_sku" form:"bing_sku" xorm:"notnull index VARCHAR(100) comment('必应图片编号')"`
	Title    string    `json:"title" form:"title" xorm:"notnull index VARCHAR(255) comment('标题')"`
	Headline string    `json:"headline" form:"headline" xorm:"notnull VARCHAR(255) comment('简介')"`
	Color    string    `json:"color" form:"color" xorm:"notnull VARCHAR(15) comment('主色调')"`
	MaxDpi   string    `json:"max_dpi" form:"max_dpi" xorm:"notnull VARCHAR(15) comment('图片最大分辨率')"`

	WallDailyForeign `json:",inline" xorm:"-"`
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
	Id            int64 `json:"id" form:"id" xorm:"pk BIGINT"`
	DailyId       int64 `json:"daily_id" form:"daily_id" xorm:"notnull index BIGINT comment('壁纸ID')"`
	ImageUrlMixin `json:",inline" xorm:"extends"`
	ImgSize       int64 `json:"img_size" form:"img_size" xorm:"notnull index BIGINT comment('图片大小，单位：字节')"`
	ImgOffset     int64 `json:"img_offset" form:"img_offset" xorm:"notnull BIGINT comment('图片在文件中偏移')"`
	ImgWidth      int   `json:"img_width" form:"img_width" xorm:"notnull INTEGER comment('图片宽度')"`
	ImgHeight     int   `json:"img_height" form:"img_height" xorm:"notnull INTEGER comment('图片高度')"`
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
	Id          int64             `json:"id" form:"id" xorm:"pk autoincr unique BIGINT"`
	DailyId     int64             `json:"daily_id" form:"daily_id" xorm:"notnull BIGINT comment('壁纸ID')"`
	NoteType    string            `json:"note_type" form:"note_type" xorm:"notnull VARCHAR(50) comment('小知识类型')"`
	NoteChinese xutils.NullString `json:"note_chinese" form:"note_chinese" xorm:"TEXT comment('中文描述')"`
	NoteEnglish xutils.NullString `json:"note_english" form:"note_english" xorm:"TEXT comment('英文描述')"`
}

// TableName WallNote的表名
func (*WallNote) TableName() string {
	return "t_wall_note"
}

// TableComment WallNote的备注
func (*WallNote) TableComment() string {
	return "壁纸小知识"
}
