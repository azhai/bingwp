# Bing Wallpaper 重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有的 Bing 壁纸展示项目重构为简洁现代的壁纸浏览应用，使用 SQLite3 存储，仅下载缩略图，支持增量更新。

**Architecture:** 采用 Go 标准库 + SQLite3 的轻量级架构，前端使用纯 HTML/CSS 实现瀑布流网格布局，后端通过 CLI 命令（update）和 HTTP 服务（serve）分离数据更新与展示逻辑。

**Tech Stack:** Go 1.21+, SQLite3 (mattn/go-sqlite3), HTML5, CSS3 Grid, net/http, encoding/json

---

## 文件结构映射

### 新建文件
```
bingwp/
├── models/
│   └── wallpaper.go          # 数据模型定义
├── services/
│   ├── database.go           # 数据库操作服务
│   ├── fetcher.go            # 数据获取服务
│   └── downloader.go         # 图片下载服务
├── handlers/
│   └── page.go               # 页面处理器（重写）
├── views/
│   ├── index.html            # 主页面模板（重写）
└── static/
    └── style.css             # 样式文件（新建）
```

### 修改文件
```
├── main.go                   # 入口文件（精简）
├── server.go                 # HTTP 服务器配置（简化）
├── update.go                 # update 命令（重写）
├── go.mod                    # 依赖管理（精简）
```

### 删除文件/目录
```
├── handlers/                 # 旧处理逻辑（大部分删除）
│   ├── crawl.go
│   ├── detail.go
│   ├── image.go
│   ├── parse.go
│   ├── task.go
│   └── downloader.go
├── services/db/              # 旧数据库层（全部删除）
│   ├── daily.go
│   ├── image.go
│   ├── note.go
│   ├── query.go
│   └── service.go
├── services/log/             # 日志服务（可选保留或删除）
├── views/                    # 旧模板（替换）
│   ├── home.tmpl
│   ├── sub_footer.tmpl
│   └── sub_header.tmpl
├── static/css/               # Tabler CSS（全部删除）
├── static/js/                # Tabler JS（全部删除）
├── static/libs/              # 第三方库（全部删除）
└── tests/                    # 旧测试（后续重建）
```

---

## Task 1: 项目结构重组与依赖精简

**Files:**
- Modify: `go.mod` - 移除不需要的依赖
- Delete: `handlers/crawl.go`, `handlers/detail.go`, `handlers/image.go`, `handlers/parse.go`, `handlers/task.go`, `handlers/downloader.go`
- Delete: `services/db/daily.go`, `services/db/image.go`, `services/db/note.go`, `services/db/query.go`, `services/db/service.go`
- Create: `models/wallpaper.go`

- [ ] **Step 1: 创建数据模型文件**

```go
// models/wallpaper.go
package models

import "time"

// Wallpaper 壁纸数据模型
type Wallpaper struct {
	ID          int64     `json:"id"`
	Date        time.Time `json:"date" db:"date"`
	Title       string    `json:"title" db:"title"`
	Caption     string    `json:"caption" db:"caption"`
	Subtitle    string    `json:"subtitle" db:"subtitle"`
	Copyright   string    `json:"copyright" db:"copyright"`
	Description string    `json:"description" db:"description"`
	BingFile    string    `json:"bing_file" db:"bing_file"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	LocalPath   string    `json:"local_path" db:"local_path"`
}

// WallpaperRaw 从 API 获取的原始数据结构
type WallpaperRaw struct {
	Title      string `json:"title"`
	Caption    string `json:"caption"`
	Subtitle   string `json:"subtitle"`
	Copyright  string `json:"copyright"`
	Description string `json:"description"`
	Date       string `json:"date"`       // 格式: "2026-04-01"
	BingURL    string `json:"bing_url"`   // 完整 URL
	URL        string `json:"url"`        // 缩略图 URL
}

// TableName 返回表名
func (*Wallpaper) TableName() string {
	return "wallpapers"
}
```

- [ ] **Step 2: 删除旧的处理器和数据库文件**

```bash
# 删除旧的 handlers
rm -f handlers/crawl.go handlers/detail.go handlers/image.go handlers/parse.go handlers/task.go handlers/downloader.go

# 删除旧的 services/db 目录
rm -rf services/db

# 删除旧的静态资源（Tabler 等）
rm -rf static/css static/js static/libs
rm -f static/logo*.svg

# 删除旧模板
rm -f views/home.tmpl views/sub_header.tmpl views/sub_footer.tmpl
```

- [ ] **Step 3: 更新 go.mod 文件**

```go
module github.com/azhai/bingwp

go 1.21

require (
	github.com/mattn/go-sqlite3 v1.14.22 // 仅保留 SQLite3 驱动
)
```

Run: `go mod tidy`
Expected: 成功清理依赖，仅保留必要包

- [ ] **Step 4: 提交更改**

```bash
git add -A
git commit -m "refactor: restructure project and simplify dependencies"
```

---

## Task 2: 数据模型与数据库层实现

**Files:**
- Create: `services/database.go`
- Test: `services/database_test.go`

- [ ] **Step 1: 编写数据库服务的单元测试**

```go
// services/database_test.go
package services

import (
	"os"
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
		Date:        testDate,
		Title:       "Test Title",
		Caption:     "Test Caption",
		BingFile:    "OHR.Test_ZH-CN123456_UHD.jpg",
		FileSize:    102400,
		LocalPath:   "images/202604/01.jpg",
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
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./services/ -v -run TestDatabase`
Expected: FAIL - 编译错误（database.go 不存在）

- [ ] **Step 3: 实现数据库服务**

```go
// services/database.go
package services

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/azhai/bingwp/models"
)

var db *sql.DB

// InitDB 初始化数据库连接并创建表
func InitDB(dbPath string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = createTable()
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

// createTable 创建 wallpapers 表
func createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS wallpapers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		caption TEXT DEFAULT '',
		subtitle TEXT DEFAULT '',
		copyright TEXT DEFAULT '',
		description TEXT DEFAULT '',
		bing_file TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		local_path TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_wallpapers_date ON wallpapers(date);
	`

	_, err := db.Exec(query)
	return err
}

// InsertWallpaper 插入一条壁纸记录（忽略重复）
func InsertWallpaper(database *sql.DB, wp *models.Wallpaper) error {
	query := `
	INSERT OR IGNORE INTO wallpapers 
	(date, title, caption, subtitle, copyright, description, bing_file, file_size, local_path)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	dateStr := wp.Date.Format("2006-01-02")
	_, err := database.Exec(query,
		dateStr,
		wp.Title,
		wp.Caption,
		wp.Subtitle,
		wp.Copyright,
		wp.Description,
		wp.BingFile,
		wp.FileSize,
		wp.LocalPath,
	)

	return err
}

// BatchInsertWallpapers 批量插入壁纸记录
func BatchInsertWallpapers(database *sql.DB, wallpapers []*models.Wallpaper) (int, error) {
	count := 0
	for _, wp := range wallpapers {
		err := InsertWallpaper(database, wp)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// GetWallpapersByMonth 获取指定月份的壁纸列表（按日期升序）
func GetWallpapersByMonth(database *sql.DB, year, month int) ([]*models.Wallpaper, error) {
	query := `
	SELECT id, date, title, caption, subtitle, copyright, description, 
	       bing_file, file_size, local_path
	FROM wallpapers 
	WHERE date LIKE ? 
	ORDER BY date ASC
	`

	datePrefix := fmt.Sprintf("%04d-%02d-", year, month)
	rows, err := database.Query(query, datePrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallpapers []*models.Wallpaper
	for rows.Next() {
		wp := &models.Wallpaper{}
		var dateStr string

		err := rows.Scan(
			&wp.ID,
			&dateStr,
			&wp.Title,
			&wp.Caption,
			&wp.Subtitle,
			&wp.Copyright,
			&wp.Description,
			&wp.BingFile,
			&wp.FileSize,
			&wp.LocalPath,
		)
		if err != nil {
			return nil, err
		}

		wp.Date, _ = time.Parse("2006-01-02", dateStr)
		wallpapers = append(wallpapers, wp)
	}

	return wallpapers, rows.Err()
}

// GetLastUpdateDate 获取最后更新的日期
func GetLastUpdateDate(database *sql.DB) (string, error) {
	query := `SELECT date FROM wallpapers ORDER BY date DESC LIMIT 1`

	var dateStr string
	err := database.QueryRow(query).Scan(&dateStr)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return dateStr, nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetDB 返回数据库实例
func GetDB() *sql.DB {
	return db
}
```

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./services/ -v -run TestDatabase`
Expected: PASS - 所有测试用例通过

- [ ] **Step 5: 提交**

```bash
git add services/database.go services/database_test.go models/wallpaper.go
git commit -m "feat: implement database layer with SQLite3"
```

---

## Task 3: 数据获取与解析服务

**Files:**
- Create: `services/fetcher.go`
- Test: `services/fetcher_test.go`

- [ ] **Step 1: 编写数据获取服务的测试**

```go
// services/fetcher_test.go
package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchMonthData(t *testing.T) {
	testData := `[{
		"title": "Test Wallpaper",
		"caption": "Test Caption",
		"subtitle": "Test Subtitle",
		"copyright": "© Test",
		"description": "Test Description",
		"date": "2026-04-01",
		"bing_url": "https://bing.com/th?id=OHR.Test_ZH-CN123456_UHD.jpg",
		"url": "https://example.com/thumb.jpg"
	}]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(testData))
	}))
	defer server.Close()

	data, err := FetchMonthDataFromURL(server.URL)
	if err != nil {
		t.Fatalf("FetchMonthData failed: %v", err)
	}

	if len(data) != 1 {
		t.Errorf("expected 1 item, got %d", len(data))
	}

	if data[0].Title != "Test Wallpaper" {
		t.Errorf("expected 'Test Wallpaper', got '%s'", data[0].Title)
	}
}

func TestExtractBingFile(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://bing.com/th?id=OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg",
			expected: "OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg",
		},
		{
			input:    "invalid-url",
			expected: "",
		},
	}

	for _, tt := range tests {
		result := ExtractBingFile(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractBingFile(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateLocalPath(t *testing.T) {
	tests := []struct {
		date     string
		expected string
	}{
		{"2026-04-01", "images/202604/01.jpg"},
		{"2026-12-31", "images/202612/31.jpg"},
	}

	for _, tt := range tests {
		result := GenerateLocalPath(tt.date)
		if result != tt.expected {
			t.Errorf("GenerateLocalPath(%q) = %q, want %q", tt.date, result, tt.expected)
		}
	}
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./services/ -v -run TestFetcher`
Expected: FAIL - fetcher.go 不存在

- [ ] **Step 3: 实现数据获取服务**

```go
// services/fetcher.go
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

const (
	DataAPIBaseURL = "https://bing.npanuhin.me/CN-zh.%s%s.json"
)

// FetchMonthData 从 API 获取指定月份的数据
func FetchMonthData(year, month int) ([]*models.WallpaperRaw, error) {
	url := fmt.Sprintf(DataAPIBaseURL, fmt.Sprintf("%04d", year), fmt.Sprintf("%02d", month))
	return FetchMonthDataFromURL(url)
}

// FetchMonthDataFromURL 从指定 URL 获取月度数据
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

// ExtractBingFile 从 bing_url 中提取文件标识
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

// GenerateLocalPath 生成本地存储路径
func GenerateLocalPath(dateStr string) string {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return ""
	}

	yyyymm := parts[0] + parts[1]
	dd := parts[2]

	return fmt.Sprintf("images/%s/%s.jpg", yyyymm, dd)
}

// ConvertToWallpaper 将原始数据转换为 Wallpaper 模型
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
```

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./services/ -v -run TestFetcher`
Expected: PASS - 所有测试通过

- [ ] **Step 5: 提交**

```bash
git add services/fetcher.go services/fetcher_test.go
git commit -m "feat: implement data fetching and parsing service"
```

---

## Task 4: 图片下载服务实现

**Files:**
- Create: `services/downloader.go`
- Test: `services/downloader_test.go`

- [ ] **Step 1: 编写下载服务的测试**

```go
// services/downloader_test.go
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
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testImageData)))
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
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./services/ -v -run TestDownloader`
Expected: FAIL - downloader.go 不存在

- [ ] **Step 3: 实现图片下载服务**

```go
// services/downloader.go
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

// DownloadThumbnail 下载缩略图到本地路径
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

// tryDownload 执行单次下载尝试
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

// GetFileSize 获取本地文件大小
func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// EnsureDirectory 确保目录存在（递归创建）
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

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
```

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./services/ -v -run TestDownloader`
Expected: PASS - 所有测试通过

- [ ] **Step 5: 提交**

```bash
git add services/downloader.go services/downloader_test.go
git commit -m "feat: implement thumbnail download service with retry logic"
```

---

## Task 5: Update 增量更新命令

**Files:**
- Rewrite: `update.go`
- Test: `update_test.go`

- [ ] **Step 1: 编写 update 命令的集成测试**

```go
// update_test.go
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
```

- [ ] **Step 2: 实现 Update 命令逻辑**

```go
// update.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/bingwp/services"
)

// UpdateCmd 更新数据命令
type UpdateCmd struct{}

// Run 执行增量更新
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
```

注意：需要在文件顶部添加 `"path/filepath"` 导入。

- [ ] **Step 3: 手动测试 update 命令**

Run:
```bash
export DB_PATH=./test_bingwp.db
export IMAGE_DIR=./test_images
go run . update
```

Expected: 成功下载数据到 test_images/ 并写入 test_bingwp.db

验证：
```bash
ls -la test_images/202604/
sqlite3 test_bingwp.db "SELECT COUNT(*) FROM wallpapers;"
sqlite3 test_bingwp.db "SELECT date, title FROM wallpapers ORDER BY date LIMIT 5;"
```

清理：
```bash
rm -rf test_images test_bingwp.db
unset DB_PATH IMAGE_DIR
```

- [ ] **Step 4: 提交**

```bash
git add update.go update_test.go
git commit -m "feat: implement incremental update command"
```

---

## Task 6: Web 服务与页面渲染

**Files:**
- Rewrite: `server.go`
- Rewrite: `main.go`
- Create: `handlers/page.go`
- Create: `views/index.html`

- [ ] **Step 1: 创建页面处理器**

```go
// handlers/page.go
package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/azhai/bingwp/services"
)

type PageData struct {
	Year        int
	Month       int
	MonthString string
	Wallpapers  []*services.WallpaperView
	PrevMonth   string
	NextMonth   string
}

type WallpaperView struct {
	ID          int64
	Date        string
	Title       string
	FileSize    string
	LocalPath   string
	Description string
}

// PageHandler 处理主页请求
func PageHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	path := r.URL.Path
	if path != "/" && len(path) >= 8 {
		ymStr := path[1:]
		if len(ymStr) == 6 {
			if y, err := strconv.Atoi(ymStr[:4]); err == nil {
				if m, err := strconv.Atoi(ymStr[4:]); err == nil && m >= 1 && m <= 12 {
					year = y
					month = m
				}
			}
		}
	}

	wallpapers, err := services.GetWallpapersByMonth(db, year, month)
	if err != nil {
		log.Printf("Error fetching wallpapers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	views := convertToViews(wallpapers)

	data := PageData{
		Year:        year,
		Month:       month,
		MonthString: fmt.Sprintf("%04d-%02d", year, month),
		Wallpapers:  views,
		PrevMonth:   getPrevMonth(year, month),
		NextMonth:   getNextMonth(year, month, now),
	}

	renderTemplate(w, "index.html", data)
}

func convertToViews(wallpapers []*models.Wallpaper) []*WallpaperView {
	var views []*WallpaperView
	for _, wp := range wallpapers {
		view := &WallpaperView{
			ID:          wp.ID,
			Date:        wp.Date.Format("2006-01-02"),
			Title:       wp.Title,
			FileSize:    formatFileSize(wp.FileSize),
			LocalPath:   wp.LocalPath,
			Description: truncateDescription(wp.Description, 100),
		}
		views = append(views, view)
	}
	return views
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func truncateDescription(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return desc[:maxLen] + "..."
}

func getPrevMonth(year, month int) string {
	if month == 1 {
		return fmt.Sprintf("/%04d%02d", year-1, 12)
	}
	return fmt.Sprintf("/%04d%02d", year, month-1)
}

func getNextMonth(year, month int, now time.Time) string {
	next := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, now.Location())
	if next.After(now) {
		return ""
	}
	return fmt.Sprintf("/%04d%02d", next.Year(), next.Month())
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.ParseFiles("./views/" + name)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
```

注意：需要添加 `"database/sql"` 和 `"github.com/azhai/bingwp/models"` 导入。

- [ ] **Step 2: 创建主页面模板（极简白底 + 瀑布流网格）**

```html
<!-- views/index.html -->
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bing Wallpaper - {{.MonthString}}</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header class="header">
            <h1>Bing Wallpaper</h1>
            <div class="month-nav">
                {{if .PrevMonth}}
                <a href="{{.PrevMonth}}" class="nav-link">← 上个月</a>
                {{end}}
                <span class="current-month">{{.Year}}年{{printf "%02d" .Month}}月</span>
                {{if .NextMonth}}
                <a href="{{.NextMonth}}" class="nav-link">下个月 →</a>
                {{end}}
            </div>
        </header>

        <main class="wallpaper-grid">
            {{range .Wallpapers}}
            <article class="wallpaper-card">
                <div class="card-image-wrapper">
                    <img src="/{{.LocalPath}}" alt="{{.Title}}" class="card-image" loading="lazy">
                </div>
                <div class="card-info">
                    <h3 class="card-title" title="{{.Title}}">{{.Title}}</h3>
                    <div class="card-meta">
                        <span class="card-date">{{.Date}}</span>
                        <span class="card-size">{{.FileSize}}</span>
                    </div>
                </div>
            </article>
            {{end}}
        </main>

        {{if not .Wallpapers}}
        <div class="empty-state">
            <p>该月份暂无壁纸数据</p>
            <p>请运行 <code>./bingwp update</code> 更新数据</p>
        </div>
        {{end}}

        <footer class="footer">
            <p>Bing Wallpaper © 2026 | 数据来源: <a href="https://bing.npanuhin.me" target="_blank">npanuhin.me</a></p>
        </footer>
    </div>
</body>
</html>
```

- [ ] **Step 3: 创建样式文件（极简白底设计）**

```css
/* static/style.css */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    background-color: #ffffff;
    color: #111827;
    line-height: 1.6;
}

.container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 20px;
}

.header {
    text-align: center;
    margin-bottom: 40px;
    padding-bottom: 20px;
    border-bottom: 1px solid #e5e7eb;
}

.header h1 {
    font-size: 32px;
    font-weight: 700;
    color: #111827;
    margin-bottom: 16px;
}

.month-nav {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 24px;
    font-size: 18px;
}

.current-month {
    font-weight: 600;
    color: #374151;
    min-width: 120px;
}

.nav-link {
    color: #2563eb;
    text-decoration: none;
    transition: color 0.2s;
}

.nav-link:hover {
    color: #1d4ed8;
    text-decoration: underline;
}

.wallpaper-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 24px;
    margin-bottom: 40px;
}

@media (max-width: 768px) {
    .wallpaper-grid {
        grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
        gap: 16px;
    }
}

@media (max-width: 480px) {
    .wallpaper-grid {
        grid-template-columns: 1fr;
    }
}

.wallpaper-card {
    background: #ffffff;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    transition: transform 0.2s, box-shadow 0.2s;
}

.wallpaper-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.12);
}

.card-image-wrapper {
    width: 100%;
    aspect-ratio: 16 / 9;
    overflow: hidden;
    background-color: #f3f4f6;
}

.card-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
}

.card-info {
    padding: 16px;
}

.card-title {
    font-size: 15px;
    font-weight: 600;
    color: #111827;
    line-height: 1.4;
    margin-bottom: 8px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.card-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    color: #6b7280;
}

.card-date {
    font-weight: 500;
}

.card-size {
    background-color: #f3f4f6;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 12px;
}

.empty-state {
    text-align: center;
    padding: 80px 20px;
    color: #6b7280;
}

.empty-state p {
    margin: 8px 0;
    font-size: 16px;
}

.empty-state code {
    background-color: #f3f4f6;
    padding: 4px 8px;
    border-radius: 4px;
    font-family: "Monaco", "Courier New", monospace;
    font-size: 14px;
    color: #dc2626;
}

.footer {
    text-align: center;
    padding: 24px 0;
    border-top: 1px solid #e5e7eb;
    color: #6b7280;
    font-size: 14px;
}

.footer a {
    color: #2563eb;
    text-decoration: none;
}

.footer a:hover {
    text-decoration: underline;
}
```

- [ ] **Step 4: 重写服务器配置**

```go
// server.go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/azhai/bingwp/services"
)

type ServerOpts struct {
	Port     int    `arg:"-p,--port" default:"8080" help:"服务端口"`
	DBPath   string `arg:"--db-path" help:"数据库路径"`
	ImageDir string `arg:"--image-dir" help:"图片目录"`
}

func NewServer(opts ServerOpts, db *sql.DB) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.FileServer(http.Dir("./")))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(opts.ImageDir))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.PageHandler(w, r, db)
	})

	addr := fmt.Sprintf(":%d", opts.Port)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
```

注意：需要添加 `"database/sql"` 和 `"github.com/azhai/bingwp/handlers"` 导入。

- [ ] **Step 5: 精简主入口文件**

```go
// main.go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/azhai/bingwp/services"
)

var args struct {
	Update    *UpdateCmd  `arg:"subcommand:update" help:"更新壁纸数据"`
	Serve     *ServeCmd   `arg:"subcommand:serve" help:"启动Web服务"`
	ServerOpts             `arg:"embed"`
}

type ServeCmd struct{}

func (c *ServeCmd) Run() {
	dbPath := args.DBPath
	if dbPath == "" {
		dbPath = "./bingwp.db"
	}

	imageDir := args.ImageDir
	if imageDir == "" {
		imageDir = "./images"
	}

	db, err := services.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	server := NewServer(ServerOpts{
		Port:     args.Port,
		DBPath:   dbPath,
		ImageDir: imageDir,
	}, db)

	addr := server.Addr
	fmt.Printf("🚀 Starting server at http://localhost%s\n", addr)
	fmt.Printf("📁 Image directory: %s\n", imageDir)
	fmt.Printf("💾 Database: %s\n", dbPath)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func main() {
	arg.MustParse(&args)

	switch {
	case args.Update != nil:
		args.Update.Run()
	case args.Serve != nil:
		args.Serve.Run()
	default:
		args.Serve = &ServeCmd{}
		args.Serve.Run()
	}
}
```

- [ ] **Step 6: 测试 Web 服务**

Run:
```bash
# 先初始化一些测试数据
export DB_PATH=./test_server.db
export IMAGE_DIR=./test_server_images
go run . update

# 启动服务
go run . serve --port 8888 --db-path ./test_server.db --image-dir ./test_server_images
```

在浏览器访问: `http://localhost:8888`
Expected: 显示当前月份的壁纸网格布局

测试路由:
- `http://localhost:8888/202604` - 查看4月份数据
- `http://localhost:8888/static/style.css` - 加载样式文件
- `http://localhost:8888/images/202604/01.jpg` - 加载图片

清理:
```bash
rm -rf test_server_images test_server.db
unset DB_PATH IMAGE_DIR
```

- [ ] **Step 7: 提交**

```bash
git add main.go server.go handlers/page.go views/index.html static/style.css
git commit -m "feat: implement web service with minimalist grid layout"
```

---

## Task 7: 前端优化与细节完善

**Files:**
- Modify: `views/index.html`
- Modify: `static/style.css`

- [ ] **Step 1: 添加响应式图片加载优化**

在 `index.html` 的 `<head>` 中添加:

```html
<meta name="theme-color" content="#ffffff">
<link rel="preconnect" href="https://fonts.googleapis.com">
```

在 `.card-image-wrapper` 中添加占位符效果：

```css
.card-image-wrapper::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg, #f3f4f6 25%, #e5e7eb 50%, #f3f4f6 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    z-index: 1;
}

@keyframes shimmer {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
}

.card-image-wrapper {
    position: relative;
}

.card-image {
    position: relative;
    z-index: 2;
}
```

- [ ] **Step 2: 添加平滑滚动和过渡动画**

```css
/* 在 style.css 末尾添加 */
html {
    scroll-behavior: smooth;
}

.wallpaper-card {
    cursor: pointer;
}

/* 点击卡片时的涟漪效果（可选） */
.wallpaper-card:active {
    transform: scale(0.98);
}

/* 图片加载错误时的样式 */
.card-image.error {
    opacity: 0.6;
}
```

在 `index.html` 底部添加 JavaScript:

```javascript
<script>
document.addEventListener('DOMContentLoaded', function() {
    const images = document.querySelectorAll('.card-image');
    
    images.forEach(img => {
        img.addEventListener('error', function() {
            this.classList.add('error');
            this.alt = '图片加载失败';
        });
        
        img.addEventListener('load', function() {
            this.style.opacity = '1';
        });
    });
});
</script>
```

- [ ] **Step 3: 添加 Favicon 和元信息**

创建简单的 SVG favicon 或使用 emoji 作为图标（可选）。

- [ ] **Step 4: 测试跨浏览器兼容性**

手动测试:
- Chrome/Edge (最新版)
- Firefox (最新版)
- Safari (macOS/iOS)
- 移动端浏览器（Chrome Mobile, Safari Mobile）

检查项:
- ✅ 网格布局正确显示
- ✅ 响应式断点正常工作
- ✅ 图片懒加载生效
- ✅ 悬停动画流畅
- ✅ 月份导航链接正确

- [ ] **Step 5: 提交**

```bash
git add views/index.html static/style.css
git commit -m "style: enhance UI with animations and responsive optimizations"
```

---

## Task 8: 最终测试与文档完善

**Files:**
- Create: `README.md` (如果需要)
- Test: 所有测试文件

- [ ] **Step 1: 运行完整测试套件**

Run:
```bash
go test ./... -v
```

Expected: 所有测试通过

- [ ] **Step 2: 端到端功能测试**

```bash
# 1. 清理环境
rm -rf bingwp.db images/

# 2. 初始化数据
./bingwp update
# 验证: 数据库创建成功，图片下载完成

# 3. 启动服务
./bingwp serve &
SERVER_PID=$!
sleep 2

# 4. 测试 API
curl -s http://localhost:8080/ | grep -q "Bing Wallpaper" && echo "✓ Homepage OK"
curl -s http://localhost:8080/202604 | grep -q "2026年04月" && echo "✓ Month page OK"
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/images/202604/01.jpg | grep -q "200" && echo "✓ Image serving OK"

# 5. 清理
kill $SERVER_PID
```

- [ ] **Step 3: 性能基准测试（可选）**

测试项目:
- 页面加载时间（首次/缓存后）
- 图片加载时间
- 数据库查询性能（大量数据时）

- [ ] **Step 4: 代码质量检查**

Run:
```bash
go vet ./...
golint ./...  # 如果安装了 golint
go fmt ./...
```

修复所有警告和建议。

- [ ] **Step 5: 最终提交并打标签**

```bash
git add -A
git commit -m "release: complete Bing Wallpaper redesign v2.0"

git tag -a v2.0 -m "Redesigned Bing Wallpaper with minimalist UI"
git push origin main --tags
```

---

## 附录: 快速开始指南

### 安装依赖

```bash
go mod tidy
```

### 首次使用

```bash
# 1. 下载历史数据
./bingwp update

# 2. 启动 Web 服务
./bingwp serve

# 3. 打开浏览器访问
# http://localhost:8080
```

### 日常维护

```bash
# 定期更新最新壁纸
./bingwp update

# 自定义端口启动
./bingwp serve --port 3000
```

### 环境变量

```bash
export DB_PATH=/path/to/bingwp.db      # 数据库路径
export IMAGE_DIR=/path/to/images        # 图片存储目录
```

---

## 实施注意事项

### 关键决策点

1. **SQLite vs 其他数据库**: 选择 SQLite 是因为单用户场景足够，无需额外部署数据库服务
2. **同步下载 vs 异步**: 采用同步下载以简化代码，后续可考虑引入 goroutine 池提升并发
3. **模板引擎**: 使用 Go 标准 `html/template` 足够简单场景，避免引入第三方依赖
4. **CSS Framework**: 纯手写 CSS 保持轻量，Grid 布局现代且兼容性好

### 已知限制

1. 不支持多语言切换（仅中文界面）
2. 无用户认证系统（纯本地应用）
3. 无搜索功能（可通过 SQL LIKE 查询扩展）
4. 图片无压缩处理（直接保存原始缩略图）

### 未来扩展方向

- [ ] 添加详情页（点击查看大图和完整描述）
- [ ] 支持按年份归档浏览
- [ ] 添加搜索功能
- [ ] 支持 RSS 输出
- [ ] Docker 化部署
- [ ] 添加 PWA 支持（离线浏览）

---

**Plan Status:** ✅ Complete
**Estimated Effort:** 4-6 hours for full implementation
**Risk Level:** Low (well-defined scope, clear requirements)
