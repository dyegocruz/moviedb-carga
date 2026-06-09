package movie

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moviedb/parameter"
	"moviedb/tmdb"
)

type fakeParamProvider struct{ host, apiKey string }

func (f *fakeParamProvider) GetByType(string) parameter.Parameter {
	return parameter.Parameter{Options: parameter.Options{TmdbHost: f.host, TmdbApiKey: f.apiKey, TmdbMaxPageLoad: 1}}
}

func TestNewService(t *testing.T) {
	s := NewService(nil, nil, nil)
	if s == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestGetMovieDetailsOnTMDBApi(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":101,"title":"Matrix"}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	got := svc.GetMovieDetailsOnTMDBApi(101, "en")
	if got.Id != 101 || got.Title != "Matrix" {
		t.Fatalf("unexpected movie: %+v", got)
	}
}

func TestPopulateMovieByIdAndLanguage_UsesHooks(t *testing.T) {
	svc := NewService(nil, nil, nil)

	var gotGetterId int
	var gotGetterLang string
	var gotPopMovie Movie
	var gotPopLang string
	var gotUpdateCast string

	svc.getMovieDetailsFn = func(id int, language string) Movie {
		gotGetterId = id
		gotGetterLang = language
		return Movie{Id: id, Title: "Hooked"}
	}
	svc.populateMovieByLanguageFn = func(itemObj Movie, language string, updateCast string) {
		gotPopMovie = itemObj
		gotPopLang = language
		gotUpdateCast = updateCast
	}

	svc.PopulateMovieByIdAndLanguage(77, "pt-BR", "Y")

	if gotGetterId != 77 || gotGetterLang != "pt-BR" {
		t.Fatalf("getter args mismatch: id=%d lang=%s", gotGetterId, gotGetterLang)
	}
	if gotPopMovie.Id != 77 || gotPopMovie.Title != "Hooked" {
		t.Fatalf("unexpected movie passed to populator: %+v", gotPopMovie)
	}
	if gotPopLang != "pt-BR" || gotUpdateCast != "Y" {
		t.Fatalf("populator args mismatch: lang=%s updateCast=%s", gotPopLang, gotUpdateCast)
	}
}

func TestApplyMovieMetadata(t *testing.T) {
	now := time.Date(2026, 6, 3, 10, 20, 30, 0, time.UTC)
	item := Movie{Id: 11, Title: "Cidade de Deus"}

	got := applyMovieMetadata(item, "pt-BR", now)

	if got.MediaType != "movie" || got.Language != "pt-BR" {
		t.Fatalf("unexpected media/language: %+v", got)
	}
	if got.SlugUrl != "movie-11" {
		t.Fatalf("unexpected slug url: %s", got.SlugUrl)
	}
	if got.Slug == "" {
		t.Fatal("expected non-empty slug")
	}
	if got.UpdatedNew != "03/06/2026 10:20:30" {
		t.Fatalf("unexpected updated format: %s", got.UpdatedNew)
	}
}

func TestDecideMovieUpsertAction(t *testing.T) {
	if got := decideMovieUpsertAction(0, 10); got != "insert" {
		t.Fatalf("expected insert, got %s", got)
	}
	if got := decideMovieUpsertAction(5, 10); got != "update" {
		t.Fatalf("expected update, got %s", got)
	}
	if got := decideMovieUpsertAction(0, 0); got != "skip" {
		t.Fatalf("expected skip, got %s", got)
	}
}

func TestShouldPopulateDiscoveredMovie(t *testing.T) {
	if !shouldPopulateDiscoveredMovie(10, 0) {
		t.Fatal("expected true for new valid movie")
	}
	if shouldPopulateDiscoveredMovie(0, 0) {
		t.Fatal("expected false for invalid id")
	}
	if shouldPopulateDiscoveredMovie(10, 10) {
		t.Fatal("expected false for existing movie")
	}
}
