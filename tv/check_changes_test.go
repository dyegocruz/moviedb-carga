package tv

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"moviedb/queue"
	"moviedb/tmdb"
)

type fakePublisher struct {
	closed    bool
	published []queue.CatalogProcessMessage
}

func (f *fakePublisher) Close() { f.closed = true }

func (f *fakePublisher) PublishJSON(queueName string, message interface{}) error {
	if msg, ok := message.(queue.CatalogProcessMessage); ok {
		f.published = append(f.published, msg)
	}
	return nil
}

func TestCheckTvChanges_PublishesAndCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":11},{"id":22},{"id":33}]}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	publisher := &fakePublisher{}
	svc.rabbitFactory = func() (rabbitPublisher, error) {
		return publisher, nil
	}

	svc.CheckTvChanges()

	if !publisher.closed {
		t.Fatal("expected publisher to be closed")
	}
	if len(publisher.published) != 3 {
		t.Fatalf("expected 3 published messages, got %d", len(publisher.published))
	}
}

func TestHandleTvEpisodeUpdate_EarlyReturnWhenEpisodeNotFound(t *testing.T) {
	svc := NewService(nil, nil, nil)

	called := false
	svc.getEpisodeByIdLanguageFn = func(id int, language string) Episode {
		called = true
		return Episode{Id: 0}
	}

	// Should not panic and should return early without touching tmdb/mongo.
	svc.HandleTvEpisodeUpdate(42, "en")

	if !called {
		t.Fatal("expected getEpisodeByIdLanguageFn to be called")
	}
}
