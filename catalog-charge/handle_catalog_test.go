package catalogCharge

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olivere/elastic"
)

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
