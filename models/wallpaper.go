package models

import "time"

type Wallpaper struct {
	ID          int64     `json:"id"`
	Date        time.Time `json:"date" db:"date"`
	Title       string    `json:"title" db:"title"`
	Caption     string    `json:"caption" db:"caption"`
	Subtitle    string    `json:"subtitle" db:"subtitle"`
	Copyright   string    `json:"copyright" db:"copyright"`
	Description string    `json:"description" db:"description"`
	BingFile    string    `json:"bing_file" db:"bing_file"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	LocalPath   string    `json:"local_path" db:"local_path"`
}

type WallpaperRaw struct {
	Title      string `json:"title"`
	Caption    string `json:"caption"`
	Subtitle   string `json:"subtitle"`
	Copyright  string `json:"copyright"`
	Description string `json:"description"`
	Date       string `json:"date"`
	BingURL    string `json:"bing_url"`
	URL        string `json:"url"`
}

func (*Wallpaper) TableName() string {
	return "wallpapers"
}
