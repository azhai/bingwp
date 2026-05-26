package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/bingwp/services"
)

const PageSize = 50

type UpdateCmd struct {
	Workers int  `arg:"-w,--workers,env:WORKERS" help:"Number of download workers" default:"8"`
	UHD     bool `arg:"--uhd" help:"Download UHD images and update image_dpi & file_size"`
}

func (c *UpdateCmd) Run() {
	conf := loadConfig()
	services.InitDirs(conf)
	setupFileLogger(conf)
	log.Printf("Starting update process...")
	log.Printf("Database path: %s", conf.DBDSN)

	err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer models.CloseDB()

	// Phase 1: Fetch list data
	phase1Count := phase1FetchListData()

	// Phase 2: Update descriptions by GUID
	phase2Count := phase2UpdateDescriptions()

	// Phase 3: Download thumbnails
	phase3Count := phase3DownloadThumbnails(c.Workers, c.UHD)

	printReport(phase1Count, phase2Count, phase3Count)
}

// ============================================================
// Phase 1: Fetch list data
// ============================================================

func phase1FetchListData() int {
	log.Printf("[Phase 1] Checking database for existing data...")

	recent, err := services.GetRecentWallpapers(PageSize)
	if err != nil {
		log.Printf("Error querying recent wallpapers: %v", err)
		recent = nil
	}

	if len(recent) == 0 {
		log.Printf("[Phase 1] Database is empty. Fetching all pages from start...")
		return fetchAllPagesFromStart()
	}

	dates := extractDates(recent)
	gapDate := findEarliestGap(dates)

	if gapDate == "" {
		latestDate := recent[0].Date
		log.Printf("[Phase 1] No gaps found. Checking for new data after %s...", latestDate)
		return fetchNewDataAfter(latestDate)
	}

	log.Printf("[Phase 1] Found gap at %s. Fetching data from gap...", gapDate)
	return fetchDataFromGap(gapDate)
}

// fetchAllPagesFromStart fetches all pages when the database is empty.
// It starts from the last page (oldest data) and works backwards.
func fetchAllPagesFromStart() int {
	firstResult, err := services.FetchPageData(1, PageSize)
	if err != nil {
		log.Fatalf("Failed to fetch first page: %v", err)
	}

	totalPages := firstResult.Response.PageCount
	totalCount := firstResult.Response.DataCount
	log.Printf("[Phase 1] Total pages: %d, Total records: %d", totalPages, totalCount)

	totalInserted := 0

	for page := totalPages; page >= 1; page-- {
		var result *models.WiliiListResponse
		if page == 1 {
			result = firstResult
		} else {
			log.Printf("[Phase 1] Fetching page %d/%d...", page, totalPages)
			result, err = services.FetchPageData(page, PageSize)
			if err != nil {
				log.Printf("Warning: Failed to fetch page %d: %v", page, err)
				continue
			}
		}

		if len(result.Response.Data) == 0 {
			continue
		}

		inserted := insertListData(result.Response.Data)
		totalInserted += inserted
		log.Printf("[Phase 1] Page %d/%d: %d inserted (total: %d)", page, totalPages, inserted, totalInserted)

		if page != 1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Printf("[Phase 1] Complete. Total inserted: %d", totalInserted)
	return totalInserted
}

// fetchNewDataAfter fetches new data after the given latest date
func fetchNewDataAfter(latestDate string) int {
	totalInserted := 0
	currentPage := 1

	for {
		result, err := services.FetchPageData(currentPage, PageSize)
		if err != nil {
			log.Printf("Warning: Failed to fetch page %d: %v", currentPage, err)
			break
		}

		if len(result.Response.Data) == 0 {
			break
		}

		newRecords := filterByDate(result.Response.Data, latestDate)
		if len(newRecords) == 0 {
			break
		}

		inserted := insertListData(newRecords)
		totalInserted += inserted
		log.Printf("[Phase 1] Page %d: %d new records (total: %d)", currentPage, inserted, totalInserted)

		lastItem := result.Response.Data[len(result.Response.Data)-1]
		if lastItem.Date <= latestDate {
			break
		}

		currentPage++
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("[Phase 1] Complete. New records inserted: %d", totalInserted)
	return totalInserted
}

// fetchDataFromGap fetches data starting from a gap date
func fetchDataFromGap(gapDate string) int {
	totalInserted := 0
	currentPage := 1
	foundGap := false

	for {
		result, err := services.FetchPageData(currentPage, PageSize)
		if err != nil {
			log.Printf("Warning: Failed to fetch page %d: %v", currentPage, err)
			break
		}

		if len(result.Response.Data) == 0 {
			break
		}

		inserted := insertListData(result.Response.Data)
		totalInserted += inserted

		lastItem := result.Response.Data[len(result.Response.Data)-1]
		if lastItem.Date <= gapDate {
			foundGap = true
		}

		if foundGap {
			earliest, _, _ := services.GetDateRange()
			if earliest != "" && lastItem.Date <= earliest {
				break
			}
		}

		log.Printf("[Phase 1] Page %d: %d inserted (total: %d)", currentPage, inserted, totalInserted)

		currentPage++
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("[Phase 1] Complete. Records inserted: %d", totalInserted)
	return totalInserted
}

// insertListData writes list data to the database in batch (without fetching descriptions)
// Data is sorted by date ascending before insertion.
func insertListData(rawData []models.WallpaperRaw) int {
	// Sort by date ascending (earliest first)
	sort.Slice(rawData, func(i, j int) bool {
		return rawData[i].Date < rawData[j].Date
	})

	wallpapers := make([]*models.Wallpaper, 0, len(rawData))
	for _, raw := range rawData {
		wp := services.ConvertToWallpaper(&raw)
		wallpapers = append(wallpapers, wp)
	}

	inserted, err := services.BatchInsertWallpapersIgnore(wallpapers)
	if err != nil {
		log.Printf("Warning: Batch insert failed: %v", err)
	}
	return inserted
}

// ============================================================
// Phase 2: Update descriptions by GUID
// ============================================================

func phase2UpdateDescriptions() int {
	log.Printf("[Phase 2] Fetching wallpapers without description...")

	wallpapers, err := services.GetWallpapersWithoutDescription()
	if err != nil {
		log.Printf("Error fetching wallpapers without description: %v", err)
		return 0
	}

	total := len(wallpapers)
	if total == 0 {
		log.Printf("[Phase 2] All wallpapers have descriptions. Skipping.")
		return 0
	}

	log.Printf("[Phase 2] Found %d wallpapers needing description update", total)

	var updates []services.DescriptionUpdate
	updated := 0
	for i, wp := range wallpapers {
		description, err := services.FetchDetailData(wp.GUID)
		if err != nil {
			log.Printf("Warning: Failed to fetch detail for %s (%s): %v", wp.GUID, wp.Date, err)
			continue
		}

		description = stripHTMLTags(description)
		if description == "" {
			continue
		}

		updates = append(updates, services.DescriptionUpdate{
			GUID:        wp.GUID,
			Description: description,
		})
		updated++

		if (i+1)%50 == 0 || i+1 == total {
			log.Printf("[Phase 2] Progress: %d/%d (%.1f%%)", i+1, total, float64(i+1)/float64(total)*100)
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Batch write all description updates in a single transaction
	if len(updates) > 0 {
		if err := services.BatchUpdateDescriptions(updates); err != nil {
			log.Printf("Warning: Batch update descriptions failed: %v", err)
		}
	}

	log.Printf("[Phase 2] Complete. Updated %d descriptions", updated)
	return updated
}

// ============================================================
// Phase 3: Download thumbnails
// ============================================================

func phase3DownloadThumbnails(workers int, uhd bool) int {
	log.Printf("[Phase 3] Checking thumbnails...")

	pending, stats := collectPendingThumbnails()
	if len(pending) == 0 {
		return 0
	}

	log.Printf("[Phase 3] %d thumbnails need rebuild: missing=%d, wrong_size=%d",
		len(pending), stats.missing, stats.wrongSize)

	// Delete old files before rebuilding
	deleteOldThumbnails(pending)

	// When uhd flag is set, download UHD images and update image_dpi & file_size
	if uhd {
		log.Printf("[Phase 3] UHD mode: downloading high-res images and updating image_dpi & file_size")
		return downloadThumbnailsHD(pending)
	}

	// When count <= PageSize, use UHD quality and record file_size
	useHD := len(pending) <= PageSize
	if useHD {
		log.Printf("[Phase 3] Small batch (%d <= %d), using UHD quality and recording file_size",
			len(pending), PageSize)
		return downloadThumbnailsHD(pending)
	}

	log.Printf("[Phase 3] Using %d workers", workers)
	return downloadThumbnailsParallel(pending, workers)
}

// rebuildStats tracks thumbnail rebuild statistics by reason
type rebuildStats struct {
	missing   int
	wrongSize int
}

// collectPendingThumbnails filters wallpapers that need thumbnail rebuild
func collectPendingThumbnails() ([]*models.Wallpaper, rebuildStats) {
	wallpapers, err := services.GetAllWallpapersOrdered()
	if err != nil {
		log.Printf("Error fetching wallpapers: %v", err)
		return nil, rebuildStats{}
	}

	if len(wallpapers) == 0 {
		log.Printf("[Phase 3] No wallpapers in database. Skipping.")
		return nil, rebuildStats{}
	}

	var pending []*models.Wallpaper
	var stats rebuildStats

	for _, wp := range wallpapers {
		needsRebuild, reason := services.NeedsRebuildThumbnail(wp)
		if !needsRebuild {
			continue
		}
		pending = append(pending, wp)
		if strings.HasPrefix(reason, "wrong_size") {
			stats.wrongSize++
		} else {
			stats.missing++
		}
	}

	if len(pending) == 0 {
		log.Printf("[Phase 3] All %d thumbnails are valid. Skipping.", len(wallpapers))
		return nil, rebuildStats{}
	}

	return pending, stats
}

// deleteOldThumbnails removes existing thumbnail files for wallpapers that need rebuild
func deleteOldThumbnails(pending []*models.Wallpaper) {
	for _, wp := range pending {
		services.DeleteOldThumbnail(wp)
	}
}

// downloadThumbnailsHD downloads UHD images sequentially, generates thumbnails,
// and records the original file size (converted to KB) and actual resolution in the database.
func downloadThumbnailsHD(pending []*models.Wallpaper) int {
	var successCount, failCount int
	var sizeUpdates []services.FileSizeUpdate

	for i, wp := range pending {
		result, err := services.DownloadThumbnailHD(wp)
		if err != nil {
			failCount++
			log.Printf("[Phase 3] Failed HD [%s]: %v", wp.Date, err)
			continue
		}
		successCount++

		// Collect file size updates for batch write
		if result.FileSize > 0 {
			sizeUpdates = append(sizeUpdates, services.FileSizeUpdate{
				GUID:     wp.GUID,
				FileSize: Byte2Kilo(result.FileSize),
			})
			// Update image_dpi to actual resolution
			if result.ImageDPI != "" {
				if err := services.UpdateImageDPI(wp.GUID, result.ImageDPI); err != nil {
					log.Printf("Warning: Failed to update image_dpi for %s: %v", wp.GUID, err)
				}
			}
		}

		if (i+1)%10 == 0 || i+1 == len(pending) {
			log.Printf("[Phase 3] Progress: %d/%d", i+1, len(pending))
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Batch write all file_size updates in a single transaction
	if len(sizeUpdates) > 0 {
		if err := services.BatchUpdateFileSizes(sizeUpdates); err != nil {
			log.Printf("Warning: Batch update file sizes failed: %v", err)
		}
	}

	log.Printf("[Phase 3] Complete. Downloaded: %d, Failed: %d", successCount, failCount)
	return successCount
}

// downloadThumbnailsParallel downloads thumbnails using multiple workers in parallel
func downloadThumbnailsParallel(pending []*models.Wallpaper, workers int) int {
	pendingCount := len(pending)
	jobs := make(chan *models.Wallpaper, pendingCount)

	var successCount, failCount int32

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for wp := range jobs {
				err := services.DownloadThumbnailForWallpaper(wp)
				if err != nil {
					atomic.AddInt32(&failCount, 1)
					log.Printf("[Worker %d] Failed [%s]: %v", workerID, wp.Date, err)
				} else {
					s := atomic.AddInt32(&successCount, 1)
					if s%50 == 0 {
						log.Printf("[Phase 3] Downloaded: %d, Failed: %d, Remaining: %d",
							s, failCount, pendingCount-int(s)-int(failCount))
					}
				}
			}
		}(w)
	}

	for _, wp := range pending {
		jobs <- wp
	}
	close(jobs)

	wg.Wait()

	log.Printf("[Phase 3] Complete. Downloaded: %d, Failed: %d", successCount, failCount)
	return int(successCount)
}

// ============================================================
// Date gap detection
// ============================================================

// extractDates extracts dates from wallpaper list (preserving original order)
func extractDates(wallpapers []*models.Wallpaper) []string {
	dates := make([]string, len(wallpapers))
	for i, wp := range wallpapers {
		dates[i] = wp.Date
	}
	return dates
}

// findEarliestGap finds the earliest gap in a date list.
// Input dates are sorted from latest to earliest; we reverse and check continuity.
func findEarliestGap(dates []string) string {
	if len(dates) < 2 {
		return ""
	}

	sorted := make([]string, len(dates))
	copy(sorted, dates)
	reverseStrings(sorted)

	for i := 0; i < len(sorted)-1; i++ {
		current, err := time.Parse("2006-01-02", sorted[i])
		if err != nil {
			continue
		}
		next, err := time.Parse("2006-01-02", sorted[i+1])
		if err != nil {
			continue
		}

		expectedNext := current.AddDate(0, 0, 1)
		if next.After(expectedNext) {
			return expectedNext.Format("2006-01-02")
		}
	}

	return ""
}

func reverseStrings(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// ============================================================
// Utility functions
// ============================================================

func filterByDate(data []models.WallpaperRaw, lastDate string) []models.WallpaperRaw {
	var filtered []models.WallpaperRaw
	for _, item := range data {
		if item.Date > lastDate {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func stripHTMLTags(html string) string {
	var result strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// DivNatRound performs rounding division of two natural numbers
func DivNatRound(a, b int64) int64 {
	if b <= 0 || a < 0 {
		return -1
	}
	if a == 0 {
		return 0
	}
	return (a + b/2) / b
}

// Byte2Kilo converts bytes to kilobytes (rounded)
func Byte2Kilo(a int64) int64 {
	return DivNatRound(a, 1024)
}

// ============================================================
// Report output
// ============================================================

func printReport(listCount, descCount, thumbCount int) {
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 58))
	fmt.Println("                    UPDATE REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("  Phase 1 - List data inserted:   %d\n", listCount)
	fmt.Printf("  Phase 2 - Descriptions updated:  %d\n", descCount)
	fmt.Printf("  Phase 3 - Thumbnails downloaded: %d\n", thumbCount)
	fmt.Println()
	fmt.Println(strings.Repeat("-", 60))
}

// setupFileLogger adds a file writer to the standard logger for error-level output
func setupFileLogger(conf *services.Config) {
	logDir := filepath.Dir(conf.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: failed to create log directory %s: %v", logDir, err)
		return
	}

	f, err := os.OpenFile(conf.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: failed to open log file %s: %v", conf.LogFile, err)
		return
	}

	// Write to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
}
