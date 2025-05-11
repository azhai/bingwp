package handlers

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	db "github.com/azhai/bingwp/models/default"
	"github.com/azhai/xgen/templater"
	xq "github.com/azhai/xgen/xquery"
	"github.com/gofiber/fiber/v3"
	"github.com/k0kubun/pp"
)

var (
	tpl = templater.NewFactory("./views", true).UpdateFuncs(template.FuncMap{
		"Date": func(dt time.Time) string {
			return dt.Format("2006-01-02")
		},
		"ImagePath": ImagePath,
		"ThumbPath": ThumbPath,
	})
)

// GetMonthBegin 获取月份的第一天零点零分
func GetMonthBegin(t time.Time) time.Time {
	yy, mm, _ := t.Date()
	loc := t.Location()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, loc)
}

// GetYearDoubleList 将年份分为左右两个列表
func GetYearDoubleList(max, min int) (lefts, rights []int) {
	if (max-min)%2 == 0 {
		max += 1
	}
	for i := max; i >= min; i -= 2 {
		rights = append(rights, i)
		lefts = append(lefts, i-1)
	}
	return
}

// PageHandler 首页
func PageHandler(ctx fiber.Ctx) (err error) {
	var dt time.Time
	yearMonth := ctx.Params("month")
	dt, err = time.Parse("200601", yearMonth)
	if err != nil || dt.After(time.Now()) {
		dt = time.Now()
	}
	monthBegin := GetMonthBegin(dt)
	nextBegin := GetMonthBegin(monthBegin.AddDate(0, 0, 31))
	where := xq.WithWhere("bing_date >= ? AND bing_date < ?",
		monthBegin.Format("2006-01-02"), nextBegin.Format("2006-01-02"))
	var rows []*db.WallDaily
	if err = db.Query(where).Asc("id").Find(&rows); err != nil {
		pp.Println(err)
	}
	rows = GetDailyNotes(GetDailyImages(rows))
	year, month := dt.Year(), fmt.Sprintf("%02d", int(dt.Month()))
	oddYears, evenYears := GetYearDoubleList(time.Now().Year(), 2009)
	data := fiber.Map{"Year": year, "Month": month, "CurrYear": monthBegin.Year(),
		"OddYears": oddYears, "EvenYears": evenYears, "Rows": rows}
	var body []byte
	if body, err = tpl.Render("home", data); err == nil {
		err = ctx.Type("html").Send(body)
	}
	return
}

// GetDailyImages 从数据库中加载每日图片的URL地址
func GetDailyImages(rows []*db.WallDaily) []*db.WallDaily {
	var ids []any
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	where := xq.WithRange("daily_id", ids...)
	var imgRows []*db.WallImage
	if err := db.Query(where).Find(&imgRows); err != nil {
		return rows
	}
	thumbs, images := make(map[int64]string), make(map[int64]string)
	for _, row := range imgRows {
		url := row.GetUrl()
		if strings.HasPrefix(row.FileName, "thumb") {
			thumbs[row.DailyId] = url
		} else {
			images[row.DailyId] = url
		}
	}
	for i, row := range rows {
		if url, ok := thumbs[row.Id]; ok {
			row.ThumbUrl = url
		}
		if url, ok := images[row.Id]; ok {
			row.ImageUrl = url
		}
		rows[i] = row
	}
	return rows
}

// GetDailyNotes 从数据库中加载每日图片的小知识
func GetDailyNotes(rows []*db.WallDaily) []*db.WallDaily {
	var ids []any
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	where := xq.WithRange("daily_id", ids...)
	var noteRows []*db.WallNote
	if err := db.Query(where).Find(&noteRows); err != nil {
		return rows
	}
	notes := make(map[int64]map[string]*db.WallNote)
	for _, row := range noteRows {
		if _, ok := notes[row.DailyId]; !ok {
			notes[row.DailyId] = map[string]*db.WallNote{
				"title": nil, "headline": nil, "description": nil,
			}
		}
		notes[row.DailyId][row.NoteType] = row
	}
	for i, row := range rows {
		if dict, ok := notes[row.Id]; ok {
			row.Notes = dict
		}
		rows[i] = row
	}
	return rows
}
