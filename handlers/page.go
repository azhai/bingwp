package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/azhai/bingwp/services/database"
	"github.com/azhai/gozzo/logging"
)

type Dict map[string]any

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
func PageHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	dt := time.Now()
	if ym := r.PathValue("month"); len(ym) >= 6 {
		dt, err = time.Parse("200601", ym)
	}
	if err != nil || dt.After(time.Now()) {
		dt = time.Now()
	}
	monthBegin := GetMonthBegin(dt)
	nextBegin := GetMonthBegin(monthBegin.AddDate(0, 0, 31))
	rows := database.GetMonthDailyRows(monthBegin, nextBegin)
	if len(rows) > 0 {
		rows = database.GetDailyNotes(database.GetDailyImages(rows))
	}
	year, month := dt.Year(), fmt.Sprintf("%02d", int(dt.Month()))
	oddYears, evenYears := GetYearDoubleList(time.Now().Year(), 2009)
	data := Dict{"Year": year, "Month": month, "CurrYear": monthBegin.Year(),
		"OddYears": oddYears, "EvenYears": evenYears, "Rows": rows}
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err = RenderTemplate(w, "home.tmpl", data); err != nil {
		logging.Error(err)
		http.Error(w, http.StatusText(500), 500)
	}
}

func RenderTemplate(w http.ResponseWriter, name string, data Dict) error {
	tmpl := template.New(name).Funcs(template.FuncMap{
		"Date": func(dt time.Time) string {
			return dt.Format("2006-01-02")
		},
	})
	tmpl = template.Must(tmpl.ParseFiles("./views/" + name))
	tmpl = template.Must(tmpl.ParseGlob("./views/sub_*.tmpl"))
	return tmpl.Execute(w, data)
}
