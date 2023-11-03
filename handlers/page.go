package handlers

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"gitee.com/azhai/fiber-u8l/v2"
	db "github.com/azhai/bingwp/models/default"
	"github.com/azhai/xgen/templater"
	xq "github.com/azhai/xgen/xquery"
	"github.com/k0kubun/pp"
)

var (
	tpl = templater.NewFactory("./views").UpdateFuncs(template.FuncMap{
		"Date": func(dt time.Time) string {
			return dt.Format("2006-01-02")
		},
		"ImagePath": ImagePath,
		"ThumbPath": ThumbPath,
	})
)

type WallInfo struct {
	Thumb, Image string
	*db.WallDaily
}

func GetMonthBegin(obj time.Time) time.Time {
	return obj.AddDate(0, 0, 1-obj.Day())
}

// PageHandler 首页
func PageHandler(ctx *fiber.Ctx) (err error) {
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
	if err = db.Query(where).Asc("id").Find(&rows); err != nil {
		pp.Println(err)
	}
	month := fmt.Sprintf("%02d", int(dt.Month()))
	infos := GetDailyImages(rows)
	data := fiber.Map{"Year": dt.Year(), "Month": month, "Rows": infos}
	var body []byte
	if body, err = tpl.Render("home", data); err == nil {
		ctx.SetType("html")
		err = ctx.Send(body)
	}
	return
}

func GetDailyImages(rows []*db.WallDaily) []*WallInfo {
	size := len(rows)
	ids, infos := make([]any, size), make([]*WallInfo, size)
	for i, row := range rows {
		ids[i] = row.Id
		infos[i] = &WallInfo{WallDaily: row}
	}
	where := xq.WithRange("daily_id", ids...)
	var imgs []*db.WallImage
	if err := db.Query(where).Find(&imgs); err != nil {
		return infos
	}
	thumbs, images := make(map[string]string), make(map[string]string)
	for _, img := range imgs {
		pos := len(img.FileName) - len(".jpg")
		dt, ver := img.FileName[pos-8:pos], ""
		if len(img.ImgMd5) > 24 {
			ver = img.ImgMd5[24:]
		}
		url := fmt.Sprintf("%s?v=%s", img.FileName, ver)
		if strings.HasPrefix(img.FileName, "thumb") {
			thumbs[dt] = url
		} else {
			images[dt] = url
		}
	}
	for i, info := range infos {
		dt := info.BingDate.Format("20060102")
		if url, ok := thumbs[dt]; ok {
			info.Thumb = url
		}
		if url, ok := images[dt]; ok {
			info.Image = url
		}
		infos[i] = info
	}
	return infos
}
