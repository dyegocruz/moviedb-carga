package catalogCharge

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"moviedb/movie"
	"moviedb/person"
	"moviedb/tv"

	"github.com/olivere/elastic"
)

type fakeBulkAdder struct {
	added []elastic.BulkableRequest
}

func (f *fakeBulkAdder) Add(r elastic.BulkableRequest) {
	f.added = append(f.added, r)
}

func TestHandleCatalogTv_UsesHookAndAddsRequests(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil)
	svc.getTvCatalogSearchInFn = func(ids []int) []tv.Serie {
		return []tv.Serie{
			{Id: 1, Language: "en", Title: "S1", OriginalTitle: "S1"},
			{Id: 2, Language: "en", Title: "S2", OriginalTitle: "S2"},
		}
	}
	adder := &fakeBulkAdder{}

	svc.handleCatalogTv([]int{1, 2}, "idx_tv", adder)

	if len(adder.added) != 2 {
		t.Fatalf("expected 2 bulk requests, got %d", len(adder.added))
	}
}

func TestHandleCatalogMovie_UsesHookAndAddsRequests(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil)
	svc.getMovieCatalogSearchInFn = func(ids []int) []movie.Movie {
		return []movie.Movie{{Id: 10, Language: "en", Title: "M"}}
	}
	adder := &fakeBulkAdder{}

	svc.handleCatalogMovie([]int{10}, "idx_mv", adder)

	if len(adder.added) != 1 {
		t.Fatalf("expected 1 request, got %d", len(adder.added))
	}
}

func TestHandleCatalogPerson_UsesHookAndAddsRequests(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil)
	svc.getPersonCatalogSearchInFn = func(language string, ids []int) []person.Person {
		if language != "en" {
			t.Fatalf("expected language en, got %s", language)
		}
		return []person.Person{{Id: 5, Name: "P"}, {Id: 6, Name: "Q"}}
	}
	adder := &fakeBulkAdder{}

	svc.handleCatalogPerson([]int{5, 6}, "idx_p", adder)

	if len(adder.added) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(adder.added))
	}
}

func TestHandleCatalogTv_EmptyDocsNoAdd(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil)
	svc.getTvCatalogSearchInFn = func(ids []int) []tv.Serie { return nil }
	adder := &fakeBulkAdder{}

	svc.handleCatalogTv([]int{}, "idx_tv", adder)

	if len(adder.added) != 0 {
		t.Fatalf("expected 0 requests, got %d", len(adder.added))
	}
}

func TestHandleElasticChargeInsertDocs_AllBranches(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil)
	svc.getTvAllByIdsFn = func(ids []int) []interface{} { return []interface{}{"a", "b"} }
	svc.getMovieAllByIdsFn = func(ids []int) []interface{} { return []interface{}{"x"} }
	svc.getPersonAllByIdsFn = func(ids []int) []interface{} { return []interface{}{"p", "q", "r"} }

	cases := map[string]int{"series": 2, "movies": 1, "persons": 3, "unknown": 0}
	for indexName, expected := range cases {
		adder := &fakeBulkAdder{}
		svc.handleElasticChargeInsertDocs(indexName, []int{1}, "new_idx", adder)
		if len(adder.added) != expected {
			t.Fatalf("indexName %s: expected %d, got %d", indexName, expected, len(adder.added))
		}
	}
}

func TestIndexNamesByAlias_ReturnsIndices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// olivere/elastic Aliases service hits /_alias/* or /<index>/_alias
		if strings.HasPrefix(r.URL.Path, "/_all/_alias") || strings.HasPrefix(r.URL.Path, "/_all/_aliases") {
			body := map[string]interface{}{
				"index_one": map[string]interface{}{
					"aliases": map[string]interface{}{
						"my_alias": map[string]interface{}{},
					},
				},
				"index_two": map[string]interface{}{
					"aliases": map[string]interface{}{
						"other_alias": map[string]interface{}{},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(body)
			return
		}
		// Root ping for client.
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"n","cluster_name":"c","version":{"number":"7.10.0"},"tagline":"t"}`))
	}))
	defer server.Close()

	client, err := elastic.NewClient(
		elastic.SetURL(server.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)
	if err != nil {
		t.Fatalf("failed to create elastic client: %v", err)
	}

	svc := NewService(nil, nil, nil, nil, nil, nil)
	indices, err := svc.IndexNamesByAlias("my_alias", client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(indices) != 1 || indices[0] != "index_one" {
		t.Fatalf("unexpected indices: %v", indices)
	}
}
