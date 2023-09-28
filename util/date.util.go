package util

import "time"

func GetDateNowByFormatUrl() string {
	currentTime := time.Now()
	// return currentTime.Format("2006-01-02")
	return currentTime.AddDate(0, 0, -1).Format("2006-01-02")
}
