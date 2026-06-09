package catalogCharge

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewServiceNotNil(t *testing.T) {
	s := NewService(nil, nil, nil, nil, nil, nil)
	if s == nil {
		t.Fatal("expected non-nil service")
	}
	if s.config == nil {
		t.Fatal("expected default config to be set")
	}
}

func TestAfterWithAndWithoutError(t *testing.T) {
	// Should not panic with nil response and nil error.
	after(1, nil, nil, nil)
	// Should not panic with error.
	after(2, nil, nil, assertErr{})
}

func TestCatalogSearchCharge_UsesExecutorHook(t *testing.T) {
	svc := &Service{}
	called := false
	svc.catalogSearchChargeExecutorFn = func() {
		called = true
	}

	svc.CatalogSearchCharge()

	if !called {
		t.Fatal("expected CatalogSearchCharge executor hook to be called")
	}
}

func TestElasticChargeInsert_UsesExecutorHook(t *testing.T) {
	svc := &Service{}
	called := false
	svc.elasticChargeInsertExecutorFn = func(indexName string, interval int64, mapping string, workers int) {
		called = true
		if indexName != "series" || interval != 1000 || mapping != "{}" || workers != 3 {
			t.Fatalf("unexpected args: %s %d %s %d", indexName, interval, mapping, workers)
		}
	}

	svc.ElasticChargeInsert("series", 1000, "{}", 3)

	if !called {
		t.Fatal("expected ElasticChargeInsert executor hook to be called")
	}
}

func TestCatalogSearchCharge_RunHookPath(t *testing.T) {
	svc := &Service{}
	fixedTime := time.Date(2026, time.June, 5, 11, 22, 33, 0, time.UTC)

	called := false
	svc.nowFn = func() time.Time { return fixedTime }
	svc.catalogSearchChargeRunFn = func(workers int, indexName string, newIndexName string) error {
		called = true
		if workers != 5 {
			t.Fatalf("unexpected workers: %d", workers)
		}
		if indexName != "catalog_search" {
			t.Fatalf("unexpected indexName: %s", indexName)
		}
		if newIndexName != "catalog_search_20260605112206" {
			t.Fatalf("unexpected newIndexName: %s", newIndexName)
		}
		return nil
	}

	svc.CatalogSearchCharge()

	if !called {
		t.Fatal("expected catalogSearchChargeRunFn to be called")
	}
}

func TestCatalogSearchCharge_RunHookErrorPanics(t *testing.T) {
	svc := &Service{}
	svc.nowFn = func() time.Time { return time.Date(2026, time.June, 5, 11, 22, 33, 0, time.UTC) }
	svc.catalogSearchChargeRunFn = func(workers int, indexName string, newIndexName string) error {
		return errors.New("forced error")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(r.(error).Error(), "forced error") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	svc.CatalogSearchCharge()
}

func TestElasticChargeInsert_RunHookPath(t *testing.T) {
	svc := &Service{}
	fixedTime := time.Date(2026, time.June, 5, 12, 0, 1, 0, time.UTC)

	called := false
	svc.nowFn = func() time.Time { return fixedTime }
	svc.getAllIdsByLanguageFn = func(collection string, language string) []int {
		if collection != "serie" || language != "en" {
			t.Fatalf("unexpected args for getAllIdsByLanguageFn: %s %s", collection, language)
		}
		return []int{11, 22}
	}
	svc.elasticChargeInsertRunFn = func(indexName string, interval int64, workers int, mapping string, newIndexName string, docsIDs []int) error {
		called = true
		if indexName != "series" || interval != 1000 || workers != 3 || mapping != "{}" {
			t.Fatalf("unexpected args: %s %d %d %s", indexName, interval, workers, mapping)
		}
		if newIndexName != "series_20260605120006" {
			t.Fatalf("unexpected newIndexName: %s", newIndexName)
		}
		if !reflect.DeepEqual(docsIDs, []int{11, 22}) {
			t.Fatalf("unexpected docsIDs: %v", docsIDs)
		}
		return nil
	}

	svc.ElasticChargeInsert("series", 1000, "{}", 3)

	if !called {
		t.Fatal("expected elasticChargeInsertRunFn to be called")
	}
}

func TestElasticChargeInsert_RunHookErrorPanics(t *testing.T) {
	svc := &Service{}
	svc.nowFn = func() time.Time { return time.Date(2026, time.June, 5, 12, 0, 1, 0, time.UTC) }
	svc.getAllIdsByLanguageFn = func(collection string, language string) []int { return []int{1} }
	svc.elasticChargeInsertRunFn = func(indexName string, interval int64, workers int, mapping string, newIndexName string, docsIDs []int) error {
		return errors.New("forced error")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(r.(error).Error(), "forced error") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	svc.ElasticChargeInsert("series", 1000, "{}", 3)
}

type assertErr struct{}

func (assertErr) Error() string { return "err" }
