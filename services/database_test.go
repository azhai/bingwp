package services

import (
	"testing"
	"time"

	"github.com/azhai/bingwp/models"
)

func TestInitDB(t *testing.T) {
	dbPath := ":memory:"
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("db should not be nil")
	}
}

func TestInsertAndGetWallpaper(t *testing.T) {
	dbPath := ":memory:"
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	testDate, _ := time.Parse("2006-01-02", "2026-04-01")
	wp := &models.Wallpaper{
		Date:      testDate,
		Title:     "Test Title",
		Caption:   "Test Caption",
		BingFile:  "OHR.Test_ZH-CN123456_UHD.jpg",
		FileSize:  102400,
		LocalPath: "images/202604/01.jpg",
	}

	err = InsertWallpaper(db, wp)
	if err != nil {
		t.Fatalf("InsertWallpaper failed: %v", err)
	}

	wallpapers, err := GetWallpapersByMonth(db, 2026, 4)
	if err != nil {
		t.Fatalf("GetWallpapersByMonth failed: %v", err)
	}

	if len(wallpapers) != 1 {
		t.Errorf("expected 1 wallpaper, got %d", len(wallpapers))
	}

	if wallpapers[0].Title != "Test Title" {
		t.Errorf("expected 'Test Title', got '%s'", wallpapers[0].Title)
	}
}

func TestGetLastUpdateDate(t *testing.T) {
	dbPath := ":memory:"
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	date, err := GetLastUpdateDate(db)
	if err != nil {
		t.Fatalf("GetLastUpdateDate failed: %v", err)
	}

	if date != "" {
		t.Errorf("expected empty date for empty db, got '%s'", date)
	}
}
