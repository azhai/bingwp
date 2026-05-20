package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/azhai/bingwp/models"
)

const DataAPIBaseURL = "https://bing.npanuhin.me/CN-zh.%s%s.json"

func FetchMonthData(year, month int) ([]*models.WallpaperRaw, error) {
	url := fmt.Sprintf(DataAPIBaseURL, fmt.Sprintf("%04d", year), fmt.Sprintf("%02d", month))
	return FetchMonthDataFromURL(url)
}

func FetchMonthDataFromURL(url string) ([]*models.WallpaperRaw, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rawData []*models.WallpaperRaw
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return rawData, nil
}

func ExtractBingFile(bingURL string) string {
	if !strings.Contains(bingURL, "id=") {
		return ""
	}

	parts := strings.Split(bingURL, "id=")
	if len(parts) < 2 {
		return ""
	}

	return parts[len(parts)-1]
}

func GenerateLocalPath(dateStr string) string {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return ""
	}

	yyyymm := parts[0] + parts[1]
	dd := parts[2]

	return fmt.Sprintf("images/%s/%s.jpg", yyyymm, dd)
}

func ConvertToWallpaper(raw *models.WallpaperRaw, fileSize int64) *models.Wallpaper {
	date, _ := time.Parse("2006-01-02", raw.Date)

	return &models.Wallpaper{
		Date:        date,
		Title:       raw.Title,
		Caption:     raw.Caption,
		Subtitle:    raw.Subtitle,
		Copyright:   raw.Copyright,
		Description: raw.Description,
		BingFile:    ExtractBingFile(raw.BingURL),
		FileSize:    fileSize,
		LocalPath:   GenerateLocalPath(raw.Date),
	}
}
