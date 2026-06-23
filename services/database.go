package services

import (
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
	err := db.Wallpaper.Insert().One(wp)
	if err != nil && isUniqueConstraintError(err) {
		return nil
	}
	return err
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
			err := db.Wallpaper.Insert().OnTransaction(tx).One(wp)
			if err != nil && isUniqueConstraintError(err) {
				continue
			}
			if err != nil {
				return err
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

// FieldUpdate holds a GUID and the column-value pairs to update
type FieldUpdate struct {
	GUID  string
	Pairs []goent.Pair
}

// BatchUpdateWallpaperFields updates multiple fields for multiple wallpapers by GUID in a single transaction
func BatchUpdateWallpaperFields(updates []FieldUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	db := getDB()
	return db.BeginTransaction(func(tx model.Transaction) error {
		for _, u := range updates {
			if len(u.Pairs) == 0 {
				continue
			}
			err := db.Wallpaper.Update().OnTransaction(tx).
				Set(u.Pairs...).
				Where("guid = ?", u.GUID).Exec()
			if err != nil {
				return fmt.Errorf("update fields for %s failed: %w", u.GUID, err)
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
			err := db.Wallpaper.Update().OnTransaction(tx).
				Set(goent.Pair{Key: "file_size", Value: u.FileSize}).
				Where("guid = ?", u.GUID).Exec()
			if err != nil {
				return fmt.Errorf("update file_size for %s failed: %w", u.GUID, err)
			}
		}
		return nil
	})
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
	return db.Wallpaper.Update().
		Set(goent.Pair{Key: "description", Value: description}).
		Where("guid = ?", guid).Exec()
}

// UpdateFileSize updates the file_size field of a wallpaper by GUID
func UpdateFileSize(guid string, fileSize int64) error {
	db := getDB()
	return db.Wallpaper.Update().
		Set(goent.Pair{Key: "file_size", Value: fileSize}).
		Where("guid = ?", guid).Exec()
}

// UpdateImageDPI updates the image_dpi field of a wallpaper by GUID
func UpdateImageDPI(guid, imageDPI string) error {
	db := getDB()
	return db.Wallpaper.Update().
		Set(goent.Pair{Key: "image_dpi", Value: imageDPI}).
		Where("guid = ?", guid).Exec()
}

// GetRecentWallpapers returns the most recent N wallpapers ordered by date descending
func GetRecentWallpapers(limit int) ([]*models.Wallpaper, error) {
	db := getDB()
	return db.Wallpaper.Select().OrderBy("date DESC").Take(limit).All()
}

// GetDateRange returns the earliest and latest dates in the database
func GetDateRange() (earliest, latest string, err error) {
	db := getDB()
	qLatest := db.Wallpaper.Select("date").OrderBy("date DESC").Take(1)
	obj, e := qLatest.One()
	if e != nil || obj == nil {
		latest = ""
	} else {
		latest = obj.Date
	}
	qEarliest := db.Wallpaper.Select("date").OrderBy("date ASC").Take(1)
	obj, e = qEarliest.One()
	if e != nil || obj == nil {
		earliest = ""
	} else {
		earliest = obj.Date
	}
	return earliest, latest, nil
}

// isUniqueConstraintError checks if the error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "UNIQUE constraint failed") ||
		contains(msg, "duplicate key") ||
		contains(msg, "Duplicate entry")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
