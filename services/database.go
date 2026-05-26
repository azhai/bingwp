package services

import (
	"context"
	"fmt"

	"github.com/azhai/bingwp/models"
	"github.com/azhai/goent"
	"github.com/azhai/goent/model"
)

// getDB returns the current database connection
func getDB() *models.Database {
	return models.GetDB()
}

// InsertWallpaperIgnore inserts a wallpaper, ignoring if date already exists
func InsertWallpaperIgnore(wp *models.Wallpaper) error {
	db := getDB()
	ctx := context.Background()
	sql := `INSERT OR IGNORE INTO wallpapers (guid, date, slug, title, headline, copyright, description, image_dpi, file_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return db.RawExecContext(ctx, sql,
		wp.GUID, wp.Date, wp.Slug, wp.Title, wp.Headline,
		wp.Copyright, wp.Description, wp.ImageDPI, wp.FileSize,
	)
}

// BatchInsertWallpapersIgnore inserts multiple wallpapers in a single transaction,
// ignoring records where date already exists. Returns the number of inserted rows.
func BatchInsertWallpapersIgnore(wallpapers []*models.Wallpaper) (int, error) {
	if len(wallpapers) == 0 {
		return 0, nil
	}

	db := getDB()
	inserted := 0
	err := db.BeginTransaction(func(tx model.Transaction) error {
		for _, wp := range wallpapers {
			sql := `INSERT OR IGNORE INTO wallpapers (guid, date, slug, title, headline, copyright, description, image_dpi, file_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
			qr := model.CreateQuery(sql, []any{wp.GUID, wp.Date, wp.Slug, wp.Title, wp.Headline,
				wp.Copyright, wp.Description, wp.ImageDPI, wp.FileSize})
			if err := tx.ExecContext(context.Background(), &qr); err != nil {
				continue
			}
			inserted++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("batch insert failed: %w", err)
	}
	return inserted, nil
}

// BatchUpdateDescriptions updates description for multiple wallpapers by GUID in a single transaction
func BatchUpdateDescriptions(updates []DescriptionUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	db := getDB()
	return db.BeginTransaction(func(tx model.Transaction) error {
		for _, u := range updates {
			sql := `UPDATE wallpapers SET description = ? WHERE guid = ?`
			qr := model.CreateQuery(sql, []any{u.Description, u.GUID})
			if err := tx.ExecContext(context.Background(), &qr); err != nil {
				return fmt.Errorf("update description for %s failed: %w", u.GUID, err)
			}
		}
		return nil
	})
}

// BatchUpdateFileSizes updates file_size for multiple wallpapers by GUID in a single transaction
func BatchUpdateFileSizes(updates []FileSizeUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	db := getDB()
	return db.BeginTransaction(func(tx model.Transaction) error {
		for _, u := range updates {
			sql := `UPDATE wallpapers SET file_size = ? WHERE guid = ?`
			qr := model.CreateQuery(sql, []any{u.FileSize, u.GUID})
			if err := tx.ExecContext(context.Background(), &qr); err != nil {
				return fmt.Errorf("update file_size for %s failed: %w", u.GUID, err)
			}
		}
		return nil
	})
}

// DescriptionUpdate holds data for a batch description update
type DescriptionUpdate struct {
	GUID        string
	Description string
}

// FileSizeUpdate holds data for a batch file_size update
type FileSizeUpdate struct {
	GUID     string
	FileSize int64
}

// GetWallpapersByMonth returns all wallpapers for a given year and month
func GetWallpapersByMonth(year, month int) ([]*models.Wallpaper, error) {
	db := getDB()
	startDate := fmt.Sprintf("%04d-%02d-01", year, month)
	endDate := fmt.Sprintf("%04d-%02d-01", year, month+1)
	if month == 12 {
		endDate = fmt.Sprintf("%04d-01-01", year+1)
	}

	filter := goent.And(
		goent.GreaterEquals(db.Wallpaper.Field("date"), startDate),
		goent.Less(db.Wallpaper.Field("date"), endDate),
	)

	return db.Wallpaper.Filter(filter).Select().OrderBy("date DESC").All()
}

// UpdateDescription updates the description field of a wallpaper by GUID
func UpdateDescription(guid, description string) error {
	db := getDB()
	ctx := context.Background()
	sql := `UPDATE wallpapers SET description = ? WHERE guid = ?`
	return db.RawExecContext(ctx, sql, description, guid)
}

// UpdateFileSize updates the file_size field of a wallpaper by GUID
func UpdateFileSize(guid string, fileSize int64) error {
	db := getDB()
	ctx := context.Background()
	sql := `UPDATE wallpapers SET file_size = ? WHERE guid = ?`
	return db.RawExecContext(ctx, sql, fileSize, guid)
}

// UpdateImageDPI updates the image_dpi field of a wallpaper by GUID
func UpdateImageDPI(guid, imageDPI string) error {
	db := getDB()
	ctx := context.Background()
	sql := `UPDATE wallpapers SET image_dpi = ? WHERE guid = ?`
	return db.RawExecContext(ctx, sql, imageDPI, guid)
}

// GetWallpapersWithoutDescription returns wallpapers that have no description
// Only checks wallpapers from 2014-05-01 onwards, as earlier ones have no description
func GetWallpapersWithoutDescription() ([]*models.Wallpaper, error) {
	db := getDB()
	filter := goent.And(
		goent.GreaterEquals(db.Wallpaper.Field("date"), "2014-05-01"),
		goent.Or(
			goent.IsNull(db.Wallpaper.Field("description")),
			goent.Equals(db.Wallpaper.Field("description"), ""),
		),
	)
	return db.Wallpaper.Filter(filter).Select().OrderBy("date ASC").All()
}

// GetAllWallpapersOrdered returns all wallpapers ordered by date ascending
func GetAllWallpapersOrdered() ([]*models.Wallpaper, error) {
	db := getDB()
	return db.Wallpaper.Select().OrderBy("date ASC").All()
}

// GetRecentWallpapers returns the most recent N wallpapers ordered by date descending
func GetRecentWallpapers(limit int) ([]*models.Wallpaper, error) {
	db := getDB()
	return db.Wallpaper.Select().OrderBy("date DESC").Take(limit).All()
}

// GetDateRange returns the earliest and latest dates in the database
func GetDateRange() (earliest, latest string, err error) {
	db := getDB()
	// Latest
	qLatest := db.Wallpaper.Select("date").OrderBy("date DESC").Take(1)
	obj, e := qLatest.One()
	if e != nil || obj == nil {
		latest = ""
	} else {
		latest = obj.Date
	}
	// Earliest
	qEarliest := db.Wallpaper.Select("date").OrderBy("date ASC").Take(1)
	obj, e = qEarliest.One()
	if e != nil || obj == nil {
		earliest = ""
	} else {
		earliest = obj.Date
	}
	return earliest, latest, nil
}
