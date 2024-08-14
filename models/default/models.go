package db

import (
	"time"

	xutils "github.com/azhai/xgen/utils"
)

// WallDaily 必应每日墙纸
type WallDaily struct {
	Id       int       `json:"id" form:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	OrigId   int       `json:"orig_id" form:"orig_id" xorm:"notnull default 0 comment('bing.wilii.cn原始ID') UNSIGNED INT(10)"`
	BingDate time.Time `json:"bing_date" form:"bing_date" xorm:"comment('必应的发布日期') unique DATE"`
	BingSku  string    `json:"bing_sku" form:"bing_sku" xorm:"notnull default '' comment('必应图片编号') index VARCHAR(100)"`
	Title    string    `json:"title" form:"title" xorm:"notnull default '' comment('简介') index VARCHAR(255)"`
	MaxDpi   string    `json:"max_dpi" form:"max_dpi" xorm:"notnull default '' comment('图片最大分辨率') VARCHAR(15)"`
}

func (*WallDaily) TableName() string {
	return "t_wall_daily"
}

// WallImage 墙纸图片
type WallImage struct {
	Id        int    `json:"id" form:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	DailyId   int    `json:"daily_id" form:"daily_id" xorm:"notnull default 0 comment('墙纸ID') index UNSIGNED INT(10)"`
	FileName  string `json:"file_name" form:"file_name" xorm:"notnull default '' comment('文件路径') VARCHAR(100)"`
	ImgMd5    string `json:"img_md5" form:"img_md5" xorm:"notnull default '' comment('图片MD5哈希') index CHAR(32)"`
	ImgSize   int    `json:"img_size" form:"img_size" xorm:"notnull default 0 comment('图片大小（单位：字节）') index UNSIGNED INT(10)"`
	ImgOffset int    `json:"img_offset" form:"img_offset" xorm:"notnull default 0 comment('图片在文件中偏移') UNSIGNED INT(10)"`
	ImgWidth  int    `json:"img_width" form:"img_width" xorm:"notnull default 0 comment('图片宽度') UNSIGNED MEDIUMINT(6)"`
	ImgHeight int    `json:"img_height" form:"img_height" xorm:"notnull default 0 comment('图片高度') UNSIGNED MEDIUMINT(6)"`
}

func (*WallImage) TableName() string {
	return "t_wall_image"
}

// WallNote 必应小知识
type WallNote struct {
	Id          int               `json:"id" form:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	DailyId     int               `json:"daily_id" form:"daily_id" xorm:"notnull default 0 comment('墙纸ID') index UNSIGNED INT(10)"`
	NoteType    string            `json:"note_type" form:"note_type" xorm:"notnull default '' comment('小知识类型') VARCHAR(50)"`
	NoteChinese xutils.NullString `json:"note_chinese" form:"note_chinese" xorm:"comment('中文描述') TEXT(65535)"`
	NoteEnglish xutils.NullString `json:"note_english" form:"note_english" xorm:"comment('英文描述') TEXT(65535)"`
}

func (*WallNote) TableName() string {
	return "t_wall_note"
}
