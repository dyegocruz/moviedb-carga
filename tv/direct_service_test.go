package tv

import (
	"reflect"
	"testing"
	"time"

	"moviedb/common"
)

func TestDirectSerieMetadataAndUpsert(t *testing.T) {
	now := time.Date(2026, time.October, 21, 12, 30, 0, 0, time.UTC)
	got := applySerieMetadata(Serie{Id: 10, Title: "Serie"}, common.LANGUAGE_PTBR, now)
	if got.Id != 10 || got.Title != "Serie" || got.MediaType != "serie" ||
		got.Language != common.LANGUAGE_PTBR || got.UpdatedAt != "21/10/2026 12:30:00" {
		t.Fatalf("unexpected metadata: %+v", got)
	}
	for _, tc := range []struct {
		existingID int
		itemID     int
		want       string
	}{{0, 10, "insert"}, {10, 10, "update"}, {0, 0, "skip"}} {
		if got := decideSerieUpsertAction(tc.existingID, tc.itemID); got != tc.want {
			t.Fatalf("upsert(%d, %d) = %q, want %q", tc.existingID, tc.itemID, got, tc.want)
		}
	}
}

func TestMergeSerieLocalizationAndTitles(t *testing.T) {
	main := Serie{Id: 10, Title: "Serie PT", Overview: "Resumo", PosterPath: "/pt.jpg",
		Localizations:       []common.LocalizationMovieTv{{Locale: "old"}},
		AlternativeTitlesDb: []common.AlternativeTitle{{Title: "Old"}}}
	en := Serie{Title: "Serie EN", Overview: "Summary", PosterPath: "/en.jpg"}
	titles := []common.AlternativeTitle{{Iso3166_1: "BR", Title: "Alternativo", Type: "Other"}}

	got := mergeSerieLocalizationAndTitles(main, en, titles)
	want := []common.LocalizationMovieTv{
		{Locale: common.LANGUAGE_PTBR, Title: "Serie PT", Synopsis: "Resumo", PosterPath: "/pt.jpg"},
		{Locale: common.LANGUAGE_EN, Title: "Serie EN", Synopsis: "Summary", PosterPath: "/en.jpg"},
	}
	if got.Id != main.Id || !reflect.DeepEqual(got.Localizations, want) ||
		!reflect.DeepEqual(got.AlternativeTitlesDb, titles) {
		t.Fatalf("unexpected merged serie: %+v", got)
	}
}
