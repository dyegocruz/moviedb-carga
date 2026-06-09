package util

import "time"

func GetDateNowByFormatUrl() string {
	currentTime := time.Now()
	return currentTime.AddDate(0, 0, -1).Format("2006-01-02")
}

func ParseStringToTime(dateStr string) (time.Time, error) {
	const layout = "02/01/2006 15:04:05"
	return time.ParseInLocation(layout, dateStr, time.Local)
}
