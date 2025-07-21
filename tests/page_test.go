package tests

import (
	"testing"
	"time"

	"github.com/azhai/bingwp/handlers"
	"github.com/azhai/bingwp/services/db"
	"github.com/stretchr/testify/assert"
)

// TestReadFirstDaily 测试壁纸列表读取
func TestReadFirstDaily(t *testing.T) {
	dt, _ := time.Parse("200601", "202505")
	monthBegin := handlers.GetMonthBegin(dt)
	loc := monthBegin.Location()
	assert.Equal(t, monthBegin, time.Date(2025, 5, 1, 0, 0, 0, 0, loc))
	nextBegin := handlers.GetMonthBegin(monthBegin.AddDate(0, 0, 31))
	assert.Equal(t, nextBegin, time.Date(2025, 6, 1, 0, 0, 0, 0, loc))

	var row *db.WallDaily
	rows := db.GetMonthDailyRows(monthBegin, nextBegin)
	if len(rows) > 0 {
		rows = db.GetDailyNotes(db.GetDailyImages(rows))
		row = rows[0]
	} else {
		row = new(db.WallDaily)
	}
	assert.NotEmpty(t, rows)
	assert.NotEmpty(t, row.Thumb)
	assert.NotEmpty(t, row.Image)
	assert.NotEmpty(t, row.Notes)

	assert.Equal(t, "PinkPlumeria_ZH-CN3890147555", row.BingSku)
	assert.Contains(t, row.Notes, "title")
	assert.Equal(t, row.Title, row.Notes["title"].NoteChinese.String)
	assert.Contains(t, row.Notes, "headline")
	assert.Equal(t, row.Headline, row.Notes["headline"].NoteChinese.String)
}

func TestQueryDaily(t *testing.T) {
	rows := db.GetLatestDailyRows(3, 0)
	rows = db.GetDailyNotes(rows)
	assert.NotEmpty(t, rows)
}
