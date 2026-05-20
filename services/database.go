package services

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/azhai/bingwp/models"
)

var db *sql.DB

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

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func GetDB() *sql.DB {
	return db
}
