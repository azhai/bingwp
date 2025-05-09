package db

import (
	"time"

	xutils "github.com/azhai/xgen/utils"
)

// WallDaily 每日壁纸
type WallDaily struct {
	Id       int64     `json:"id" form:"id" xorm:"pk BIGINT"`
	Guid     string    `json:"guid" form:"guid" xorm:"notnull comment('bing.wilii.cn原始ID') CHAR(10)"`
	BingDate time.Time `json:"bing_date" form:"bing_date" xorm:"comment('必应的发布日期') unique DATE"`
	BingSku  string    `json:"bing_sku" form:"bing_sku" xorm:"notnull comment('必应图片编号') index VARCHAR(100)"`
	Title    string    `json:"title" form:"title" xorm:"notnull comment('标题') index VARCHAR(255)"`
	Headline string    `json:"headline" form:"headline" xorm:"notnull comment('简介') VARCHAR(255)"`
	Color    string    `json:"color" form:"color" xorm:"notnull comment('主色调') VARCHAR(15)"`
	MaxDpi   string    `json:"max_dpi" form:"max_dpi" xorm:"notnull comment('图片最大分辨率') VARCHAR(15)"`
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
	Id        int64  `json:"id" form:"id" xorm:"pk BIGINT"`
	DailyId   int64  `json:"daily_id" form:"daily_id" xorm:"notnull comment('墙纸ID') index BIGINT"`
	FileName  string `json:"file_name" form:"file_name" xorm:"notnull comment('文件路径') VARCHAR(100)"`
	ImgMd5    string `json:"img_md5" form:"img_md5" xorm:"notnull comment('图片MD5哈希') index CHAR(32)"`
	ImgSize   int64  `json:"img_size" form:"img_size" xorm:"notnull comment('图片大小（单位：字节）') index BIGINT"`
	ImgOffset int64  `json:"img_offset" form:"img_offset" xorm:"notnull comment('图片在文件中偏移') BIGINT"`
	ImgWidth  int    `json:"img_width" form:"img_width" xorm:"notnull comment('图片宽度') INTEGER"`
	ImgHeight int    `json:"img_height" form:"img_height" xorm:"notnull comment('图片高度') INTEGER"`
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
	Id          int64             `json:"id" form:"id" xorm:"pk BIGINT serial"`
	DailyId     int64             `json:"daily_id" form:"daily_id" xorm:"notnull comment('墙纸ID') index BIGINT"`
	NoteType    string            `json:"note_type" form:"note_type" xorm:"notnull comment('小知识类型') VARCHAR(50)"`
	NoteChinese xutils.NullString `json:"note_chinese" form:"note_chinese" xorm:"comment('中文描述') TEXT"`
	NoteEnglish xutils.NullString `json:"note_english" form:"note_english" xorm:"comment('英文描述') TEXT"`
}

// TableName WallNote的表名
func (*WallNote) TableName() string {
	return "t_wall_note"
}

// TableComment WallNote的备注
func (*WallNote) TableComment() string {
	return "墙纸小知识"
}
