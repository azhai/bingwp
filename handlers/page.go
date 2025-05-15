package handlers

import (
	"fmt"
	"text/template"
	"time"

	"github.com/azhai/bingwp/services/database"
	"github.com/azhai/xgen/templater"
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

// PageHandler 首页，按月显示壁纸
func PageHandler(ctx fiber.Ctx) (err error) {
	var dt time.Time
	yearMonth := ctx.Params("month")
	dt, err = time.Parse("200601", yearMonth)
	if err != nil || dt.After(time.Now()) {
		dt = time.Now()
	}
	monthBegin := GetMonthBegin(dt)
	nextBegin := GetMonthBegin(monthBegin.AddDate(0, 0, 31))
	rows := database.GetMonthDailyRows(monthBegin, nextBegin)
	rows = database.GetDailyNotes(database.GetDailyImages(rows))
	if len(rows) > 0 {
		pp.Println(rows[0])
	}
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
