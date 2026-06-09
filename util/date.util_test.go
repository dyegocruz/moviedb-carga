package util

import (
	"testing"
	"time"
)

func TestGetDateNowByFormatUrl(t *testing.T) {
	got := GetDateNowByFormatUrl()

	if _, err := time.Parse("2006-01-02", got); err != nil {
		t.Fatalf("expected date in yyyy-mm-dd format, got %q, err=%v", got, err)
	}
}
