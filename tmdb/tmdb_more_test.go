package tmdb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"moviedb/parameter"
)

type fakeProviderExtra struct{ host, apiKey string }

func (f *fakeProviderExtra) GetByType(string) parameter.Parameter {
	return parameter.Parameter{Options: parameter.Options{TmdbHost: f.host, TmdbApiKey: f.apiKey, TmdbMaxPageLoad: 9}}
}

func TestNewServiceNotNil(t *testing.T) {
	s := NewService(&fakeProviderExtra{host: "http://example", apiKey: "k"})
	if s == nil {
		t.Fatal("expected service")
	}
}

func TestGetTvSeason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tv/1/season/2" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewService(&fakeProviderExtra{host: server.URL, apiKey: "k"})
	resp := s.GetTvSeason(1, 2, "en")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetTvSeasonEpisodeCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tv/1/season/2/episode/3/credits" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewService(&fakeProviderExtra{host: server.URL, apiKey: "k"})
	resp := s.GetTvSeasonEpisodeCredits(1, 2, 3, "en")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetTvSeasonEpisode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tv/1/season/2/episode/3" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("append_to_response") != "credits" {
			http.Error(w, "missing append_to_response", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewService(&fakeProviderExtra{host: server.URL, apiKey: "k"})
	resp := s.GetTvSeasonEpisode(1, 2, 3, "en")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
