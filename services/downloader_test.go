package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadThumbnail(t *testing.T) {
	testImageData := []byte("fake-image-data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "15")
		w.Write(testImageData)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "test.jpg")

	fileSize, err := DownloadThumbnail(server.URL, localPath)
	if err != nil {
		t.Fatalf("DownloadThumbnail failed: %v", err)
	}

	if fileSize != int64(len(testImageData)) {
		t.Errorf("expected file size %d, got %d", len(testImageData), fileSize)
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		t.Error("file should exist after download")
	}
}

func TestGetFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	size := GetFileSize(testFile)
	if size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), size)
	}
}

func TestEnsureDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "subdir", "nested")

	err := EnsureDirectory(dirPath)
	if err != nil {
		t.Fatalf("EnsureDirectory failed: %v", err)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("path should be a directory")
	}
}
