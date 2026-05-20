package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxConcurrentDownloads = 5
	MaxRetries             = 3
	RetryDelay             = 2 * time.Second
)

func DownloadThumbnail(url, localPath string) (int64, error) {
	if url == "" || localPath == "" {
		return 0, fmt.Errorf("url or local path cannot be empty")
	}

	if err := EnsureDirectory(filepath.Dir(localPath)); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		fileSize, err := tryDownload(url, localPath)
		if err == nil {
			return fileSize, nil
		}

		lastErr = err
		if attempt < MaxRetries {
			time.Sleep(RetryDelay * time.Duration(attempt))
		}
	}

	return 0, fmt.Errorf("after %d retries, last error: %w", MaxRetries, lastErr)
}

func tryDownload(url, localPath string) (int64, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        MaxConcurrentDownloads,
			MaxIdleConnsPerHost: MaxConcurrentDownloads,
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, resp.Body)
	if err != nil {
		os.Remove(localPath)
		return 0, fmt.Errorf("failed to write file: %w", err)
	}

	return written, nil
}

func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func EnsureDirectory(dirPath string) error {
	if dirPath == "" {
		return nil
	}

	dirPath = strings.TrimRight(dirPath, "/\\")
	if dirPath == "." {
		return nil
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}

	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
