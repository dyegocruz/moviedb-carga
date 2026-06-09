package app

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"moviedb/common"
	"moviedb/queue"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeMovie struct {
	mu    sync.Mutex
	calls []struct {
		id   int
		lang string
	}
}

func (f *fakeMovie) PopulateMovieByIdAndLanguage(id int, language string, _ string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, struct {
		id   int
		lang string
	}{id, language})
}

type fakeTv struct {
	mu         sync.Mutex
	serieCalls []struct {
		id   int
		lang string
	}
	episodeCalls []struct {
		id   int
		lang string
	}
}

func (f *fakeTv) PopulateSerieByIdAndLanguage(id int, language string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.serieCalls = append(f.serieCalls, struct {
		id   int
		lang string
	}{id, language})
}

func (f *fakeTv) HandleTvEpisodeUpdate(id int, language string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.episodeCalls = append(f.episodeCalls, struct {
		id   int
		lang string
	}{id, language})
}

type fakePerson struct {
	mu    sync.Mutex
	calls []struct {
		id   int
		lang string
	}
}

func (f *fakePerson) PopulatePersonByIdAndLanguage(id int, language string, _ string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, struct {
		id   int
		lang string
	}{id, language})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeBody(t *testing.T, mediaType string, id int) []byte {
	t.Helper()
	b, err := json.Marshal(queue.CatalogProcessMessage{Id: id, MediaType: mediaType})
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}
	return b
}

// waitFor polls a condition with a short timeout.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHandleCatalogMessage_InvalidJSON(t *testing.T) {
	err := handleCatalogMessage([]byte("not-json"), &fakeMovie{}, &fakeTv{}, &fakePerson{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleCatalogMessage_Movie(t *testing.T) {
	m := &fakeMovie{}
	err := handleCatalogMessage(makeBody(t, common.MEDIA_TYPE_MOVIE, 99), m, &fakeTv{}, &fakePerson{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Synchronous call (EN) must be done; async call (PT-BR) may still be running.
	waitFor(t, func() bool {
		m.mu.Lock()
		defer m.mu.Unlock()
		for _, c := range m.calls {
			if c.id == 99 && c.lang == common.LANGUAGE_EN {
				return true
			}
		}
		return false
	})
}

func TestHandleCatalogMessage_TV(t *testing.T) {
	tv := &fakeTv{}
	err := handleCatalogMessage(makeBody(t, common.MEDIA_TYPE_TV, 55), &fakeMovie{}, tv, &fakePerson{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	waitFor(t, func() bool {
		tv.mu.Lock()
		defer tv.mu.Unlock()
		for _, c := range tv.serieCalls {
			if c.id == 55 && c.lang == common.LANGUAGE_EN {
				return true
			}
		}
		return false
	})
}

func TestHandleCatalogMessage_TVEpisode(t *testing.T) {
	tv := &fakeTv{}
	err := handleCatalogMessage(makeBody(t, common.MEDIA_TYPE_TV_EPISODE, 7), &fakeMovie{}, tv, &fakePerson{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	waitFor(t, func() bool {
		tv.mu.Lock()
		defer tv.mu.Unlock()
		for _, c := range tv.episodeCalls {
			if c.id == 7 && c.lang == common.LANGUAGE_EN {
				return true
			}
		}
		return false
	})
}

func TestHandleCatalogMessage_Person(t *testing.T) {
	p := &fakePerson{}
	err := handleCatalogMessage(makeBody(t, common.MEDIA_TYPE_PERSON, 3), &fakeMovie{}, &fakeTv{}, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	waitFor(t, func() bool {
		p.mu.Lock()
		defer p.mu.Unlock()
		for _, c := range p.calls {
			if c.id == 3 && c.lang == common.LANGUAGE_EN {
				return true
			}
		}
		return false
	})
}

func TestHandleCatalogMessage_UnknownMediaType(t *testing.T) {
	m := &fakeMovie{}
	tv := &fakeTv{}
	p := &fakePerson{}

	err := handleCatalogMessage(makeBody(t, "UNKNOWN", 1), m, tv, p)
	if err != nil {
		t.Fatalf("unexpected error for unknown media type: %v", err)
	}

	// Wait a tick then verify nothing was called.
	time.Sleep(20 * time.Millisecond)
	m.mu.Lock()
	defer m.mu.Unlock()
	tv.mu.Lock()
	defer tv.mu.Unlock()
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(m.calls)+len(tv.serieCalls)+len(tv.episodeCalls)+len(p.calls) != 0 {
		t.Fatal("no service should have been called for an unknown media type")
	}
}
