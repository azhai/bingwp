package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/azhai/bingwp/models"
	"github.com/disintegration/imaging"
)

var (
	ImageDir = "./images"
	ThumbDir = "./thumbs"
)

const (
	ListAPIBaseURL   = "https://api.wilii.cn/api/bing?page=%d&pageSize=%d"
	DetailAPIURL     = "https://api.wilii.cn/api/Bing/%s"
	BingImageBaseURL = "https://bing.com/th?id="
)

// Thumbnail dimensions
const (
	ThumbnailHeight        = 240
	ThumbnailWidth         = 400
	MinThumbnailSize int64 = 1024
)

// ============================================================
// Directory initialization
// ============================================================

// InitDirs creates image and thumbnail directories if they don't exist
func InitDirs(conf *Config) {
	if conf.ImageDir != "" {
		ImageDir = conf.ImageDir
	}
	if conf.ThumbDir != "" {
		ThumbDir = conf.ThumbDir
	}
	os.MkdirAll(ImageDir, 0755)
	os.MkdirAll(ThumbDir, 0755)
}

// ============================================================
// API data fetching
// ============================================================

// FetchPageData fetches a page of wallpaper list data from the API
func FetchPageData(page, pageSize int) (*models.WiliiListResponse, error) {
	url := fmt.Sprintf(ListAPIBaseURL, page, pageSize)
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d for page %d", resp.StatusCode, page)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result models.WiliiListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API error: %s", result.Msg)
	}

	return &result, nil
}

// FetchDetailData fetches the description for a wallpaper by GUID
func FetchDetailData(guid string) (string, error) {
	url := fmt.Sprintf(DetailAPIURL, guid)
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch detail for %s: %w", guid, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d for detail %s", resp.StatusCode, guid)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read detail response: %w", err)
	}

	var result models.WiliiDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse detail JSON: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("detail API error: %s", result.Msg)
	}

	return result.Response.Description, nil
}

// ============================================================
// Slug extraction and URL building
// ============================================================

// resolutionSuffixRe matches suffixes like _1920x1080, _800x480
var resolutionSuffixRe = regexp.MustCompile(`_(\d+)x(\d+)$`)

// ExtractSlug extracts the core image ID from a filepath string.
// It strips the OHR. prefix, resolution suffix, and extension, e.g.:
//
//	"/path/to/OHR.BumbleBee_ZH-CN6429376340_1920x1080.jpg"
//	=> "BumbleBee_ZH-CN6429376340"
func ExtractSlug(filepathStr string) string {
	if filepathStr == "" {
		return ""
	}
	base := filepath.Base(filepathStr)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	// Strip OHR. prefix
	name = strings.TrimPrefix(name, "OHR.")
	// Strip resolution suffix
	if m := resolutionSuffixRe.FindStringSubmatch(name); len(m) == 3 {
		name = strings.TrimSuffix(name, m[0])
	}
	return name
}

// ExtractResolution extracts the resolution from a filepath string, e.g.:
//
//	"/path/to/BumbleBee_ZH-CN6429376340_1920x1080.jpg"
//	=> "1920x1080"
func ExtractResolution(filepathStr string) string {
	if filepathStr == "" {
		return ""
	}
	base := filepath.Base(filepathStr)
	// Strip extension first so regex can match at end
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if m := resolutionSuffixRe.FindStringSubmatch(name); len(m) == 3 {
		return m[1] + "x" + m[2]
	}
	return ""
}

// ConvertToWallpaper converts raw API data to a Wallpaper model
func ConvertToWallpaper(raw *models.WallpaperRaw) *models.Wallpaper {
	return &models.Wallpaper{
		GUID:        raw.GUID,
		Date:        raw.Date,
		Title:       raw.Title,
		Headline:    raw.Headline,
		Copyright:   raw.Copyright,
		Description: &raw.Description,
		Slug:        ExtractSlug(raw.Filepath),
		ImageDPI:    ExtractResolution(raw.Filepath),
	}
}

// GetBingImageURL constructs the Bing image download URL from slug and resolution.
// Result: "https://bing.com/th?id=OHR.BumbleBee_ZH-CN6429376340_1920x1080.jpg"
// If imageDPI is larger than 1920x1080, uses "UHD" instead.
func GetBingImageURL(slug, imageDPI string) string {
	if slug == "" {
		return ""
	}
	if imageDPI == "" {
		imageDPI = "1920x1080"
	}
	// Use UHD for resolutions larger than 1920x1080
	if isUHDResolution(imageDPI) {
		imageDPI = "UHD"
	}
	return BingImageBaseURL + "OHR." + slug + "_" + imageDPI + ".jpg"
}

// isUHDResolution checks if the resolution string represents a UHD image
func isUHDResolution(dpi string) bool {
	if dpi == "UHD" {
		return true
	}
	parts := strings.SplitN(dpi, "x", 2)
	if len(parts) != 2 {
		return false
	}
	w, err1 := strconv.Atoi(parts[0])
	h, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return w > 1920 || h > 1080
}

// GetBingImageURLHD constructs the Bing image URL with highest quality (UHD)
func GetBingImageURLHD(slug string) string {
	if slug == "" {
		return ""
	}
	return BingImageBaseURL + "OHR." + slug + "_UHD.jpg"
}

// BuildSourceURL reconstructs the wilii source URL from date, slug and resolution
func BuildSourceURL(date, slug, imageDPI string) string {
	if date == "" || slug == "" || imageDPI == "" {
		return ""
	}
	if len(date) < 10 {
		return ""
	}
	// https://bing.wilii.cn/OneDrive/bingimages/2026/05/23/Name_ZH-CN123_1920x1080.jpg
	year := date[:4]
	month := date[5:7]
	day := date[8:10]
	return fmt.Sprintf("https://bing.wilii.cn/OneDrive/bingimages/%s/%s/%s/%s_%s.jpg",
		year, month, day, slug, imageDPI)
}

// ============================================================
// Thumbnail path and validation
// ============================================================

// GetThumbnailPath returns the relative thumbnail path: year/date.jpg
func GetThumbnailPath(date string) string {
	if len(date) < 7 {
		return ""
	}
	year := date[:4]
	day := strings.ReplaceAll(date, "-", "")
	return filepath.Join(year, day+".jpg")
}

// NeedsRebuildThumbnail checks if a thumbnail needs to be rebuilt.
// Returns true if: missing, wrong dimensions (not 480x360), or too small.
func NeedsRebuildThumbnail(wp *models.Wallpaper) (bool, string) {
	relPath := GetThumbnailPath(wp.Date)
	localPath := filepath.Join(ThumbDir, relPath)

	if !FileExists(localPath) {
		return true, "missing"
	}

	size := GetFileSize(localPath)
	if size < MinThumbnailSize {
		return true, "too_small"
	}

	f, err := os.Open(localPath)
	if err != nil {
		return true, "cannot_open"
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return true, "cannot_decode"
	}

	if cfg.Width != ThumbnailWidth || cfg.Height != ThumbnailHeight {
		return true, fmt.Sprintf("wrong_size (%dx%d, need %dx%d)",
			cfg.Width, cfg.Height, ThumbnailWidth, ThumbnailHeight)
	}

	return false, ""
}

// DeleteOldThumbnail removes an existing thumbnail file so it can be rebuilt
func DeleteOldThumbnail(wp *models.Wallpaper) {
	relPath := GetThumbnailPath(wp.Date)
	localPath := filepath.Join(ThumbDir, relPath)
	os.Remove(localPath)
}

// ============================================================
// Image download and processing
// ============================================================

// DownloadResult holds the result of a download operation
type DownloadResult struct {
	FileSize int64  // Original file size in bytes
	ImageDPI string // Actual resolution, e.g. "3840x2160"
}

// DownloadThumbnailHD downloads a UHD image, generates a thumbnail,
// and returns the original file size and actual resolution.
func DownloadThumbnailHD(wp *models.Wallpaper) (DownloadResult, error) {
	return downloadAndProcess(wp, true)
}

// DownloadThumbnailForWallpaper downloads an image and generates a thumbnail
func DownloadThumbnailForWallpaper(wp *models.Wallpaper) error {
	_, err := downloadAndProcess(wp, false)
	return err
}

// GetLocalImagePath returns the local image file path in ImageDir.
// Format: ./images/2014/20140501.jpg
func GetLocalImagePath(date string) string {
	relPath := GetThumbnailPath(date)
	return filepath.Join(ImageDir, relPath)
}

// downloadAndProcess downloads an image and generates a thumbnail.
// It first checks for a local image in ImageDir before downloading.
// When uhd=true, downloads the UHD image, saves it to ImageDir, and makes thumbnail.
//   - Even if thumbnail exists, still downloads UHD image if not in ImageDir.
//
// When uhd=false, uses the appropriate source based on date.
// Returns download result with file size and actual resolution.
func downloadAndProcess(wp *models.Wallpaper, uhd bool) (DownloadResult, error) {
	relPath := GetThumbnailPath(wp.Date)
	thumbPath := filepath.Join(ThumbDir, relPath)
	localImgPath := GetLocalImagePath(wp.Date)

	var body []byte
	var err error
	var result DownloadResult

	// First, try to read from local image directory
	if FileExists(localImgPath) {
		body, err = os.ReadFile(localImgPath)
		if err == nil {
			result.FileSize = int64(len(body))
		}
	}

	if len(body) == 0 {
		if uhd {
			// UHD mode: download high-res image directly
			uhdURL := GetBingImageURLHD(wp.Slug)
			if uhdURL == "" {
				return DownloadResult{}, fmt.Errorf("no UHD URL for slug: %s", wp.Slug)
			}
			body, err = downloadImage(uhdURL)
			if err != nil {
				return DownloadResult{}, fmt.Errorf("UHD download failed for %s: %w", wp.Slug, err)
			}
			result.FileSize = int64(len(body))
		} else if wp.Date >= "2019-07-01" {
			// After 2019-07-01 (non-UHD): download 1920x1080 from Bing
			bingURL := GetBingImageURL(wp.Slug, "1920x1080")
			if bingURL != "" {
				body, err = downloadImage(bingURL)
			}
			if err != nil {
				return DownloadResult{}, fmt.Errorf("download failed for %s: %w", wp.Slug, err)
			}
			result.FileSize = int64(len(body))
		} else {
			// Before 2019-07-01: use wilii source URL first, fallback to Bing
			sourceURL := BuildSourceURL(wp.Date, wp.Slug, wp.ImageDPI)
			bingURL := GetBingImageURL(wp.Slug, wp.ImageDPI)

			if sourceURL != "" {
				body, err = downloadImage(sourceURL)
			}
			if err != nil && bingURL != "" {
				body, err = downloadImage(bingURL)
			}
			if err != nil {
				return DownloadResult{}, fmt.Errorf("download failed for %s: %w", wp.Slug, err)
			}
			result.FileSize = int64(len(body))
		}
	}

	img, err := decodeImage(body)
	if err != nil {
		return DownloadResult{}, err
	}

	// Record actual resolution
	imgBounds := img.Bounds()
	result.ImageDPI = fmt.Sprintf("%dx%d", imgBounds.Dx(), imgBounds.Dy())

	// Save original image to ImageDir (if not already there)
	if !FileExists(localImgPath) {
		if err := EnsureDirectory(filepath.Dir(localImgPath)); err == nil {
			os.WriteFile(localImgPath, body, 0644)
		}
	}

	// Generate thumbnail only if it doesn't exist
	if !FileExists(thumbPath) {
		thumbnail := resizeAndCrop(img)

		if err := EnsureDirectory(filepath.Dir(thumbPath)); err != nil {
			return DownloadResult{}, fmt.Errorf("failed to create directory: %w", err)
		}

		if err := saveThumbnail(thumbnail, thumbPath); err != nil {
			return DownloadResult{}, err
		}
	}

	return result, nil
}

// downloadImage downloads an image from the given URL and returns the raw bytes
func downloadImage(url string) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return body, nil
}

// decodeImage decodes raw bytes into an image.Image
func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

// resizeAndCrop resizes the image to ThumbnailHeight (keeping aspect ratio)
// and crops to ThumbnailWidth from center
func resizeAndCrop(img image.Image) image.Image {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	ratio := float64(ThumbnailHeight) / float64(srcH)
	newWidth := int(float64(srcW) * ratio)
	if newWidth < 1 {
		newWidth = 1
	}

	resized := imaging.Resize(img, newWidth, ThumbnailHeight, imaging.Lanczos)

	if newWidth > ThumbnailWidth {
		resized = imaging.CropCenter(resized, ThumbnailWidth, ThumbnailHeight)
	}

	return resized
}

// saveThumbnail saves the processed image as a high-quality JPEG
func saveThumbnail(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail file: %w", err)
	}
	defer f.Close()

	// Quality 95 for near-lossless JPEG compression
	if err := imaging.Encode(f, img, imaging.JPEG, imaging.JPEGQuality(95)); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return nil
}
