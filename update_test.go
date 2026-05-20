package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/azhai/bingwp/services"
)

func TestUpdateCmd_Run(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	os.Setenv("DB_PATH", dbPath)
	os.Setenv("IMAGE_DIR", tmpDir)
	defer func() {
		os.Unsetenv("DB_PATH")
		os.Unsetenv("IMAGE_DIR")
	}()

	cmd := &UpdateCmd{}

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("UpdateCmd.Run failed: %v", err)
	}

	t.Logf("Update completed in %v", duration)

	db, err := services.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	lastDate, err := services.GetLastUpdateDate(db)
	if err != nil {
		t.Fatalf("GetLastUpdateDate failed: %v", err)
	}

	if lastDate == "" {
		t.Error("should have at least one wallpaper after update")
	}

	t.Logf("Last updated date: %s", lastDate)
}
