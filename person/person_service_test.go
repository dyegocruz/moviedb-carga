package person

import (
	"sync"
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

func TestPopulatePersonByIdAndLanguage_UsesHooks(t *testing.T) {
	svc := NewService(nil, nil)

	var gotGetterId int
	var gotGetterLang string
	var gotPopPerson Person
	var gotPopLang string
	var gotUpdatePerson string

	svc.getPersonDetailsFn = func(id int, language string) Person {
		gotGetterId = id
		gotGetterLang = language
		return Person{Id: id, Name: "Hooked Person"}
	}
	svc.populatePersonByLanguageFn = func(itemObj Person, language string, updatePerson string) {
		gotPopPerson = itemObj
		gotPopLang = language
		gotUpdatePerson = updatePerson
	}

	svc.PopulatePersonByIdAndLanguage(12, "en", "N")

	if gotGetterId != 12 || gotGetterLang != "en" {
		t.Fatalf("getter args mismatch: id=%d lang=%s", gotGetterId, gotGetterLang)
	}
	if gotPopPerson.Id != 12 || gotPopPerson.Name != "Hooked Person" {
		t.Fatalf("unexpected person passed to populator: %+v", gotPopPerson)
	}
	if gotPopLang != "en" || gotUpdatePerson != "N" {
		t.Fatalf("populator args mismatch: lang=%s updatePerson=%s", gotPopLang, gotUpdatePerson)
	}
}

func TestApplyPersonMetadata(t *testing.T) {
	now := time.Date(2026, 6, 3, 11, 22, 33, 0, time.UTC)
	item := Person{Id: 9, Name: "Keanu Reeves"}

	got := applyPersonMetadata(item, "en", now)

	if got.Language != "en" || got.SlugUrl != "person-9" {
		t.Fatalf("unexpected language/slug url: %+v", got)
	}
	if got.Slug == "" {
		t.Fatal("expected non-empty slug")
	}
	if got.UpdatedNew != "03/06/2026 11:22:33" {
		t.Fatalf("unexpected updated format: %s", got.UpdatedNew)
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

func TestCheckPersonChanges_UsesPopulateHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":[{"id":101},{"id":202}]}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, tmdbSvc)

	var mu sync.Mutex
	calls := map[string]int{}
	svc.populatePersonByIdFn = func(id int, language string, updatePerson string) {
		mu.Lock()
		defer mu.Unlock()
		calls[language]++
		if updatePerson != "Y" {
			t.Fatalf("unexpected update flag: %s", updatePerson)
		}
	}

	svc.CheckPersonChanges()
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if calls["pt-BR"] != 2 {
		t.Fatalf("expected 2 pt-BR calls, got %d", calls["pt-BR"])
	}
	if calls["en"] != 2 {
		t.Fatalf("expected 2 en calls, got %d", calls["en"])
	}
}
