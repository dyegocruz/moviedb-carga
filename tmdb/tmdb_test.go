package tmdb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"moviedb/parameter"
)

// fakeParamProvider implements ParameterProvider for tests.
type fakeParamProvider struct {
	host    string
	apiKey  string
	maxPage int
}

func (f *fakeParamProvider) GetByType(_ string) parameter.Parameter {
	return parameter.Parameter{
		Options: parameter.Options{
			TmdbHost:        f.host,
			TmdbApiKey:      f.apiKey,
			TmdbMaxPageLoad: f.maxPage,
		},
	}
}

func newTestService(t *testing.T, host string) *Service {
	t.Helper()
	return &Service{parameter: &fakeParamProvider{host: host, apiKey: "testkey", maxPage: 3}}
}

func TestMaxPageLoad(t *testing.T) {
	svc := &Service{parameter: &fakeParamProvider{maxPage: 7}}
	if got := svc.MaxPageLoad(); got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
}

func TestGetChangesByDataType_SinglePage(t *testing.T) {
	want := []ChangedElement{{Id: 10}, {Id: 20}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChangeResults{Results: want, Page: 1, TotalPages: 1}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	got := svc.GetChangesByDataType(DATATYPE_MOVIE, 1)

	if len(got) != len(want) {
		t.Fatalf("expected %d results, got %d", len(want), len(got))
	}
	for i, el := range got {
		if el.Id != want[i].Id {
			t.Errorf("result[%d]: expected id=%d, got %d", i, want[i].Id, el.Id)
		}
	}
}

func TestGetChangesByDataType_MultiPage(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		page := calls
		totalPages := 2
		resp := ChangeResults{
			Results:    []ChangedElement{{Id: page * 100}},
			Page:       page,
			TotalPages: totalPages,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	got := svc.GetChangesByDataType(DATATYPE_TV, 1)

	if len(got) != 2 {
		t.Fatalf("expected 2 results across 2 pages, got %d", len(got))
	}
}

func TestGetDetailsByIdLanguageAndDataType_Movie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("append_to_response") != "credits" {
			http.Error(w, "bad append_to_response", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	resp := svc.GetDetailsByIdLanguageAndDataType(42, "en", DATATYPE_MOVIE)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetDetailsByIdLanguageAndDataType_Person(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("append_to_response") != "combined_credits" {
			http.Error(w, "bad append_to_response", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	resp := svc.GetDetailsByIdLanguageAndDataType(1, "pt-BR", DATATYPE_PERSON)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetAlternativeTitlesByIdAndDataType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	resp := svc.GetAlternativeTitlesByIdAndDataType(5, DATATYPE_MOVIE)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetDiscoverMovies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/discover/movie" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	resp := svc.GetDiscoverMoviesByLanguageGenreAndPage("en", "28", "1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetDiscoverTv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/discover/tv" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	resp := svc.GetDiscoverTvByLanguageGenreAndPage("pt-BR", "18", "2")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetPopularPerson(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/person/popular" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := newTestService(t, server.URL)
	resp := svc.GetPopularPerson("en", "1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
