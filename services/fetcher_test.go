package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azhai/bingwp/models"
)

func TestFetchPageData(t *testing.T) {
	testData := `{
		"status": 200,
		"success": true,
		"msg": "获取成功",
		"response": {
			"page": 1,
			"pageCount": 124,
			"dataCount": 6154,
			"pageSize": 2,
			"data": [{
				"guid": "f38f7830c1",
				"date": "2026-05-20",
				"title": "Test Wallpaper",
				"copyright": "© Test Author",
				"headline": "Test Headline",
				"filepath": "/path/to/BumbleBee_ZH-CN6429376340_1920x1080.jpg"
			}]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(testData))
	}))
	defer server.Close()

	result, err := FetchPageData(1, 2)
	if err != nil {
		t.Fatalf("FetchPageData failed: %v", err)
	}

	if len(result.Response.Data) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Response.Data))
	}

	if result.Response.Data[0].GUID != "f38f7830c1" {
		t.Errorf("expected GUID 'f38f7830c1', got '%s'", result.Response.Data[0].GUID)
	}
}

func TestExtractSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://bing.wilii.cn/OneDrive/bingimages/2026/05/20/BumbleBee_ZH-CN6429376340_1920x1080.jpg",
			expected: "BumbleBee_ZH-CN6429376340",
		},
		{
			input:    "https://bing.wilii.cn/OneDrive/bingimages/2026/05/20/OHR.TestImage_ZH-CN123_1920x1080.jpg",
			expected: "TestImage_ZH-CN123",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		result := ExtractSlug(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractSlug(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractResolution(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://bing.wilii.cn/OneDrive/bingimages/2026/05/20/BumbleBee_ZH-CN6429376340_1920x1080.jpg",
			expected: "1920x1080",
		},
		{
			input:    "/path/to/Image_ZH-CN123_800x480.jpg",
			expected: "800x480",
		},
		{
			input:    "/path/to/Image_ZH-CN123.jpg",
			expected: "",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		result := ExtractResolution(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractResolution(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestConvertToWallpaper(t *testing.T) {
	raw := &models.WallpaperRaw{
		GUID:      "test123",
		Date:      "2026-05-20",
		Title:     "Test Title",
		Copyright: "© Author",
		Headline:  "Test Headline",
		Filepath:  "/path/TestImage_ZH-CN123_1920x1080.jpg",
	}

	wp := ConvertToWallpaper(raw)
	if wp.Title != "Test Title" {
		t.Errorf("expected 'Test Title', got '%s'", wp.Title)
	}

	if wp.Headline != "Test Headline" {
		t.Errorf("expected 'Test Headline', got '%s'", wp.Headline)
	}

	if wp.Slug != "TestImage_ZH-CN123" {
		t.Errorf("expected 'TestImage_ZH-CN123', got '%s'", wp.Slug)
	}

	if wp.ImageDPI != "1920x1080" {
		t.Errorf("expected '1920x1080', got '%s'", wp.ImageDPI)
	}
}
