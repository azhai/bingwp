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
