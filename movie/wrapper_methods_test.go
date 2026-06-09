package movie

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"moviedb/common"
)

func TestMovieWrappers_UseHooks(t *testing.T) {
	svc := NewService(nil, nil, nil)

	svc.getAllByIdsFn = func(ids []int) []interface{} {
		return []interface{}{Movie{Id: ids[0], Title: "M1"}}
	}
	svc.getCatalogSearchInFn = func(ids []int) []Movie {
		return []Movie{{Id: ids[0], Language: "en", Title: "Catalog"}}
	}
	svc.getMovieByIdLanguageFn = func(id int, language string) Movie {
		return Movie{Id: id, Language: language, Title: "ByID"}
	}
	svc.insertMovieFn = func(itemObj Movie, language string) interface{} { return "movie-id" }
	updated := 0
	svc.updateMovieFn = func(itemObj Movie, language string) { updated++ }
	deleted := 0
	svc.deleteMovieFn = func(id int) { deleted = id }
	svc.getCountAllFn = func() int64 { return 77 }
	svc.generateMovieCatalogCheckFn = func(language string) map[int]common.CatalogCheck {
		return map[int]common.CatalogCheck{7: {Id: 7}}
	}

	all := svc.GetAllByIds([]int{1})
	if len(all) != 1 {
		t.Fatalf("expected one doc, got %d", len(all))
	}

	catalog := svc.GetCatalogSearchIn([]int{2})
	if len(catalog) != 1 || catalog[0].Id != 2 || catalog[0].Language != "en" {
		t.Fatalf("unexpected catalog docs: %+v", catalog)
	}

	byID := svc.GetMovieByIdAndLanguage(3, "pt-BR")
	if byID.Id != 3 || byID.Language != "pt-BR" {
		t.Fatalf("unexpected movie by id+lang: %+v", byID)
	}

	if inserted := svc.InsertMovie(Movie{Id: 4}, "en"); inserted != "movie-id" {
		t.Fatalf("unexpected inserted id: %v", inserted)
	}

	svc.UpdateMovie(Movie{Id: 5}, "en")
	if updated != 1 {
		t.Fatalf("expected one update call, got %d", updated)
	}

	svc.DeleteMovie(6)
	if deleted != 6 {
		t.Fatalf("expected deleted id 6, got %d", deleted)
	}

	if count := svc.GetCountAll(); count != 77 {
		t.Fatalf("unexpected count: %d", count)
	}

	catalogCheck := svc.GenerateMovieCatalogCheck("en")
	if catalogCheck[7].Id != 7 {
		t.Fatalf("unexpected catalog check: %+v", catalogCheck)
	}
}

func TestPopulateMovies_UsesHooks(t *testing.T) {
	svc := NewService(nil, nil, nil)

	svc.maxPageLoadFn = func() int { return 2 }
	svc.getDiscoverMoviesFn = func(language string, idGenre string, page string) ResultMovie {
		if page == "1" {
			return ResultMovie{Results: []Movie{{Id: 1}, {Id: 0}}}
		}
		return ResultMovie{Results: []Movie{{Id: 2}}}
	}
	svc.getMovieByIdLanguageFn = func(id int, language string) Movie {
		return Movie{Id: 0}
	}
	svc.getMovieDetailsFn = func(id int, language string) Movie {
		return Movie{Id: id, Language: language, Title: "X"}
	}

	var mu sync.Mutex
	calls := map[string]int{}
	done := make(chan struct{}, 4)
	svc.populateMovieByLanguageFn = func(itemObj Movie, language string, updateCast string) {
		mu.Lock()
		calls[language]++
		mu.Unlock()
		done <- struct{}{}
	}

	svc.PopulateMovies(common.LANGUAGE_EN, "28")

	for i := 0; i < 4; i++ {
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting populate calls")
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(calls, map[string]int{common.LANGUAGE_PTBR: 2, common.LANGUAGE_EN: 2}) {
		t.Fatalf("unexpected populate calls: %+v", calls)
	}
}
