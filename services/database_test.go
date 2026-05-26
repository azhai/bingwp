package services

import (
	"testing"

	"github.com/azhai/bingwp/models"
)

func TestInitDB(t *testing.T) {
	db, err := models.InitDB("file:goe?mode=memory&cache=shared", "")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer models.CloseDB()

	if db == nil {
		t.Fatal("db should not be nil")
	}
}

func TestInsertAndGetWallpaper(t *testing.T) {
	_, err := models.InitDB("file:goe?mode=memory&cache=shared", "")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer models.CloseDB()

	wp := &models.Wallpaper{
		Date:     "2026-04-01",
		Title:    "Test Title",
		Headline: "Test Headline",
		Slug:     "TestImage_ZH-CN123456",
	}

	err = InsertWallpaperIgnore(wp)
	if err != nil {
		t.Fatalf("InsertWallpaperIgnore failed: %v", err)
	}

	wallpapers, err := GetWallpapersByMonth(2026, 4)
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
