package catalogCharge

import (
	"testing"
	"time"

	"moviedb/common"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/tv"
)

func TestCatalogCharge_CallsTvAndMovieHandlers(t *testing.T) {
	svc := &Service{}

	called := make(chan string, 2)
	svc.checkAndUpdateCatalogByFileFn = func(mediaType string) {
		called <- mediaType
	}

	svc.CatalogCharge()

	seen := map[string]bool{}
	for i := 0; i < 2; i++ {
		select {
		case mediaType := <-called:
			seen[mediaType] = true
		case <-time.After(300 * time.Millisecond):
			t.Fatal("timed out waiting for CatalogCharge calls")
		}
	}

	if !seen[common.MEDIA_TYPE_TV] || !seen[common.MEDIA_TYPE_MOVIE] {
		t.Fatalf("expected tv and movie calls, got %+v", seen)
	}
}

func TestCatalogUpdates_CallsMovieAndTvHandlers(t *testing.T) {
	svc := &Service{}

	movieCalled := make(chan struct{}, 1)
	tvCalled := make(chan struct{}, 1)

	svc.checkMoviesChangesFn = func() { movieCalled <- struct{}{} }
	svc.checkTvChangesFn = func() { tvCalled <- struct{}{} }

	svc.CatalogUpdates()

	select {
	case <-movieCalled:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for movie changes call")
	}

	select {
	case <-tvCalled:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for tv changes call")
	}
}

func TestElasticGeneralCharge_DelegatesToCatalogSearchCharge(t *testing.T) {
	svc := &Service{}
	called := false
	svc.catalogSearchChargeFn = func() { called = true }

	svc.ElasticGeneralCharge()

	if !called {
		t.Fatal("expected CatalogSearchCharge to be called")
	}
}

func TestGeneralCatalogHandler_DelegatesToCatalogAndUpdates(t *testing.T) {
	svc := &Service{}
	catalogCalled := false
	updatesCalled := false

	svc.catalogChargeFn = func() { catalogCalled = true }
	svc.catalogUpdatesFn = func() { updatesCalled = true }

	svc.GeneralCatalogHandler()

	if !catalogCalled || !updatesCalled {
		t.Fatalf("expected both calls, got catalog=%v updates=%v", catalogCalled, updatesCalled)
	}
}

func TestShouldPublishCatalogMessage(t *testing.T) {
	catalog := map[int]common.CatalogCheck{
		1: {Id: 0},
		2: {Id: 2},
	}

	if !shouldPublishCatalogMessage(1, catalog) {
		t.Fatal("expected id=1 to be published")
	}
	if shouldPublishCatalogMessage(2, catalog) {
		t.Fatal("did not expect id=2 to be published")
	}
	if !shouldPublishCatalogMessage(3, catalog) {
		t.Fatal("expected missing key to be published")
	}
}

func TestIdsMissingFromDaily(t *testing.T) {
	catalog := map[int]common.CatalogCheck{
		10: {Id: 10},
		20: {Id: 20},
		30: {Id: 30},
	}
	daily := map[int]bool{10: true, 30: true}

	missing := idsMissingFromDaily(catalog, daily)
	if len(missing) != 1 {
		t.Fatalf("expected 1 missing id, got %d (%v)", len(missing), missing)
	}
	if missing[0] != 20 {
		t.Fatalf("expected missing id 20, got %d", missing[0])
	}
}

func TestCollectionByIndexName(t *testing.T) {
	if collectionByIndexName("series") != "serie" {
		t.Fatalf("unexpected mapping for series: %q", collectionByIndexName("series"))
	}
	if collectionByIndexName("movies") != "movie" {
		t.Fatalf("unexpected mapping for movies: %q", collectionByIndexName("movies"))
	}
	if collectionByIndexName("persons") != "person" {
		t.Fatalf("unexpected mapping for persons: %q", collectionByIndexName("persons"))
	}
	if collectionByIndexName("unknown") != "" {
		t.Fatalf("expected empty mapping for unknown index")
	}
}

func TestShouldFlushBatch(t *testing.T) {
	if !shouldFlushBatch(0, 1000) {
		t.Fatal("expected flush on first element with interval 1000")
	}
	if shouldFlushBatch(1, 1000) {
		t.Fatal("did not expect flush for index 1 and interval 1000")
	}
	if !shouldFlushBatch(2000, 1000) {
		t.Fatal("expected flush for index divisible by interval")
	}
	if shouldFlushBatch(1, 0) {
		t.Fatal("did not expect flush when interval is zero")
	}
	if shouldFlushBatch(1, -1) {
		t.Fatal("did not expect flush when interval is negative")
	}
}

func TestBuildCatalogTvLocalized_GroupsByIdAndAccumulatesLocations(t *testing.T) {
	docs := []tv.Serie{
		{Id: 1, Title: "Serie PT", Language: common.LANGUAGE_PTBR, PosterPath: "/pt.jpg", FirstAirDate: "2020-01-01", OriginalLanguage: "ja", OriginalTitle: "S1", Popularity: 9.9},
		{Id: 1, Title: "Serie EN", Language: common.LANGUAGE_EN, PosterPath: "/en.jpg", FirstAirDate: "2020-01-01", OriginalLanguage: "ja", OriginalTitle: "S1", Popularity: 9.9},
	}

	got := buildCatalogTvLocalized(docs)
	if len(got) != 1 {
		t.Fatalf("expected 1 grouped catalog item, got %d", len(got))
	}

	item := got[0]
	if item.Id != 1 || item.CatalogType != common.MEDIA_TYPE_TV {
		t.Fatalf("unexpected item identity: %+v", item)
	}
	if item.ReleaseDate != "2020-01-01" || item.OriginalLanguage != "ja" || item.OriginalTitle != "S1" {
		t.Fatalf("unexpected static fields: %+v", item)
	}
	if len(item.Locations) != 2 {
		t.Fatalf("expected 2 localized locations, got %d", len(item.Locations))
	}
}

func TestBuildCatalogMovieLocalized_GroupsByIdAndAccumulatesLocations(t *testing.T) {
	docs := []movie.Movie{
		{Id: 2, Title: "Filme PT", Language: common.LANGUAGE_PTBR, PosterPath: "/pt.jpg", ReleaseDate: "2021-02-02", OriginalLanguage: "en", OriginalTitle: "M1", Popularity: 7.7},
		{Id: 2, Title: "Movie EN", Language: common.LANGUAGE_EN, PosterPath: "/en.jpg", ReleaseDate: "2021-02-02", OriginalLanguage: "en", OriginalTitle: "M1", Popularity: 7.7},
	}

	got := buildCatalogMovieLocalized(docs)
	if len(got) != 1 {
		t.Fatalf("expected 1 grouped catalog item, got %d", len(got))
	}

	item := got[0]
	if item.Id != 2 || item.CatalogType != common.MEDIA_TYPE_MOVIE {
		t.Fatalf("unexpected item identity: %+v", item)
	}
	if item.ReleaseDate != "2021-02-02" || item.OriginalLanguage != "en" || item.OriginalTitle != "M1" {
		t.Fatalf("unexpected static fields: %+v", item)
	}
	if len(item.Locations) != 2 {
		t.Fatalf("expected 2 localized locations, got %d", len(item.Locations))
	}
}

func TestBuildCatalogPersonDocs(t *testing.T) {
	docs := []person.Person{
		{Id: 10, Name: "Actor 1", ProfilePath: "/a1.jpg", Popularity: 1.1},
		{Id: 20, Name: "Actor 2", ProfilePath: "/a2.jpg", Popularity: 2.2},
	}

	got := buildCatalogPersonDocs(docs)
	if len(got) != 2 {
		t.Fatalf("expected 2 catalog items, got %d", len(got))
	}

	if got[0].CatalogType != common.MEDIA_TYPE_PERSON || got[1].CatalogType != common.MEDIA_TYPE_PERSON {
		t.Fatalf("unexpected catalog types: %+v", got)
	}
	if got[0].Id != 10 || got[0].Name != "Actor 1" || got[0].ProfilePath != "/a1.jpg" {
		t.Fatalf("unexpected first item: %+v", got[0])
	}
}
