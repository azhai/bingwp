package handlers

import (
	"fmt"
	"text/template"
	"time"

	"github.com/azhai/bingwp/cmd"

	db "github.com/azhai/bingwp/models/default"

	"gitee.com/azhai/fiber-u8l/v2"
	"github.com/azhai/xgen/templater"
	xq "github.com/azhai/xgen/xquery"
)

var (
	tpl = templater.NewFactory("./views").UpdateFuncs(template.FuncMap{
		"Date": func(dt time.Time) string {
			return dt.Format("2006-01-02")
		},
		"ImgPath": func(dt time.Time) string {
			return dt.Format("200601/20060102")
		},
	})
)

func GetMonthBegin(obj time.Time) time.Time {
	return obj.AddDate(0, 0, 1-obj.Day())
}

// HomeHandler 首页
func HomeHandler(ctx *fiber.Ctx) (err error) {
	var dt time.Time
	yearMonth := ctx.ParamStr("month")
	dt, err = time.Parse("200601", yearMonth)
	if err != nil || dt.After(time.Now()) {
		dt = time.Now()
	}
	monthBegin := GetMonthBegin(dt)
	nextBegin := GetMonthBegin(monthBegin.AddDate(0, 0, 31))
	where := xq.WithWhere("bing_date >= ? AND bing_date < ?",
		monthBegin.Format("2006-01-02"), nextBegin.Format("2006-01-02"))
	var rows []*db.WallDaily
	err = db.Query(where).Desc("id").Find(&rows)
	month := fmt.Sprintf("%02d", int(dt.Month()))
	data := fiber.Map{"Rows": rows, "Year": dt.Year(), "Month": month}
	var body []byte
	if body, err = tpl.Render("home", data); err == nil {
		ctx.SetType("html")
		err = ctx.Send(body)
	}
	return
}

// DataHandler 数据
func DataHandler(ctx *fiber.Ctx) (err error) {
	data := make(fiber.Map)
	err = ctx.Reply(data)
	return
}

// logErrorIf 记录错误到日志
func logErrorIf(err error) {
	logger := cmd.GetDefaultLogger()
	if logger != nil && err != nil {
		logger.Error(err)
	}
}
