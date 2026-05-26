package handlers

import (
	"fmt"
	"time"
)

func getPrevMonth(year, month int) string {
	if year == 2009 && month == 7 {
		return ""
	}
	if month == 1 {
		return fmt.Sprintf("/%04d%02d", year-1, 12)
	}
	return fmt.Sprintf("/%04d%02d", year, month-1)
}

func getNextMonth(year, month int, now time.Time) string {
	next := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, now.Location())
	if next.After(now) {
		return ""
	}
	return fmt.Sprintf("/%04d%02d", next.Year(), next.Month())
}
