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

func TestGetTvAlternativeTitlesById(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tv/33/alternative_titles" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"id":33,"results":[{"iso_3166_1":"JP","title":"Romaji Name","type":"Romaji"},{"iso_3166_1":"BR","title":"Titulo BR","type":""}]}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	alts := svc.GetTvAlternativeTitlesById(33)
	if alts[common.LANGUAGE_ISO_JP] != "Romaji Name" {
		t.Fatalf("expected JP Romaji title, got %+v", alts)
	}
	if alts[common.LANGUAGE_ISO_BR] != "Titulo BR" {
		t.Fatalf("expected BR title, got %+v", alts)
	}
}

func TestGetSerieDetailsOnTMDBApi_UsesAlternativeTitleForJapanesePtBr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tv/33":
			_, _ = w.Write([]byte(`{"id":33,"name":"Original","original_language":"ja"}`))
		case "/tv/33/alternative_titles":
			_, _ = w.Write([]byte(`{"id":33,"results":[{"iso_3166_1":"BR","title":"Titulo Brasileiro","type":""}]}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	got := svc.GetSerieDetailsOnTMDBApi(33, common.LANGUAGE_PTBR)
	if got.Title != "Titulo Brasileiro" {
		t.Fatalf("expected localized title, got %q", got.Title)
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

func TestPopulateSerieByIdAndLanguage_UsesHooks(t *testing.T) {
	svc := NewService(nil, nil, nil)

	var gotGetterId int
	var gotGetterLang string
	var gotPopSerie Serie
	var gotPopLang string

	svc.getSerieDetailsFn = func(id int, language string) Serie {
		gotGetterId = id
		gotGetterLang = language
		return Serie{Id: id, Title: "Hooked Serie"}
	}
	svc.populateSerieByLanguageFn = func(itemObj Serie, language string) {
		gotPopSerie = itemObj
		gotPopLang = language
	}

	svc.PopulateSerieByIdAndLanguage(88, "pt-BR")

	if gotGetterId != 88 || gotGetterLang != "pt-BR" {
		t.Fatalf("getter args mismatch: id=%d lang=%s", gotGetterId, gotGetterLang)
	}
	if gotPopSerie.Id != 88 || gotPopSerie.Title != "Hooked Serie" {
		t.Fatalf("unexpected serie passed to populator: %+v", gotPopSerie)
	}
	if gotPopLang != "pt-BR" {
		t.Fatalf("unexpected populator language: %s", gotPopLang)
	}
}

func TestApplySerieMetadata(t *testing.T) {
	now := time.Date(2026, 6, 3, 9, 8, 7, 0, time.UTC)
	item := Serie{Id: 30, Title: "Shogun"}

	got := applySerieMetadata(item, "pt-BR", now)

	if got.MediaType != "serie" || got.Language != "pt-BR" {
		t.Fatalf("unexpected media/language: %+v", got)
	}
	if got.SlugUrl != "serie-30" {
		t.Fatalf("unexpected slug url: %s", got.SlugUrl)
	}
	if got.Slug == "" {
		t.Fatal("expected non-empty slug")
	}
	if got.UpdatedNew != "03/06/2026 09:08:07" {
		t.Fatalf("unexpected updated format: %s", got.UpdatedNew)
	}
}

func TestApplyLocalizedSerieTitle_BrPriority(t *testing.T) {
	serie := Serie{Title: "Original", OriginalLanguage: common.LANGUAGE_JA}
	alts := map[string]string{common.LANGUAGE_ISO_BR: "Titulo BR", common.LANGUAGE_ISO_JP: "Titulo JP"}

	got := applyLocalizedSerieTitle(serie, common.LANGUAGE_PTBR, alts)
	if got.Title != "Titulo BR" {
		t.Fatalf("expected BR title priority, got %q", got.Title)
	}
}

func TestApplyLocalizedSerieTitle_JpFallback(t *testing.T) {
	serie := Serie{Title: "Original", OriginalLanguage: common.LANGUAGE_JA}
	alts := map[string]string{common.LANGUAGE_ISO_JP: "Titulo JP"}

	got := applyLocalizedSerieTitle(serie, common.LANGUAGE_PTBR, alts)
	if got.Title != "Titulo JP" {
		t.Fatalf("expected JP fallback title, got %q", got.Title)
	}
}

func TestApplyLocalizedSerieTitle_NoOverrideForEnglish(t *testing.T) {
	serie := Serie{Title: "Original", OriginalLanguage: common.LANGUAGE_JA}
	alts := map[string]string{common.LANGUAGE_ISO_BR: "Titulo BR"}

	got := applyLocalizedSerieTitle(serie, common.LANGUAGE_EN, alts)
	if got.Title != "Original" {
		t.Fatalf("did not expect title override for english, got %q", got.Title)
	}
}

func TestShouldFetchAlternativeTitles(t *testing.T) {
	if !shouldFetchAlternativeTitles(common.LANGUAGE_PTBR, common.LANGUAGE_JA) {
		t.Fatal("expected true for pt-BR and original ja")
	}
	if shouldFetchAlternativeTitles(common.LANGUAGE_EN, common.LANGUAGE_JA) {
		t.Fatal("expected false for english")
	}
	if shouldFetchAlternativeTitles(common.LANGUAGE_PTBR, common.LANGUAGE_EN) {
		t.Fatal("expected false for non-japanese original language")
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
	if !shouldUpdateLatestSeasonEpisode(3, 3, 10, 20) {
		t.Fatal("expected update for latest season and recent episodes")
	}
	if shouldUpdateLatestSeasonEpisode(2, 3, 10, 20) {
		t.Fatal("did not expect update for non-latest season")
	}
	if shouldUpdateLatestSeasonEpisode(3, 3, 5, 20) {
		t.Fatal("did not expect update for old episode")
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
