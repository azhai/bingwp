package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/azhai/bingwp/services"
)

type APIWallpaper struct {
	ID            int64   `json:"id"`
	GUID          string  `json:"guid"`
	Date          string  `json:"date"`
	Title         string  `json:"title"`
	Headline      string  `json:"headline"`
	Copyright     string  `json:"copyright"`
	Slug          string  `json:"slug"`
	Description   *string `json:"description"`
	ThumbnailPath string  `json:"thumbnailPath"`
	BingURL       string  `json:"bingUrl"`
}

type APIResponse struct {
	Year       int            `json:"year"`
	Month      int            `json:"month"`
	MonthLabel string         `json:"monthLabel"`
	PrevMonth  string         `json:"prevMonth,omitempty"`
	NextMonth  string         `json:"nextMonth,omitempty"`
	Wallpapers []APIWallpaper `json:"wallpapers"`
}

func APIWallpapersHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// 从查询参数读取 year 和 month
	if y := r.URL.Query().Get("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			month = v
		}
	}

	wallpapers, err := services.GetWallpapersByMonth(year, month)
	if err != nil {
		log.Printf("Error fetching wallpapers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	apiWPs := make([]APIWallpaper, 0, len(wallpapers))
	for _, wp := range wallpapers {
		apiWPs = append(apiWPs, APIWallpaper{
			ID:            wp.ID,
			GUID:          wp.GUID,
			Date:          wp.Date,
			Title:         wp.Title,
			Headline:      wp.Headline,
			Copyright:     wp.Copyright,
			Slug:          wp.Slug,
			Description:   wp.Description,
			ThumbnailPath: services.GetThumbnailPath(wp.Date),
			BingURL:       services.GetBingImageURL(wp.Slug, wp.ImageDPI),
		})
	}

	resp := APIResponse{
		Year:       year,
		Month:      month,
		MonthLabel: fmt.Sprintf("%04d年%02d月", year, month),
		PrevMonth:  getPrevMonth(year, month),
		NextMonth:  getNextMonth(year, month, now),
		Wallpapers: apiWPs,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
