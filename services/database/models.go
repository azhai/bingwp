package database

import (
	"strconv"
	"strings"
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

// ForeignValue WallDaily的外键
func (m *WallDaily) ForeignValue() any {
	return m.BingDate.Format("2006-01-02")
}

// ScanFrom 从src中读取数据写入当前对象
func (m *WallDaily) ScanFrom(src ScanSource, err error) error {
	if err == nil {
		err = src.Scan(&m.Id, &m.Guid, &m.BingDate,
			&m.BingSku, &m.Title, &m.Headline, &m.Color, &m.MaxDpi)
	}
	return err
}

func (m *WallDaily) SetId(id int64, err error) error {
	if err == nil {
		m.Id = id
	}
	return err
}

func (m *WallDaily) InsertSQL() string {
	return "INSERT " + m.TableName() + " (guid, bing_date, bing_sku, " +
		"title, headline, color, max_dpi) VALUES ($1, $2, $3, $4, $5, $6, $7)"
}

func (m *WallDaily) RowValues() []any {
	return []any{m.Guid, m.BingDate, m.BingSku, m.Title, m.Headline, m.Color, m.MaxDpi}
}

func (m *WallDaily) UpdateSQL() string {
	return "UPDATE " + m.TableName() + " SET guid=@guid, bing_date=@bing_date, " +
		"bing_sku=@bing_sku, title=@title, headline=@headline, color=@color, max_dpi=@max_dpi WHERE id=@id"
}

// WallDailyList 每日壁纸列表
type WallDailyList []*WallDaily

// GetIds 获取所有的主键ID
func (ms WallDailyList) GetIds() string {
	var ids []string
	for _, row := range ms {
		ids = append(ids, strconv.FormatInt(row.Id, 10))
	}
	if len(ids) == 0 {
		return "0"
	}
	return strings.Join(ids, ",")
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

// ForeignValue WallImage的外键
func (m *WallImage) ForeignValue() any {
	return m.DailyId
}

// ScanFrom 从src中读取数据写入当前对象
func (m *WallImage) ScanFrom(src ScanSource, err error) error {
	if err == nil {
		err = src.Scan(&m.Id, &m.DailyId, &m.FileName,
			&m.ImgMd5, &m.ImgSize, &m.ImgOffset, &m.ImgWidth, &m.ImgHeight)
	}
	return err
}

func (m *WallImage) Insert() error {
	table := m.TableName()
	sql := "INSERT " + table + " (daily_id, file_name, img_md5, img_size, img_offset, img_width, img_height) VALUES (?,?,?,?,?,?,?)"
	res, err := New().Exec(sql, m.DailyId, m.FileName, m.ImgMd5, m.ImgSize, m.ImgOffset, m.ImgWidth, m.ImgHeight)
	if err == nil {
		m.Id, err = res.LastInsertId()
	}
	return err
}

// WallNote 壁纸小知识
type WallNote struct {
	Id          int64      `json:"id" form:"id" db:"pk;serial;type:int"`
	DailyId     int64      `json:"daily_id" form:"daily_id" db:"index;type:int"`
	NoteType    string     `json:"note_type" form:"note_type" db:"type:varchar(50)"`
	NoteChinese NullString `json:"note_chinese" form:"note_chinese" db:"type:text"`
	NoteEnglish NullString `json:"note_english" form:"note_english" db:"type:text"`
}

// TableName WallNote的表名
func (*WallNote) TableName() string {
	return "t_wall_note"
}

// TableComment WallNote的备注
func (*WallNote) TableComment() string {
	return "壁纸小知识"
}

// ForeignValue WallNote的外键
func (m *WallNote) ForeignValue() any {
	return m.DailyId
}

// ScanFrom 从src中读取数据写入当前对象
func (m *WallNote) ScanFrom(src ScanSource, err error) error {
	if err == nil {
		err = src.Scan(&m.Id, &m.DailyId, &m.NoteType,
			&m.NoteChinese, &m.NoteEnglish)
	}
	return err
}

func (m *WallNote) Insert() error {
	table := m.TableName()
	sql := "INSERT " + table + " (daily_id, note_type, note_chinese, note_english) VALUES (?,?,?,?)"
	res, err := New().Exec(sql, m.DailyId, m.NoteType, m.NoteChinese, m.NoteEnglish)
	if err == nil {
		m.Id, err = res.LastInsertId()
	}
	return err
}
