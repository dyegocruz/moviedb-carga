package util

import "time"

func GetDateNowByFormatUrl() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02")
}
