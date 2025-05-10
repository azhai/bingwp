package handlers

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/azhai/bingwp/services/orm"
	"github.com/azhai/xgen/templater"
	"github.com/go-goe/goe"
	"github.com/go-goe/goe/query/where"
	"github.com/gofiber/fiber/v3"
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
	orm.WallDaily
}

func GetMonthBegin(t time.Time) time.Time {
	yy, mm, _ := t.Date()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, t.Location())
}

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
	rows := queryDailyRows(monthBegin, nextBegin)
	infos := GetDailyImages(rows)
	year, month := dt.Year(), fmt.Sprintf("%02d", int(dt.Month()))
	oddYears, evenYears := GetYearDoubleList(time.Now().Year(), 2009)
	data := fiber.Map{"Year": year, "Month": month,
		"OddYears": oddYears, "EvenYears": evenYears, "Rows": infos}
	var body []byte
	if body, err = tpl.Render("home", data); err == nil {
		err = ctx.Type("html").Send(body)
	}
	return
}

func GetDailyImages(dayRows []orm.WallDaily) []*WallInfo {
	size := len(dayRows)
	ids, infos := make([]int64, size), make([]*WallInfo, size)
	for i, row := range dayRows {
		ids[i] = row.Id
		infos[i] = &WallInfo{WallDaily: row}
	}
	imgRows := queryDailyImages(ids)
	thumbs, images := make(map[string]string), make(map[string]string)
	for _, img := range imgRows {
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

func queryDailyRows(monthBegin, nextBegin time.Time) (rows []orm.WallDaily) {
	db, err := orm.Serv(), error(nil)
	rows, err = goe.Select(db.Daily).Where(
		where.And(
			where.GreaterEquals(&db.Daily.BingDate, monthBegin),
			where.Less(&db.Daily.BingDate, nextBegin),
		),
	).OrderByAsc(&db.Daily.Id).AsSlice()
	if err != nil {
		pp.Println(err)
		panic(err)
	}
	pp.Println("queryDailyRows", len(rows), monthBegin, nextBegin)
	return
}

func queryDailyImages(ids []int64) (rows []orm.WallImage) {
	if len(ids) == 0 {
		return
	}
	db, err := orm.Serv(), error(nil)
	rows, err = goe.Select(db.Image).Where(
		where.In(&db.Image.DailyId, ids),
	).AsSlice()
	if err != nil {
		pp.Println(err)
		panic(err)
	}
	pp.Println("queryDailyImages", len(rows), ids)
	return
}
