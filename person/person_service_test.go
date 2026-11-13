package person

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
	s := NewService(nil, nil)
	if s == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestGetPersonDetailsOnApiDb(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":202,"name":"Keanu Reeves"}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, tmdbSvc)

	got := svc.GetPersonDetailsOnApiDb(202, "en")
	if got.Id != 202 || got.Name != "Keanu Reeves" {
		t.Fatalf("unexpected person: %+v", got)
	}
}

func TestApplyPersonMetadata(t *testing.T) {
	now := time.Date(2026, 6, 3, 11, 22, 33, 0, time.UTC)
	item := Person{Id: 9, Name: "Keanu Reeves"}

	got := applyPersonMetadata(item, "en", now)

	if got.Language != "en" {
		t.Fatalf("unexpected language: %s", got.Language)
	}
	if got.UpdatedAt != "03/06/2026 11:22:33" {
		t.Fatalf("unexpected updated format: %s", got.UpdatedAt)
	}
}

func TestDecidePersonUpsertAction(t *testing.T) {
	if got := decidePersonUpsertAction(0, 1, "N"); got != "insert" {
		t.Fatalf("expected insert, got %s", got)
	}
	if got := decidePersonUpsertAction(1, 1, "Y"); got != "update" {
		t.Fatalf("expected update, got %s", got)
	}
	if got := decidePersonUpsertAction(1, 1, "N"); got != "skip" {
		t.Fatalf("expected skip, got %s", got)
	}
	if got := decidePersonUpsertAction(0, 0, "Y"); got != "skip" {
		t.Fatalf("expected skip, got %s", got)
	}
}

func TestShouldPopulateDiscoveredPerson(t *testing.T) {
	if !shouldPopulateDiscoveredPerson(9) {
		t.Fatal("expected true for positive id")
	}
	if shouldPopulateDiscoveredPerson(0) {
		t.Fatal("expected false for zero id")
	}
}
