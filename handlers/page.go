package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/bingwp/services"
)

type PageData struct {
	Year        int
	Month       int
	MonthString string
	Wallpapers  []*WallpaperView
	PrevMonth   string
	NextMonth   string
}

type WallpaperView struct {
	ID          int64
	Date        string
	Title       string
	FileSize    string
	LocalPath   string
	Description string
}

func PageHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	path := r.URL.Path
	if path != "/" && len(path) >= 8 {
		ymStr := path[1:]
		if len(ymStr) == 6 {
			if y, err := strconv.Atoi(ymStr[:4]); err == nil {
				if m, err := strconv.Atoi(ymStr[4:]); err == nil && m >= 1 && m <= 12 {
					year = y
					month = m
				}
			}
		}
	}

	wallpapers, err := services.GetWallpapersByMonth(db, year, month)
	if err != nil {
		log.Printf("Error fetching wallpapers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	views := convertToViews(wallpapers)

	data := PageData{
		Year:        year,
		Month:       month,
		MonthString: fmt.Sprintf("%04d-%02d", year, month),
		Wallpapers:  views,
		PrevMonth:   getPrevMonth(year, month),
		NextMonth:   getNextMonth(year, month, now),
	}

	renderTemplate(w, "index.html", data)
}

func convertToViews(wallpapers []*models.Wallpaper) []*WallpaperView {
	var views []*WallpaperView
	for _, wp := range wallpapers {
		view := &WallpaperView{
			ID:          wp.ID,
			Date:        wp.Date.Format("2006-01-02"),
			Title:       wp.Title,
			FileSize:    formatFileSize(wp.FileSize),
			LocalPath:   wp.LocalPath,
			Description: truncateDescription(wp.Description, 100),
		}
		views = append(views, view)
	}
	return views
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func truncateDescription(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return desc[:maxLen] + "..."
}

func getPrevMonth(year, month int) string {
	if month == 1 {
		return fmt.Sprintf("/%04d%02d", year-1, 12)
	}
	return fmt.Sprintf("/%04d%02d", year, month-1)
}

func getNextMonth(year, month int, now time.Time) string {
	next := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, now.Location())
	if next.After(now) {
		return ""
	}
	return fmt.Sprintf("/%04d%02d", next.Year(), next.Month())
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.ParseFiles("./views/" + name)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
