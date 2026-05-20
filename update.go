package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/bingwp/services"
)

type UpdateCmd struct{}

func (c *UpdateCmd) Run() {
	dbPath := getEnvOrDefault("DB_PATH", "./bingwp.db")
	imageDir := getEnvOrDefault("IMAGE_DIR", "./images")

	log.Printf("Starting update process...")
	log.Printf("Database path: %s", dbPath)
	log.Printf("Image directory: %s", imageDir)

	db, err := services.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	lastDate, err := services.GetLastUpdateDate(db)
	if err != nil {
		log.Fatalf("Failed to get last update date: %v", err)
	}

	today := time.Now()
	startDate := parseStartDate(lastDate)

	log.Printf("Updating from %s to %s...", startDate.Format("2006-01-02"), today.Format("2006-01-02"))

	totalUpdated := 0
	currentDate := startDate

	for !currentDate.After(today) {
		year, month, _ := currentDate.Date()

		rawData, err := services.FetchMonthData(year, int(month))
		if err != nil {
			log.Printf("Warning: Failed to fetch data for %04d-%02d: %v", year, month, err)
			currentDate = currentDate.AddDate(0, 1, 0)
			continue
		}

		newWallpapers := filterNewData(rawData, lastDate)

		for _, raw := range newWallpapers {
			localPath := filepath.Join(imageDir, services.GenerateLocalPath(raw.Date))

			if services.FileExists(localPath) {
				fileSize := services.GetFileSize(localPath)
				wp := services.ConvertToWallpaper(raw, fileSize)

				err = services.InsertWallpaper(db, wp)
				if err != nil {
					log.Printf("Warning: Failed to insert record for %s: %v", raw.Date, err)
					continue
				}

				totalUpdated++
				log.Printf("✓ [%s] %s (cached)", raw.Date, raw.Title)
			} else {
				fileSize, err := services.DownloadThumbnail(raw.URL, localPath)
				if err != nil {
					log.Printf("Warning: Failed to download thumbnail for %s: %v", raw.Date, err)
					continue
				}

				wp := services.ConvertToWallpaper(raw, fileSize)
				err = services.InsertWallpaper(db, wp)
				if err != nil {
					log.Printf("Warning: Failed to insert record for %s: %v", raw.Date, err)
					continue
				}

				totalUpdated++
				log.Printf("✓ [%s] %s (%d bytes)", raw.Date, raw.Title, fileSize)
			}
		}

		currentDate = currentDate.AddDate(0, 1, 0)
	}

	fmt.Printf("\n✅ Update completed! Total: %d new wallpapers\n", totalUpdated)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseStartDate(lastDate string) time.Time {
	if lastDate == "" {
		return time.Date(2009, 1, 1, 0, 0, 0, 0, time.Local)
	}

	date, err := time.Parse("2006-01-02", lastDate)
	if err != nil {
		log.Printf("Warning: Invalid last date format, using default start date")
		return time.Date(2009, 1, 1, 0, 0, 0, 0, time.Local)
	}

	return date
}

func filterNewData(rawData []*models.WallpaperRaw, lastDate string) []*models.WallpaperRaw {
	if lastDate == "" {
		return rawData
	}

	var filtered []*models.WallpaperRaw
	for _, raw := range rawData {
		if raw.Date > lastDate {
			filtered = append(filtered, raw)
		}
	}

	return filtered
}
