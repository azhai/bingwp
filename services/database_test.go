package services

import (
	"testing"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/goent/drivers"
)

func testDBConfig() drivers.DatabaseConfig {
	return drivers.DatabaseConfig{
		Type: "sqlite",
		DSN:  "file:goe?mode=memory&cache=shared",
	}
}

func TestInitDB(t *testing.T) {
	db, err := models.OpenDB(testDBConfig())
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer models.CloseDB()

	if db == nil {
		t.Fatal("db should not be nil")
	}
}

func TestInsertAndGetWallpaper(t *testing.T) {
	models.SetAutoMigrate(true)
	_, err := models.OpenDB(testDBConfig())
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
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
