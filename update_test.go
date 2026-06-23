package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/bingwp/services"
	"github.com/azhai/goent/drivers"
)

func TestUpdateCmd_Run(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	os.Setenv("DB_PATH", dbPath)
	os.Setenv("IMAGE_DIR", tmpDir)
	os.Setenv("THUMB_DIR", tmpDir)
	defer func() {
		os.Unsetenv("DB_PATH")
		os.Unsetenv("IMAGE_DIR")
		os.Unsetenv("THUMB_DIR")
	}()

	cmd := &UpdateCmd{}

	start := time.Now()
	cmd.Run()
	duration := time.Since(start)

	t.Logf("Update completed in %v", duration)

	cfg := drivers.DatabaseConfig{
		Type: "sqlite",
		DSN:  dbPath,
	}
	_, err := models.OpenDB(cfg)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer models.CloseDB()

	_, latest, err := services.GetDateRange()
	if err != nil {
		t.Fatalf("GetDateRange failed: %v", err)
	}

	if latest == "" {
		t.Error("should have at least one wallpaper after update")
	}

	t.Logf("Last updated date: %s", latest)
}
