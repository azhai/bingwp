package services

import (
	"os"
	"strings"
)

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
