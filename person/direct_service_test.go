package person

import (
	"testing"
	"time"

	"moviedb/common"
)

func TestDirectPersonMetadataAndUpsert(t *testing.T) {
	now := time.Date(2026, time.October, 21, 12, 30, 0, 0, time.UTC)
	got := applyPersonMetadata(Person{Id: 10, Name: "Person"}, common.LANGUAGE_PTBR, now)
	if got.Id != 10 || got.Name != "Person" || got.Language != common.LANGUAGE_PTBR ||
		got.UpdatedAt != "21/10/2026 12:30:00" {
		t.Fatalf("unexpected metadata: %+v", got)
	}
	for _, tc := range []struct {
		existingID int
		itemID     int
		update     string
		want       string
	}{
		{0, 10, "N", "insert"},
		{0, 10, "Y", "insert"},
		{10, 10, "Y", "update"},
		{10, 10, "N", "skip"},
		{0, 0, "Y", "skip"},
	} {
		if got := decidePersonUpsertAction(tc.existingID, tc.itemID, tc.update); got != tc.want {
			t.Fatalf("upsert(%d, %d, %q) = %q, want %q",
				tc.existingID, tc.itemID, tc.update, got, tc.want)
		}
	}
}
