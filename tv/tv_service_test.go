package tv

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moviedb/common"
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

func TestGetSerieDetailsOnTMDBApi_ReturnsRequestedLanguageDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tv/33" || r.URL.Query().Get("language") != common.LANGUAGE_PTBR {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"id":33,"name":"Titulo Brasileiro","original_language":"ja"}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	got := svc.GetSerieDetailsOnTMDBApi(33, common.LANGUAGE_PTBR)
	if got.Title != "Titulo Brasileiro" {
		t.Fatalf("expected requested-language title, got %q", got.Title)
	}
}

func TestGetSerieDetailsOnTMDBApi_EnglishDoesNotOverrideTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tv/44" {
			_, _ = w.Write([]byte(`{"id":44,"name":"Name EN","original_language":"ja"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	got := svc.GetSerieDetailsOnTMDBApi(44, common.LANGUAGE_EN)
	if got.Title != "Name EN" {
		t.Fatalf("expected original english title, got %q", got.Title)
	}
}

func TestApplySerieMetadata(t *testing.T) {
	now := time.Date(2026, 6, 3, 9, 8, 7, 0, time.UTC)
	item := Serie{Id: 30, Title: "Shogun"}

	got := applySerieMetadata(item, "pt-BR", now)

	if got.MediaType != "serie" || got.Language != "pt-BR" {
		t.Fatalf("unexpected media/language: %+v", got)
	}
	if got.UpdatedAt != "03/06/2026 09:08:07" {
		t.Fatalf("unexpected updated format: %s", got.UpdatedAt)
	}
}

func TestShouldInsertEpisode(t *testing.T) {
	if !shouldInsertEpisode(0) {
		t.Fatal("expected insert for missing episode")
	}
	if shouldInsertEpisode(10) {
		t.Fatal("expected no insert for existing episode")
	}
}

func TestShouldUpdateLatestSeasonEpisode(t *testing.T) {
	oldUpdatedAt := time.Now().AddDate(0, 0, -8).Format("02/01/2006 15:04:05")
	if !shouldUpdateLatestSeasonEpisode(
		Episode{SeasonNumber: 3, EpisodeNumber: 10, UpdatedAt: oldUpdatedAt},
		Season{EpisodeCount: 20},
		3,
	) {
		t.Fatal("expected update for an episode not updated in the last week")
	}
	if shouldUpdateLatestSeasonEpisode(
		Episode{SeasonNumber: 2, EpisodeNumber: 10, UpdatedAt: oldUpdatedAt},
		Season{EpisodeCount: 50},
		3,
	) {
		t.Fatal("did not expect update for an old episode in a non-latest season")
	}
	if !shouldUpdateLatestSeasonEpisode(
		Episode{SeasonNumber: 2, EpisodeNumber: 25},
		Season{EpisodeCount: 50},
		3,
	) {
		t.Fatal("expected update for a recent episode without an update timestamp")
	}
	if shouldUpdateLatestSeasonEpisode(
		Episode{SeasonNumber: 3, EpisodeNumber: 10, UpdatedAt: time.Now().Format("02/01/2006 15:04:05")},
		Season{EpisodeCount: 20},
		3,
	) {
		t.Fatal("did not expect update for a recently updated episode")
	}
}

func TestDecideSerieUpsertAction(t *testing.T) {
	if got := decideSerieUpsertAction(0, 9); got != "insert" {
		t.Fatalf("expected insert, got %s", got)
	}
	if got := decideSerieUpsertAction(7, 9); got != "update" {
		t.Fatalf("expected update, got %s", got)
	}
	if got := decideSerieUpsertAction(0, 0); got != "skip" {
		t.Fatalf("expected skip, got %s", got)
	}
}

func TestShouldPopulateDiscoveredSerie(t *testing.T) {
	if !shouldPopulateDiscoveredSerie(4, 0) {
		t.Fatal("expected true for new serie")
	}
	if shouldPopulateDiscoveredSerie(0, 0) {
		t.Fatal("expected false for invalid id")
	}
	if shouldPopulateDiscoveredSerie(4, 4) {
		t.Fatal("expected false for existing serie")
	}
}
