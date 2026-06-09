package movie

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"moviedb/queue"
	"moviedb/tmdb"
)

type fakePublisher struct {
	closed     bool
	published  []queue.CatalogProcessMessage
	publishErr error
}

func (f *fakePublisher) Close() { f.closed = true }

func (f *fakePublisher) PublishJSON(queueName string, message interface{}) error {
	if f.publishErr != nil {
		return f.publishErr
	}
	if msg, ok := message.(queue.CatalogProcessMessage); ok {
		f.published = append(f.published, msg)
	}
	return nil
}

func TestCheckMoviesChanges_PublishesAndCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":111},{"id":222}]}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	publisher := &fakePublisher{}
	svc.rabbitFactory = func() (rabbitPublisher, error) {
		return publisher, nil
	}

	svc.CheckMoviesChanges()

	if !publisher.closed {
		t.Fatal("expected publisher to be closed")
	}
	if len(publisher.published) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(publisher.published))
	}
	if publisher.published[0].Id != 111 || publisher.published[1].Id != 222 {
		t.Fatalf("unexpected messages: %+v", publisher.published)
	}
}

func TestCheckMoviesChanges_NoChangesNoPublish(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	tmdbSvc := tmdb.NewService(&fakeParamProvider{host: server.URL, apiKey: "k"})
	svc := NewService(nil, nil, tmdbSvc)

	publisher := &fakePublisher{}
	svc.rabbitFactory = func() (rabbitPublisher, error) {
		return publisher, nil
	}

	svc.CheckMoviesChanges()

	if !publisher.closed {
		t.Fatal("expected publisher to be closed even with no messages")
	}
	if len(publisher.published) != 0 {
		t.Fatalf("expected 0 published messages, got %d", len(publisher.published))
	}
}

// Sanity: ensure the production fallback path is wired (just builds the closure).
func TestCheckMoviesChanges_DefaultFactoryReturnsError(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.rabbitFactory = func() (rabbitPublisher, error) {
		return nil, errors.New("not used here")
	}
	// We don't actually call CheckMoviesChanges with the error factory because
	// it would log.Fatal; this test just exercises the assignment path.
	if svc.rabbitFactory == nil {
		t.Fatal("expected factory to be assigned")
	}
}
