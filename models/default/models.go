package db

import (
	"time"

	xutils "github.com/azhai/xgen/utils"
)

// ------------------------------------------------------------
// WallDaily 必应每日墙纸
// ------------------------------------------------------------
type WallDaily struct {
	Id       int               `json:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	Brief    string            `json:"brief" xorm:"notnull default '' comment('简介') index VARCHAR(255)"`
	MaxDpi   string            `json:"max_dpi" xorm:"notnull default '' comment('图片最大分辨率') VARCHAR(50)"`
	BingDate time.Time         `json:"bing_date" xorm:"comment('必应的发布日期') index DATE"`
	OrigId   int               `json:"orig_id" xorm:"notnull default 0 comment('原始ID') UNSIGNED INT(10)"`
	OrigUrl  xutils.NullString `json:"orig_url" xorm:"comment('缩略图原始地址') VARCHAR(300)"`
}

func (WallDaily) TableName() string {
	return "t_wall_daily"
}

// ------------------------------------------------------------
// WallImage 墙纸图片
// ------------------------------------------------------------
type WallImage struct {
	Id        int    `json:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	DailyId   int    `json:"daily_id" xorm:"notnull default 0 comment('墙纸ID') index UNSIGNED INT(10)"`
	FileExt   string `json:"file_ext" xorm:"notnull default '' comment('文件扩展名') VARCHAR(10)"`
	SaveDir   string `json:"save_dir" xorm:"notnull default '' comment('保存路径') VARCHAR(100)"`
	ImgMd5    string `json:"img_md5" xorm:"notnull default '' comment('图片MD5哈希') index CHAR(32)"`
	ImgSize   int    `json:"img_size" xorm:"notnull default 0 comment('图片大小（单位：字节）') index UNSIGNED INT(10)"`
	ImgOffset int    `json:"img_offset" xorm:"notnull default 0 comment('图片在文件中偏移') UNSIGNED INT(10)"`
	ImgWidth  int    `json:"img_width" xorm:"notnull default 0 comment('图片宽度') UNSIGNED MEDIUMINT(6)"`
	ImgHeight int    `json:"img_height" xorm:"notnull default 0 comment('图片高度') UNSIGNED MEDIUMINT(6)"`
}

func (WallImage) TableName() string {
	return "t_wall_image"
}

// ------------------------------------------------------------
// WallLocation 地理定位
// ------------------------------------------------------------
type WallLocation struct {
	Id        int     `json:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	DailyId   int     `json:"daily_id" xorm:"notnull default 0 comment('墙纸ID') index UNSIGNED INT(10)"`
	Geohash   string  `json:"geohash" xorm:"notnull default '' comment('GEO哈希') index VARCHAR(25)"`
	Latitude  float64 `json:"latitude" xorm:"notnull comment('纬度') FLOAT(9,6)"`
	Longitude float64 `json:"longitude" xorm:"notnull comment('经度') FLOAT(9,6)"`
	IsoCode   string  `json:"iso_code" xorm:"notnull default '' comment('国家代码') index CHAR(2)"`
	Country   string  `json:"country" xorm:"comment('国家') VARCHAR(100)"`
	City      string  `json:"city" xorm:"comment('城市') VARCHAR(255)"`
}

func (WallLocation) TableName() string {
	return "t_wall_location"
}

// ------------------------------------------------------------
// WallNote 必应小知识
// ------------------------------------------------------------
type WallNote struct {
	Id          int               `json:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	DailyId     int               `json:"daily_id" xorm:"notnull default 0 comment('墙纸ID') index UNSIGNED INT(10)"`
	NoteType    string            `json:"note_type" xorm:"notnull default '' comment('小知识类型') VARCHAR(50)"`
	NoteChinese xutils.NullString `json:"note_chinese" xorm:"comment('中文描述') TEXT(65535)"`
	NoteEnglish xutils.NullString `json:"note_english" xorm:"comment('英文描述') TEXT(65535)"`
}

func (WallNote) TableName() string {
	return "t_wall_note"
}

// ------------------------------------------------------------
// WallTag 墙纸标签
// ------------------------------------------------------------
type WallTag struct {
	Id      int    `json:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	DailyId int    `json:"daily_id" xorm:"notnull default 0 comment('墙纸ID') index UNSIGNED INT(10)"`
	TagName string `json:"tag_name" xorm:"notnull default '' comment('标签') index VARCHAR(100)"`
}

func (WallTag) TableName() string {
	return "t_wall_tag"
}
