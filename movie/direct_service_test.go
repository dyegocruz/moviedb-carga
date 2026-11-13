package movie

import (
	"reflect"
	"testing"
	"time"

	"moviedb/common"
)

func TestDirectMovieMetadataAndUpsert(t *testing.T) {
	now := time.Date(2026, time.October, 21, 12, 30, 0, 0, time.UTC)
	got := applyMovieMetadata(Movie{Id: 10, Title: "Movie"}, common.LANGUAGE_PTBR, now)
	if got.Id != 10 || got.Title != "Movie" || got.MediaType != "movie" ||
		got.Language != common.LANGUAGE_PTBR || got.UpdatedAt != "21/10/2026 12:30:00" {
		t.Fatalf("unexpected metadata: %+v", got)
	}
	for _, tc := range []struct {
		existingID int
		itemID     int
		want       string
	}{{0, 10, "insert"}, {10, 10, "update"}, {0, 0, "skip"}} {
		if got := decideMovieUpsertAction(tc.existingID, tc.itemID); got != tc.want {
			t.Fatalf("upsert(%d, %d) = %q, want %q", tc.existingID, tc.itemID, got, tc.want)
		}
	}
}

func TestMergeMovieLocalizationAndTitles(t *testing.T) {
	main := Movie{Id: 10, Title: "Filme", Overview: "Resumo", PosterPath: "/pt.jpg",
		Localizations:       []common.LocalizationMovieTv{{Locale: "old"}},
		AlternativeTitlesDb: []common.AlternativeTitle{{Title: "Old"}}}
	en := Movie{Title: "Movie", Overview: "Summary", PosterPath: "/en.jpg"}
	titles := []common.AlternativeTitle{{Iso3166_1: "BR", Title: "Alternativo", Type: "Other"}}

	got := mergeMovieLocalizationAndTitles(main, en, titles)
	want := []common.LocalizationMovieTv{
		{Locale: common.LANGUAGE_PTBR, Title: "Filme", Synopsis: "Resumo", PosterPath: "/pt.jpg"},
		{Locale: common.LANGUAGE_EN, Title: "Movie", Synopsis: "Summary", PosterPath: "/en.jpg"},
	}
	if got.Id != main.Id || !reflect.DeepEqual(got.Localizations, want) ||
		!reflect.DeepEqual(got.AlternativeTitlesDb, titles) {
		t.Fatalf("unexpected merged movie: %+v", got)
	}
}
