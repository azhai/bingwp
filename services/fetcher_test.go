package services

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestClassifyDecodeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected DecodeFailureReason
	}{
		{"unknown format", errors.New("image: unknown format"), ReasonFormatUnsupported},
		{"memory error", errors.New("runtime: out of memory"), ReasonMemoryExceeded},
		{"cannot allocate", errors.New("cannot allocate memory"), ReasonMemoryExceeded},
		{"invalid data", errors.New("jpeg: invalid format"), ReasonDataCorrupted},
		{"truncated data", errors.New("jpeg: truncated data"), ReasonDataCorrupted},
		{"corrupt data", errors.New("png: corrupt file"), ReasonDataCorrupted},
		{"short huffman data", errors.New("invalid JPEG format: short Huffman data"), ReasonDataCorrupted},
		{"other error", errors.New("some other error"), ReasonUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyDecodeError(tt.err)
			if result != tt.expected {
				t.Errorf("classifyDecodeError(%q) = %q, want %q", tt.err.Error(), result, tt.expected)
			}
		})
	}
}

func TestDetectSmallOrCorruptFile(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		wantCorrupt bool
		wantReason  DecodeFailureReason
	}{
		{"zero bytes", []byte{}, true, ReasonDataCorrupted},
		{"small file", make([]byte, 100), true, ReasonDataCorrupted},
		{"normal size", make([]byte, 2048), false, DecodeFailureReason("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corrupt, reason := detectSmallOrCorruptFile(tt.data)
			if corrupt != tt.wantCorrupt {
				t.Errorf("detectSmallOrCorruptFile() corrupt = %v, want %v", corrupt, tt.wantCorrupt)
			}
			if reason != tt.wantReason {
				t.Errorf("detectSmallOrCorruptFile() reason = %q, want %q", reason, tt.wantReason)
			}
		})
	}
}

func createTestJPEGBytes() []byte {
	var buf bytes.Buffer
	img := createTestImage(200, 200)
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

func createTestImage(w, h int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func TestDecodeImage(t *testing.T) {
	t.Run("valid jpeg", func(t *testing.T) {
		data := createTestJPEGBytes()
		img, failure := decodeImage(data)
		if failure != nil {
			t.Errorf("expected success, got failure: %v", failure.Err)
		}
		if img == nil {
			t.Error("expected image, got nil")
		}
	})

	t.Run("zero bytes", func(t *testing.T) {
		img, failure := decodeImage([]byte{})
		if failure == nil {
			t.Error("expected failure for zero bytes")
		} else if failure.Reason != ReasonDataCorrupted {
			t.Errorf("expected ReasonDataCorrupted, got %q", failure.Reason)
		}
		if img != nil {
			t.Error("expected nil image")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		data := make([]byte, 2048)
		copy(data, []byte("this is not an image"))
		img, failure := decodeImage(data)
		if failure == nil {
			t.Error("expected failure for invalid format")
		} else if failure.Reason != ReasonFormatUnsupported {
			t.Errorf("expected ReasonFormatUnsupported, got %q", failure.Reason)
		}
		if img != nil {
			t.Error("expected nil image")
		}
	})

	t.Run("truncated jpeg", func(t *testing.T) {
		data := createTestJPEGBytes()
		if len(data) <= int(MinThumbnailSize) {
			t.Fatalf("test JPEG too small: %d bytes", len(data))
		}
		truncated := data[:len(data)/2]
		img, failure := decodeImage(truncated)
		if failure == nil {
			t.Error("expected failure for truncated data")
		}
		if img != nil {
			t.Error("expected nil image")
		}
	})
}

func TestFetchImageDataReadFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bingwp_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "test*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	os.Chmod(tmpPath, 0000)
	defer os.Chmod(tmpPath, 0644)

	wp := &models.Wallpaper{Slug: "test", Date: "2026-05-20"}
	var result DownloadResult
	_, isLocal, readErr := fetchImageData(wp, false, tmpPath, &result)

	if !isLocal {
		t.Error("expected isLocal=true for local file")
	}
	if readErr == nil {
		t.Error("expected readErr for permission denied file")
	} else if readErr.Reason != ReasonReadFailed {
		t.Errorf("expected ReasonReadFailed, got %q", readErr.Reason)
	}
}

func TestCheckLocalSkip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bingwp_skip_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldImageDir := ImageDir
	oldThumbDir := ThumbDir
	ImageDir = filepath.Join(tmpDir, "images")
	ThumbDir = filepath.Join(tmpDir, "thumbs")
	defer func() {
		ImageDir = oldImageDir
		ThumbDir = oldThumbDir
	}()

	os.MkdirAll(ImageDir, 0755)
	os.MkdirAll(ThumbDir, 0755)

	wp := &models.Wallpaper{Date: "2026-05-20", Slug: "test"}
	localImgPath := GetLocalImagePath(wp.Date)

	t.Run("image not exist", func(t *testing.T) {
		result, _ := checkLocalSkip(wp, localImgPath)
		if result != SkipNone {
			t.Errorf("expected SkipNone, got %d", result)
		}
	})

	t.Run("image too small", func(t *testing.T) {
		EnsureDirectory(filepath.Dir(localImgPath))
		os.WriteFile(localImgPath, make([]byte, 100), 0644)
		defer os.Remove(localImgPath)

		result, _ := checkLocalSkip(wp, localImgPath)
		if result != SkipNone {
			t.Errorf("expected SkipNone, got %d", result)
		}
	})

	t.Run("image decode config fail", func(t *testing.T) {
		EnsureDirectory(filepath.Dir(localImgPath))
		badData := make([]byte, 2048)
		copy(badData, []byte("not an image"))
		os.WriteFile(localImgPath, badData, 0644)
		defer os.Remove(localImgPath)

		result, _ := checkLocalSkip(wp, localImgPath)
		if result != SkipNone {
			t.Errorf("expected SkipNone, got %d", result)
		}
	})

	t.Run("valid image no thumbnail", func(t *testing.T) {
		EnsureDirectory(filepath.Dir(localImgPath))
		jpegData := createTestJPEGBytes()
		os.WriteFile(localImgPath, jpegData, 0644)
		defer os.Remove(localImgPath)

		result, info := checkLocalSkip(wp, localImgPath)
		if result != SkipThumbOnly {
			t.Errorf("expected SkipThumbOnly, got %d", result)
		}
		if info.FileSize == 0 {
			t.Error("expected FileSize > 0 for SkipThumbOnly")
		}
	})

	t.Run("valid image wrong thumbnail size", func(t *testing.T) {
		EnsureDirectory(filepath.Dir(localImgPath))
		jpegData := createTestJPEGBytes()
		os.WriteFile(localImgPath, jpegData, 0644)
		defer os.Remove(localImgPath)

		thumbRelPath := GetThumbnailPath(wp.Date)
		thumbPath := filepath.Join(ThumbDir, thumbRelPath)
		EnsureDirectory(filepath.Dir(thumbPath))
		wrongThumb := createTestJPEGBytes()
		os.WriteFile(thumbPath, wrongThumb, 0644)
		defer os.Remove(thumbPath)

		result, info := checkLocalSkip(wp, localImgPath)
		if result != SkipThumbOnly {
			t.Errorf("expected SkipThumbOnly, got %d", result)
		}
		if info.FileSize == 0 {
			t.Error("expected FileSize > 0 for SkipThumbOnly")
		}
	})

	t.Run("valid image and thumbnail", func(t *testing.T) {
		EnsureDirectory(filepath.Dir(localImgPath))
		jpegData := createTestJPEGBytes()
		os.WriteFile(localImgPath, jpegData, 0644)
		defer os.Remove(localImgPath)

		thumbRelPath := GetThumbnailPath(wp.Date)
		thumbPath := filepath.Join(ThumbDir, thumbRelPath)
		EnsureDirectory(filepath.Dir(thumbPath))

		thumbImg := createTestImage(ThumbnailWidth, ThumbnailHeight)
		var thumbBuf bytes.Buffer
		jpeg.Encode(&thumbBuf, thumbImg, &jpeg.Options{Quality: 95})
		os.WriteFile(thumbPath, thumbBuf.Bytes(), 0644)
		defer os.Remove(thumbPath)

		result, info := checkLocalSkip(wp, localImgPath)
		if result != SkipFull {
			t.Errorf("expected SkipFull, got %d", result)
		}
		if info.FileSize == 0 {
			t.Error("expected FileSize > 0")
		}
		if info.ImageDPI == "" {
			t.Error("expected ImageDPI not empty")
		}
	})
}

func TestValidateDownloadIntegrity(t *testing.T) {
	t.Run("content length matches", func(t *testing.T) {
		result := validateDownloadIntegrity(make([]byte, 1000), 1000)
		if result.Status != IntegrityIntact {
			t.Errorf("expected IntegrityIntact, got %s", result.Status)
		}
	})

	t.Run("content length mismatch", func(t *testing.T) {
		result := validateDownloadIntegrity(make([]byte, 500), 1000)
		if result.Status != IntegrityTruncated {
			t.Errorf("expected IntegrityTruncated, got %s", result.Status)
		}
		if result.ExpectedSize != 1000 {
			t.Errorf("expected ExpectedSize=1000, got %d", result.ExpectedSize)
		}
		if result.ActualSize != 500 {
			t.Errorf("expected ActualSize=500, got %d", result.ActualSize)
		}
	})

	t.Run("content length zero", func(t *testing.T) {
		result := validateDownloadIntegrity(make([]byte, 1000), 0)
		if result.Status != IntegritySkipped {
			t.Errorf("expected IntegritySkipped, got %s", result.Status)
		}
	})

	t.Run("content length negative", func(t *testing.T) {
		result := validateDownloadIntegrity(make([]byte, 1000), -1)
		if result.Status != IntegritySkipped {
			t.Errorf("expected IntegritySkipped, got %s", result.Status)
		}
	})
}
