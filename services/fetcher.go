package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
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
	MaxDownloadRetry       = 2
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

// FetchDetailData fetches the detail data for a wallpaper by GUID.
// Returns all fields (title, headline, copyright, description, filepath) for backfilling.
func FetchDetailData(guid string) (*models.DetailData, error) {
	url := fmt.Sprintf(DetailAPIURL, guid)
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch detail for %s: %w", guid, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d for detail %s", resp.StatusCode, guid)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read detail response: %w", err)
	}

	var result models.WiliiDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse detail JSON: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("detail API error: %s", result.Msg)
	}

	return &result.Response, nil
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

type DecodeFailureReason string

const (
	ReasonFormatUnsupported DecodeFailureReason = "format_unsupported"
	ReasonDataCorrupted     DecodeFailureReason = "data_corrupted"
	ReasonMemoryExceeded    DecodeFailureReason = "memory_exceeded"
	ReasonReadFailed        DecodeFailureReason = "read_failed"
	ReasonDownloadTruncated DecodeFailureReason = "download_truncated"
	ReasonUnknown           DecodeFailureReason = "unknown"
)

type DecodeFailure struct {
	Reason  DecodeFailureReason
	Err     error
	Path    string
	Attempt int
}

type LocalSkipResult int

const (
	SkipNone LocalSkipResult = iota
	SkipThumbOnly
	SkipFull
)

type LocalCheckInfo struct {
	FileSize int64
	ImageDPI string
}

type IntegrityStatus string

const (
	IntegrityIntact    IntegrityStatus = "intact"
	IntegrityTruncated IntegrityStatus = "truncated"
	IntegritySkipped   IntegrityStatus = "skipped"
)

type IntegrityCheckResult struct {
	Status       IntegrityStatus
	ExpectedSize int64
	ActualSize   int64
}

type DownloadSource string

const (
	SourceCDN  DownloadSource = "wilii_cdn"
	SourceBing DownloadSource = "bing"
	SourceUHD  DownloadSource = "bing_uhd"
)

type DownloadData struct {
	Body   []byte
	Source DownloadSource
	URL    string
}

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

func DownloadThumbnailForWallpaper(wp *models.Wallpaper) (DownloadResult, error) {
	return downloadAndProcess(wp, false)
}

// GetLocalImagePath returns the local image file path in ImageDir.
// Format: ./images/2014/20140501.jpg
func GetLocalImagePath(date string) string {
	relPath := GetThumbnailPath(date)
	return filepath.Join(ImageDir, relPath)
}

func checkLocalSkip(wp *models.Wallpaper, localImgPath string) (LocalSkipResult, LocalCheckInfo) {
	if !FileExists(localImgPath) {
		return SkipNone, LocalCheckInfo{}
	}

	size := GetFileSize(localImgPath)
	if size < MinThumbnailSize {
		return SkipNone, LocalCheckInfo{}
	}

	f, err := os.Open(localImgPath)
	if err != nil {
		return SkipNone, LocalCheckInfo{}
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return SkipNone, LocalCheckInfo{}
	}

	info := LocalCheckInfo{
		FileSize: size,
		ImageDPI: fmt.Sprintf("%dx%d", cfg.Width, cfg.Height),
	}

	needsRebuild, _ := NeedsRebuildThumbnail(wp)
	if needsRebuild {
		return SkipThumbOnly, info
	}

	return SkipFull, info
}

func generateThumbnailFromLocal(wp *models.Wallpaper, localImgPath string) error {
	relPath := GetThumbnailPath(wp.Date)
	thumbPath := filepath.Join(ThumbDir, relPath)

	f, err := os.Open(localImgPath)
	if err != nil {
		return fmt.Errorf("failed to open local image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode local image: %w", err)
	}

	thumbnail := resizeAndCrop(img)

	if err := EnsureDirectory(filepath.Dir(thumbPath)); err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	if err := saveThumbnail(thumbnail, thumbPath); err != nil {
		return err
	}

	return nil
}

func repairCorruptedImage(wp *models.Wallpaper, uhd bool, localImgPath string, failure *DecodeFailure) {
	strategy := "single_source_retry"
	if wp.Date < "2019-07-01" && !uhd {
		strategy = "multi_source_fallback"
	}

	if strings.Contains(failure.Err.Error(), "short Huffman data") {
		log.Printf("[Phase 3] Corrupted image detected [%s]: %s (truncated_type) (%s)",
			wp.Date, failure.Reason, failure.Err)
	} else {
		log.Printf("[Phase 3] Corrupted image detected [%s]: %s (%s)",
			wp.Date, failure.Reason, failure.Err)
	}

	os.Remove(localImgPath)

	body, err := downloadRemoteImage(wp, uhd)
	if err != nil {
		logRepairResult(wp.Date, failure.Reason, strategy, "abandoned", err)
		return
	}

	img, decodeFailure := decodeImage(body)
	if decodeFailure != nil {
		logRepairResult(wp.Date, failure.Reason, strategy, "abandoned", decodeFailure.Err)
		return
	}

	if err := EnsureDirectory(filepath.Dir(localImgPath)); err == nil {
		os.WriteFile(localImgPath, body, 0644)
	}

	thumbPath := filepath.Join(ThumbDir, GetThumbnailPath(wp.Date))
	if !FileExists(thumbPath) {
		thumbnail := resizeAndCrop(img)
		if err := EnsureDirectory(filepath.Dir(thumbPath)); err == nil {
			saveThumbnail(thumbnail, thumbPath)
		}
	}

	logRepairResult(wp.Date, failure.Reason, strategy, "repaired", nil)
}

func downloadAndProcess(wp *models.Wallpaper, uhd bool) (DownloadResult, error) {
	relPath := GetThumbnailPath(wp.Date)
	thumbPath := filepath.Join(ThumbDir, relPath)
	localImgPath := GetLocalImagePath(wp.Date)

	skipResult, localInfo := checkLocalSkip(wp, localImgPath)

	switch skipResult {
	case SkipFull:
		log.Printf("[Phase 3] Skipped [%s]: local image and thumbnail already exist", wp.Date)
		return DownloadResult{
			FileSize: localInfo.FileSize,
			ImageDPI: localInfo.ImageDPI,
		}, nil

	case SkipThumbOnly:
		log.Printf("[Phase 3] Skipped download [%s]: local image exists, generating thumbnail from local", wp.Date)
		if err := generateThumbnailFromLocal(wp, localImgPath); err != nil {
			log.Printf("[Phase 3] Failed to generate thumbnail from local [%s]: %v, falling back to download", wp.Date, err)
			repairCorruptedImage(wp, uhd, localImgPath, &DecodeFailure{
				Reason: classifyDecodeError(err),
				Err:    err,
				Path:   localImgPath,
			})
		} else {
			return DownloadResult{
				FileSize: localInfo.FileSize,
				ImageDPI: localInfo.ImageDPI,
			}, nil
		}
	}

	var result DownloadResult

	for attempt := 0; attempt <= MaxDownloadRetry; attempt++ {
		body, isLocal, readErr := fetchImageData(wp, uhd, localImgPath, &result)

		if readErr != nil && isLocal {
			logDecodeFailure(wp.Date, readErr, attempt, false)
			os.Remove(localImgPath)
			continue
		}

		if body == nil {
			if attempt < MaxDownloadRetry {
				continue
			}
			return DownloadResult{}, fmt.Errorf("download failed after %d attempts", attempt+1)
		}

		img, failure := decodeImage(body)
		if failure != nil {
			failure.Path = localImgPath
			failure.Attempt = attempt

			if isLocal {
				os.Remove(localImgPath)
				repairCorruptedImage(wp, uhd, localImgPath, failure)
				continue
			}

			if attempt < MaxDownloadRetry {
				logDecodeFailure(wp.Date, failure, attempt, false)
				continue
			}

			logDecodeFailure(wp.Date, failure, attempt, true)
			return DownloadResult{}, fmt.Errorf("failed to decode image: %s (%v) (attempts=%d)",
				failure.Reason, failure.Err, attempt+1)
		}

		imgBounds := img.Bounds()
		result.ImageDPI = fmt.Sprintf("%dx%d", imgBounds.Dx(), imgBounds.Dy())

		if !FileExists(localImgPath) {
			if err := EnsureDirectory(filepath.Dir(localImgPath)); err == nil {
				os.WriteFile(localImgPath, body, 0644)
			}
		}

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

	return DownloadResult{}, fmt.Errorf("failed after %d retries", MaxDownloadRetry)
}

func fetchImageData(wp *models.Wallpaper, uhd bool, localImgPath string, result *DownloadResult) ([]byte, bool, *DecodeFailure) {
	if FileExists(localImgPath) {
		body, err := os.ReadFile(localImgPath)
		if err != nil {
			return nil, true, &DecodeFailure{
				Reason: ReasonReadFailed,
				Err:    err,
				Path:   localImgPath,
			}
		}
		if len(body) > 0 {
			result.FileSize = int64(len(body))
			return body, true, nil
		}
		os.Remove(localImgPath)
	}

	body, err := downloadRemoteImage(wp, uhd)
	if err != nil {
		reason := ReasonUnknown
		if strings.Contains(err.Error(), "download truncated") {
			reason = ReasonDownloadTruncated
		}
		return nil, false, &DecodeFailure{
			Reason: reason,
			Err:    err,
		}
	}
	result.FileSize = int64(len(body))
	return body, false, nil
}

func downloadAndValidate(url string, source DownloadSource, date string) ([]byte, error) {
	data, contentLength, err := downloadImage(url)
	if err != nil {
		return nil, fmt.Errorf("download from %s failed: %w", source, err)
	}
	body := data.Body

	checkResult := validateDownloadIntegrity(body, contentLength)
	if checkResult.Status == IntegrityTruncated {
		logIntegrityCheck(date, checkResult, source, url)
		return nil, fmt.Errorf("download truncated: expected %d bytes, got %d bytes",
			checkResult.ExpectedSize, checkResult.ActualSize)
	}

	return body, nil
}

func downloadRemoteImage(wp *models.Wallpaper, uhd bool) ([]byte, error) {
	if uhd {
		uhdURL := GetBingImageURLHD(wp.Slug)
		if uhdURL == "" {
			return nil, fmt.Errorf("no UHD URL for slug: %s", wp.Slug)
		}
		return downloadAndValidate(uhdURL, SourceUHD, wp.Date)
	}

	if wp.Date >= "2019-07-01" {
		bingURL := GetBingImageURL(wp.Slug, "1920x1080")
		if bingURL != "" {
			return downloadAndValidate(bingURL, SourceBing, wp.Date)
		}
		return nil, fmt.Errorf("no download URL for slug: %s", wp.Slug)
	}

	sourceURL := BuildSourceURL(wp.Date, wp.Slug, wp.ImageDPI)
	bingURL := GetBingImageURL(wp.Slug, wp.ImageDPI)

	if sourceURL != "" {
		body, err := downloadAndValidate(sourceURL, SourceCDN, wp.Date)
		if err == nil {
			return body, nil
		}
		log.Printf("[Phase 3] CDN source failed [%s]: %v, falling back to Bing", wp.Date, err)
	}

	if bingURL != "" {
		body, err := downloadAndValidate(bingURL, SourceBing, wp.Date)
		if err == nil {
			return body, nil
		}
		log.Printf("[Phase 3] Bing source also failed [%s]: %v", wp.Date, err)
	}

	return nil, fmt.Errorf("download failed for %s: all sources unavailable or corrupted", wp.Slug)
}

func downloadImage(url string) (*DownloadData, int64, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read image data: %w", err)
	}

	contentLength := resp.ContentLength
	return &DownloadData{Body: body}, contentLength, nil
}

func classifyDecodeError(err error) DecodeFailureReason {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "unknown format"):
		return ReasonFormatUnsupported
	case strings.Contains(msg, "memory") || strings.Contains(msg, "cannot allocate"):
		return ReasonMemoryExceeded
	case strings.Contains(msg, "invalid") || strings.Contains(msg, "truncated") ||
		strings.Contains(msg, "corrupt") || strings.Contains(msg, "short Huffman data"):
		return ReasonDataCorrupted
	default:
		return ReasonUnknown
	}
}

func validateDownloadIntegrity(body []byte, contentLength int64) IntegrityCheckResult {
	actualSize := int64(len(body))
	if contentLength > 0 {
		if actualSize == contentLength {
			return IntegrityCheckResult{
				Status:       IntegrityIntact,
				ExpectedSize: contentLength,
				ActualSize:   actualSize,
			}
		}
		return IntegrityCheckResult{
			Status:       IntegrityTruncated,
			ExpectedSize: contentLength,
			ActualSize:   actualSize,
		}
	}
	return IntegrityCheckResult{
		Status:       IntegritySkipped,
		ExpectedSize: 0,
		ActualSize:   actualSize,
	}
}

func logIntegrityCheck(date string, result IntegrityCheckResult, source DownloadSource, url string) {
	diff := result.ExpectedSize - result.ActualSize
	log.Printf("[Phase 3] Download truncated [%s]: expected %d bytes, got %d bytes (diff=%d) source=%s url=%s",
		date, result.ExpectedSize, result.ActualSize, diff, source, url)
}

func logRepairResult(date string, triggerReason DecodeFailureReason, strategy string, outcome string, err error) {
	if outcome == "repaired" {
		log.Printf("[Phase 3] Image repaired [%s]: trigger=%s strategy=%s result=%s",
			date, triggerReason, strategy, outcome)
	} else {
		log.Printf("[Phase 3] Image repair abandoned [%s]: trigger=%s strategy=%s result=%s error=%v",
			date, triggerReason, strategy, outcome, err)
	}
}

func detectSmallOrCorruptFile(data []byte) (bool, DecodeFailureReason) {
	if len(data) == 0 {
		return true, ReasonDataCorrupted
	}
	if int64(len(data)) < MinThumbnailSize {
		return true, ReasonDataCorrupted
	}
	return false, ""
}

func logDecodeFailure(date string, failure *DecodeFailure, attempt int, abandoned bool) {
	detail := failure.Err.Error()
	maxAttempt := MaxDownloadRetry + 1
	if abandoned {
		log.Printf("[Phase 3] Failed HD [%s]: decode abandoned: %s (%s) (attempts=%d)",
			date, failure.Reason, detail, attempt+1)
	} else {
		log.Printf("[Phase 3] Failed HD [%s]: failed to decode image: %s (%s) (attempt=%d/%d)",
			date, failure.Reason, detail, attempt+1, maxAttempt)
	}
}

func decodeImage(data []byte) (image.Image, *DecodeFailure) {
	if corrupt, reason := detectSmallOrCorruptFile(data); corrupt {
		return nil, &DecodeFailure{
			Reason: reason,
			Err:    fmt.Errorf("file too small: %d bytes", len(data)),
		}
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		reason := classifyDecodeError(err)
		return nil, &DecodeFailure{
			Reason: reason,
			Err:    err,
		}
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
