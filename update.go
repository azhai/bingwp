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
	"github.com/azhai/goent"
)

const PageSize = 50

type UpdateCmd struct {
	Workers int  `arg:"-w,--workers,env:WORKERS" help:"Number of download workers" default:"8"`
	UHD     bool `arg:"--uhd" help:"Download UHD images and update image_dpi & file_size"`
}

func (c *UpdateCmd) Run() {
	cfg := services.LoadConfig()
	if err := setup(cfg); err != nil {
		log.Fatalf("Failed to run update: %v", err)
	}
	defer models.CloseDB()
	setupFileLogger(cfg)

	// Phase 1: Fetch list data
	phase1Count := phase1FetchListData()

	// Phase 1.5: Backfill missing fields from list API
	phase15Count := phase15BackfillMissingFields()

	// Phase 2: Update descriptions by GUID
	phase2Count := phase2UpdateDescriptions()

	// Phase 3: Download thumbnails
	phase3Count := phase3DownloadThumbnails(c.Workers, c.UHD)

	printReport(phase1Count, phase15Count, phase2Count, phase3Count)
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

	gapDate := findEarliestGap(recent)

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
	// Fetch page 1 first to get total page count
	firstResult, err := services.FetchPageData(1, PageSize)
	if err != nil {
		log.Fatalf("Failed to fetch first page: %v", err)
	}

	totalPages := firstResult.Response.PageCount
	log.Printf("[Phase 1] Total pages: %d, Total records: %d",
		totalPages, firstResult.Response.DataCount)

	totalInserted := 0

	// Process page 1 data first (already fetched)
	if len(firstResult.Response.Data) > 0 {
		inserted := insertListData(firstResult.Response.Data)
		totalInserted += inserted
		log.Printf("[Phase 1] Page 1/%d: %d inserted (total: %d)",
			totalPages, inserted, totalInserted)
	}

	// Process remaining pages from last (oldest) to page 2
	for page := totalPages; page >= 2; page-- {
		log.Printf("[Phase 1] Fetching page %d/%d...", page, totalPages)
		result, err := services.FetchPageData(page, PageSize)
		if err != nil {
			log.Printf("Warning: Failed to fetch page %d: %v", page, err)
			continue
		}

		if len(result.Response.Data) == 0 {
			continue
		}

		inserted := insertListData(result.Response.Data)
		totalInserted += inserted
		log.Printf("[Phase 1] Page %d/%d: %d inserted (total: %d)",
			page, totalPages, inserted, totalInserted)

		time.Sleep(100 * time.Millisecond)
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

		data := result.Response.Data
		if len(data) == 0 {
			break
		}

		// Filter in-place: keep only records newer than latestDate
		newRecords := filterByDateInPlace(data, latestDate)
		if len(newRecords) == 0 {
			break
		}

		inserted := insertListData(newRecords)
		totalInserted += inserted
		log.Printf("[Phase 1] Page %d: %d new records (total: %d)",
			currentPage, inserted, totalInserted)

		// Stop if the last record is older than or equal to latestDate
		if data[len(data)-1].Date <= latestDate {
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
// Phase 1.5: Backfill missing fields from list API
// ============================================================

func phase15BackfillMissingFields() int {
	log.Printf("[Phase 1.5] Checking recent wallpapers for missing fields...")

	recent, err := services.GetRecentWallpapers(PageSize)
	if err != nil {
		log.Printf("Warning: Failed to query recent wallpapers: %v", err)
		return 0
	}

	var needing []*models.Wallpaper
	for _, wp := range recent {
		if wp.Title == "" || wp.Headline == "" || wp.Copyright == "" {
			needing = append(needing, wp)
		}
	}

	if len(needing) == 0 {
		log.Printf("[Phase 1.5] No missing fields in recent %d wallpapers. Skipping.", len(recent))
		return 0
	}

	log.Printf("[Phase 1.5] Found %d wallpapers with missing fields", len(needing))

	result, err := services.FetchPageData(1, PageSize)
	if err != nil {
		log.Printf("Warning: Phase 1.5 list API failed: %v", err)
		return 0
	}

	apiData := make(map[string]*models.WallpaperRaw, len(result.Response.Data))
	for i := range result.Response.Data {
		apiData[result.Response.Data[i].GUID] = &result.Response.Data[i]
	}

	updates := make([]services.FieldUpdate, 0, len(needing))
	for _, wp := range needing {
		raw, ok := apiData[wp.GUID]
		if !ok {
			continue
		}
		pairs := buildBackfillPairs(wp, raw)
		if len(pairs) > 0 {
			updates = append(updates, services.FieldUpdate{
				GUID:  wp.GUID,
				Pairs: pairs,
			})
		}
	}

	if len(updates) > 0 {
		if err := services.BatchUpdateWallpaperFields(updates); err != nil {
			log.Printf("Warning: Batch update fields failed: %v", err)
		}
	}

	log.Printf("[Phase 1.5] Complete. Backfilled %d wallpapers", len(updates))
	return len(updates)
}

func buildBackfillPairs(wp *models.Wallpaper, raw *models.WallpaperRaw) []goent.Pair {
	var pairs []goent.Pair
	if wp.Title == "" && raw.Title != "" {
		pairs = append(pairs, goent.Pair{Key: "title", Value: raw.Title})
	}
	if wp.Headline == "" && raw.Headline != "" {
		pairs = append(pairs, goent.Pair{Key: "headline", Value: raw.Headline})
	}
	if wp.Copyright == "" && raw.Copyright != "" {
		pairs = append(pairs, goent.Pair{Key: "copyright", Value: raw.Copyright})
	}
	return pairs
}

// ============================================================
// Phase 2: Update descriptions by GUID
// ============================================================

func phase2UpdateDescriptions() int {
	log.Printf("[Phase 2] Checking recent %d wallpapers for missing fields...", PageSize)

	recent, err := services.GetRecentWallpapers(PageSize)
	if err != nil {
		log.Printf("Error fetching recent wallpapers: %v", err)
		return 0
	}

	// Filter wallpapers with any missing field (description, title, headline, copyright)
	var needing []*models.Wallpaper
	for _, wp := range recent {
		descEmpty := wp.Description == nil || *wp.Description == ""
		if descEmpty || wp.Title == "" || wp.Headline == "" || wp.Copyright == "" {
			needing = append(needing, wp)
		}
	}

	total := len(needing)
	if total == 0 {
		log.Printf("[Phase 2] All recent wallpapers have complete data. Skipping.")
		return 0
	}

	log.Printf("[Phase 2] Found %d wallpapers needing detail update", total)

	updates := make([]services.FieldUpdate, 0, total)
	updated := 0
	for i, wp := range needing {
		detail, err := services.FetchDetailData(wp.GUID)
		if err != nil {
			log.Printf("Warning: Failed to fetch detail for %s (%s): %v", wp.GUID, wp.Date, err)
			continue
		}

		pairs := buildDetailBackfillPairs(wp, detail)
		if len(pairs) == 0 {
			continue
		}

		updates = append(updates, services.FieldUpdate{
			GUID:  wp.GUID,
			Pairs: pairs,
		})
		updated++

		if (i+1)%10 == 0 || i+1 == total {
			log.Printf("[Phase 2] Progress: %d/%d (%.1f%%)", i+1, total, float64(i+1)/float64(total)*100)
		}

		time.Sleep(50 * time.Millisecond)
	}

	if len(updates) > 0 {
		if err := services.BatchUpdateWallpaperFields(updates); err != nil {
			log.Printf("Warning: Batch update fields failed: %v", err)
		}
	}

	log.Printf("[Phase 2] Complete. Updated %d wallpapers", updated)
	return updated
}

// buildDetailBackfillPairs compares existing wallpaper with detail API data and returns
// pairs for fields that are currently empty but available in the detail response.
func buildDetailBackfillPairs(wp *models.Wallpaper, detail *models.DetailData) []goent.Pair {
	var pairs []goent.Pair

	if wp.Title == "" && detail.Title != "" {
		pairs = append(pairs, goent.Pair{Key: "title", Value: detail.Title})
	}
	if wp.Headline == "" && detail.Headline != "" {
		pairs = append(pairs, goent.Pair{Key: "headline", Value: detail.Headline})
	}
	if wp.Copyright == "" && detail.Copyright != "" {
		pairs = append(pairs, goent.Pair{Key: "copyright", Value: detail.Copyright})
	}

	descEmpty := wp.Description == nil || *wp.Description == ""
	if descEmpty && detail.Description != "" {
		description := stripHTMLTags(detail.Description)
		if description != "" {
			pairs = append(pairs, goent.Pair{Key: "description", Value: description})
		}
	}

	return pairs
}

// ============================================================
// Phase 3: Download thumbnails
// ============================================================

func backfillMetadata(guid string, result services.DownloadResult) {
	if result.FileSize <= 0 {
		return
	}

	fileSizeKB := Byte2Kilo(result.FileSize)
	if err := services.UpdateFileSize(guid, fileSizeKB); err != nil {
		log.Printf("Warning: Failed to update file_size for %s: %v", guid, err)
	}

	if result.ImageDPI != "" {
		if err := services.UpdateImageDPI(guid, result.ImageDPI); err != nil {
			log.Printf("Warning: Failed to update image_dpi for %s: %v", guid, err)
		}
	}

	log.Printf("[Phase 3] Backfilled %s: ImageDPI=%s, FileSize=%dKB", guid, result.ImageDPI, fileSizeKB)
}

func phase3DownloadThumbnails(workers int, uhd bool) int {
	log.Printf("[Phase 3] Checking thumbnails...")

	pending, stats := collectPendingThumbnails()
	if len(pending) == 0 {
		return 0
	}

	log.Printf("[Phase 3] %d thumbnails need rebuild: missing=%d, wrong_size=%d",
		len(pending), stats.missing, stats.wrongSize)

	deleteOldThumbnails(pending, stats)

	if uhd {
		log.Printf("[Phase 3] UHD mode: downloading high-res images and updating metadata")
	} else {
		log.Printf("[Phase 3] FHD mode: downloading 1920x1080 images and updating metadata")
	}

	return processWallpapers(pending, uhd, workers)
}

// rebuildStats tracks thumbnail rebuild statistics by reason
type rebuildStats struct {
	missing   int
	wrongSize int
}

// collectPendingThumbnails filters recent wallpapers that need thumbnail rebuild.
// Only checks the most recent PageSize wallpapers to keep the update fast.
func collectPendingThumbnails() ([]*models.Wallpaper, rebuildStats) {
	wallpapers, err := services.GetRecentWallpapers(PageSize)
	if err != nil {
		log.Printf("Error fetching recent wallpapers: %v", err)
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
		needsBackfill := wp.ImageDPI == "" || wp.FileSize <= 0

		if !needsRebuild && !needsBackfill {
			continue
		}
		pending = append(pending, wp)
		if needsRebuild {
			if strings.HasPrefix(reason, "wrong_size") {
				stats.wrongSize++
			} else {
				stats.missing++
			}
		}
	}

	if len(pending) == 0 {
		log.Printf("[Phase 3] All %d thumbnails are valid and metadata is complete. Skipping.", len(wallpapers))
		return nil, rebuildStats{}
	}

	return pending, stats
}

func deleteOldThumbnails(pending []*models.Wallpaper, stats rebuildStats) {
	if stats.missing > 0 || stats.wrongSize > 0 {
		for _, wp := range pending {
			needsRebuild, _ := services.NeedsRebuildThumbnail(wp)
			if needsRebuild {
				services.DeleteOldThumbnail(wp)
			}
		}
	}
}

func processWallpapers(pending []*models.Wallpaper, uhd bool, workers int) int {
	if uhd || workers <= 1 {
		return processWallpapersSequential(pending, uhd)
	}

	log.Printf("[Phase 3] Using %d workers", workers)
	successCount, failCount := processWallpapersParallel(pending, workers)
	log.Printf("[Phase 3] Complete. Processed: %d, Failed: %d", successCount, failCount)
	return successCount
}

func processWallpapersSequential(pending []*models.Wallpaper, uhd bool) int {
	var successCount, failCount int

	for i, wp := range pending {
		var result services.DownloadResult
		var err error

		if uhd {
			result, err = services.DownloadThumbnailHD(wp)
		} else {
			result, err = services.DownloadThumbnailForWallpaper(wp)
		}

		if err != nil {
			failCount++
			log.Printf("[Phase 3] Failed [%s]: %v", wp.Date, err)
			continue
		}
		successCount++

		backfillMetadata(wp.GUID, result)

		if (i+1)%10 == 0 || i+1 == len(pending) {
			log.Printf("[Phase 3] Progress: %d/%d", i+1, len(pending))
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("[Phase 3] Complete. Processed: %d, Failed: %d", successCount, failCount)
	return successCount
}

func processWallpapersParallel(pending []*models.Wallpaper, workers int) (int, int) {
	pendingCount := len(pending)
	// Small buffered channel: workers consume as fast as they can, producer
	// blocks until a slot is free. This avoids buffering all jobs at once.
	jobs := make(chan *models.Wallpaper, workers)

	var successCount, failCount int32

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for wp := range jobs {
				result, err := services.DownloadThumbnailForWallpaper(wp)
				if err != nil {
					atomic.AddInt32(&failCount, 1)
					log.Printf("[Worker %d] Failed [%s]: %v", workerID, wp.Date, err)
				} else {
					s := atomic.AddInt32(&successCount, 1)
					backfillMetadata(wp.GUID, result)
					if s%50 == 0 {
						log.Printf("[Phase 3] Processed: %d, Failed: %d, Remaining: %d",
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
	return int(successCount), int(failCount)
}

// ============================================================
// Date gap detection
// ============================================================

// findEarliestGap finds the earliest gap in a wallpaper list sorted by date descending.
// Iterates from earliest (end) to latest (start), checking day-by-day continuity.
// Returns the missing date (the day after the earlier record) if a gap is found.
func findEarliestGap(wallpapers []*models.Wallpaper) string {
	n := len(wallpapers)
	if n < 2 {
		return ""
	}

	// Walk from the end (earliest) towards the start (latest)
	for i := n - 1; i > 0; i-- {
		current, err := time.Parse("2006-01-02", wallpapers[i].Date)
		if err != nil {
			continue
		}
		prev, err := time.Parse("2006-01-02", wallpapers[i-1].Date)
		if err != nil {
			continue
		}

		// prev is one day later in descending order; expect exactly 1 day difference
		expectedPrev := current.AddDate(0, 0, 1)
		if prev.After(expectedPrev) {
			return expectedPrev.Format("2006-01-02")
		}
	}

	return ""
}

// ============================================================
// Utility functions
// ============================================================

// filterByDateInPlace filters records in-place, keeping only those with Date > lastDate.
// Returns a subslice of the input; no allocation.
func filterByDateInPlace(data []models.WallpaperRaw, lastDate string) []models.WallpaperRaw {
	kept := data[:0]
	for _, item := range data {
		if item.Date > lastDate {
			kept = append(kept, item)
		}
	}
	return kept
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

func printReport(listCount, backfillCount, descCount, thumbCount int) {
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 58))
	fmt.Println("                    UPDATE REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("  Phase 1   - List data inserted:      %d\n", listCount)
	fmt.Printf("  Phase 1.5 - Missing fields backfilled: %d\n", backfillCount)
	fmt.Printf("  Phase 2   - Descriptions updated:     %d\n", descCount)
	fmt.Printf("  Phase 3   - Thumbnails downloaded:    %d\n", thumbCount)
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
