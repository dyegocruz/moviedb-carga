package catalogCharge

import (
	"errors"
	"reflect"
	"testing"

	"moviedb/services"
)

func TestExecuteCatalogSearchCharge_Success(t *testing.T) {
	svc := &Service{}

	created := ""
	bulkCreated := 0
	bulkClosed := 0
	stages := map[string][][]int{
		services.CollectionSerie:  {},
		services.CollectionMovie:  {},
		services.CollectionPerson: {},
	}
	actions := make([]string, 0)

	err := svc.executeCatalogSearchCharge(
		5,
		"catalog_search",
		"catalog_search_20260603",
		func(name string) error {
			created = name
			return nil
		},
		func(workerCount int) (bulkAdder, func(), error) {
			bulkCreated++
			return &fakeBulkAdder{}, func() { bulkClosed++ }, nil
		},
		func(collection string) []int {
			switch collection {
			case services.CollectionSerie:
				return []int{1, 2}
			case services.CollectionMovie:
				return []int{3, 4}
			case services.CollectionPerson:
				return []int{5, 6}
			default:
				return nil
			}
		},
		func(ids []int, newIndexName string, bulk bulkAdder) {
			cp := append([]int{}, ids...)
			stages[services.CollectionSerie] = append(stages[services.CollectionSerie], cp)
		},
		func(ids []int, newIndexName string, bulk bulkAdder) {
			cp := append([]int{}, ids...)
			stages[services.CollectionMovie] = append(stages[services.CollectionMovie], cp)
		},
		func(ids []int, newIndexName string, bulk bulkAdder) {
			cp := append([]int{}, ids...)
			stages[services.CollectionPerson] = append(stages[services.CollectionPerson], cp)
		},
		func(alias string) ([]string, error) { return []string{"old_idx"}, nil },
		func(index, alias string) { actions = append(actions, "add:"+index+":"+alias) },
		func(index, alias string) { actions = append(actions, "remove:"+index+":"+alias) },
		func(index string) { actions = append(actions, "delete:"+index) },
		func(index string) { actions = append(actions, "count:"+index) },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created != "catalog_search_20260603" {
		t.Fatalf("unexpected created index: %s", created)
	}
	if bulkCreated != 3 || bulkClosed != 3 {
		t.Fatalf("expected 3 bulk create/close, got %d/%d", bulkCreated, bulkClosed)
	}

	wantBatches := [][]int{{1}, {2}}
	if !reflect.DeepEqual(stages[services.CollectionSerie], wantBatches) {
		t.Fatalf("unexpected serie batches: %v", stages[services.CollectionSerie])
	}
	if !reflect.DeepEqual(stages[services.CollectionMovie], [][]int{{3}, {4}}) {
		t.Fatalf("unexpected movie batches: %v", stages[services.CollectionMovie])
	}
	if !reflect.DeepEqual(stages[services.CollectionPerson], [][]int{{5}, {6}}) {
		t.Fatalf("unexpected person batches: %v", stages[services.CollectionPerson])
	}

	wantActions := []string{
		"add:catalog_search_20260603:catalog_search",
		"remove:old_idx:catalog_search",
		"delete:old_idx",
		"count:catalog_search",
	}
	if !reflect.DeepEqual(actions, wantActions) {
		t.Fatalf("unexpected alias actions: got=%v want=%v", actions, wantActions)
	}
}

func TestExecuteCatalogSearchCharge_CreateIndexError(t *testing.T) {
	svc := &Service{}
	wantErr := errors.New("create failed")

	err := svc.executeCatalogSearchCharge(
		5,
		"catalog_search",
		"catalog_search_20260603",
		func(name string) error { return wantErr },
		func(workerCount int) (bulkAdder, func(), error) {
			t.Fatal("newBulk should not be called when createIndex fails")
			return nil, nil, nil
		},
		func(collection string) []int { return nil },
		nil,
		nil,
		nil,
		func(alias string) ([]string, error) { return nil, nil },
		func(index, alias string) {},
		func(index, alias string) {},
		func(index string) {},
		func(index string) {},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestExecuteElasticChargeInsert_Success(t *testing.T) {
	svc := &Service{}

	created := ""
	bulkCreated := 0
	bulkClosed := 0
	handled := make([][]int, 0)
	actions := make([]string, 0)

	err := svc.executeElasticChargeInsert(
		"series",
		2,
		3,
		"series_20260603",
		func(name string) error {
			created = name
			return nil
		},
		func(workerCount int) (bulkAdder, func(), error) {
			bulkCreated++
			return &fakeBulkAdder{}, func() { bulkClosed++ }, nil
		},
		[]int{10, 20, 30},
		func(ids []int, newIndexName string, bulk bulkAdder) {
			cp := append([]int{}, ids...)
			handled = append(handled, cp)
		},
		func(alias string) ([]string, error) { return []string{"old_series"}, nil },
		func(index, alias string) { actions = append(actions, "add:"+index+":"+alias) },
		func(index, alias string) { actions = append(actions, "remove:"+index+":"+alias) },
		func(index string) { actions = append(actions, "delete:"+index) },
		func(index string) { actions = append(actions, "count:"+index) },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created != "series_20260603" {
		t.Fatalf("unexpected created index: %s", created)
	}
	if bulkCreated != 1 || bulkClosed != 1 {
		t.Fatalf("expected 1 bulk create/close, got %d/%d", bulkCreated, bulkClosed)
	}
	if !reflect.DeepEqual(handled, [][]int{{10}, {20, 30}}) {
		t.Fatalf("unexpected handled batches: %v", handled)
	}

	wantActions := []string{
		"add:series_20260603:series",
		"remove:old_series:series",
		"delete:old_series",
		"count:series",
	}
	if !reflect.DeepEqual(actions, wantActions) {
		t.Fatalf("unexpected alias actions: got=%v want=%v", actions, wantActions)
	}
}

func TestExecuteElasticChargeInsert_CreateIndexError(t *testing.T) {
	svc := &Service{}
	wantErr := errors.New("create failed")

	err := svc.executeElasticChargeInsert(
		"series",
		1000,
		3,
		"series_20260603",
		func(name string) error { return wantErr },
		func(workerCount int) (bulkAdder, func(), error) {
			t.Fatal("newBulk should not be called when createIndex fails")
			return nil, nil, nil
		},
		nil,
		func(ids []int, newIndexName string, bulk bulkAdder) {},
		func(alias string) ([]string, error) { return nil, nil },
		func(index, alias string) {},
		func(index, alias string) {},
		func(index string) {},
		func(index string) {},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestExecuteCatalogSearchCharge_NewBulkErrorStillRotatesAlias(t *testing.T) {
	svc := &Service{}

	actions := make([]string, 0)

	err := svc.executeCatalogSearchCharge(
		5,
		"catalog_search",
		"catalog_search_20260603",
		func(name string) error { return nil },
		func(workerCount int) (bulkAdder, func(), error) {
			return nil, nil, errors.New("bulk failed")
		},
		func(collection string) []int {
			t.Fatal("getIDs should not be called when bulk creation fails")
			return nil
		},
		func(ids []int, newIndexName string, bulk bulkAdder) { t.Fatal("unexpected tv handler call") },
		func(ids []int, newIndexName string, bulk bulkAdder) { t.Fatal("unexpected movie handler call") },
		func(ids []int, newIndexName string, bulk bulkAdder) { t.Fatal("unexpected person handler call") },
		func(alias string) ([]string, error) { return []string{"old_idx"}, nil },
		func(index, alias string) { actions = append(actions, "add:"+index+":"+alias) },
		func(index, alias string) { actions = append(actions, "remove:"+index+":"+alias) },
		func(index string) { actions = append(actions, "delete:"+index) },
		func(index string) { actions = append(actions, "count:"+index) },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantActions := []string{
		"add:catalog_search_20260603:catalog_search",
		"remove:old_idx:catalog_search",
		"delete:old_idx",
		"count:catalog_search",
	}
	if !reflect.DeepEqual(actions, wantActions) {
		t.Fatalf("unexpected alias actions: got=%v want=%v", actions, wantActions)
	}
}

func TestExecuteCatalogSearchCharge_AliasLookupErrorStillCompletes(t *testing.T) {
	svc := &Service{}

	actions := make([]string, 0)

	err := svc.executeCatalogSearchCharge(
		5,
		"catalog_search",
		"catalog_search_20260603",
		func(name string) error { return nil },
		func(workerCount int) (bulkAdder, func(), error) {
			return &fakeBulkAdder{}, func() {}, nil
		},
		func(collection string) []int { return nil },
		func(ids []int, newIndexName string, bulk bulkAdder) {},
		func(ids []int, newIndexName string, bulk bulkAdder) {},
		func(ids []int, newIndexName string, bulk bulkAdder) {},
		func(alias string) ([]string, error) { return nil, errors.New("alias error") },
		func(index, alias string) { actions = append(actions, "add:"+index+":"+alias) },
		func(index, alias string) { actions = append(actions, "remove:"+index+":"+alias) },
		func(index string) { actions = append(actions, "delete:"+index) },
		func(index string) { actions = append(actions, "count:"+index) },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No previous aliases => add + count only.
	wantActions := []string{
		"add:catalog_search_20260603:catalog_search",
		"count:catalog_search",
	}
	if !reflect.DeepEqual(actions, wantActions) {
		t.Fatalf("unexpected alias actions: got=%v want=%v", actions, wantActions)
	}
}

func TestExecuteElasticChargeInsert_NewBulkErrorStillRotatesAlias(t *testing.T) {
	svc := &Service{}

	actions := make([]string, 0)

	err := svc.executeElasticChargeInsert(
		"series",
		1000,
		3,
		"series_20260603",
		func(name string) error { return nil },
		func(workerCount int) (bulkAdder, func(), error) {
			return nil, nil, errors.New("bulk failed")
		},
		[]int{1, 2, 3},
		func(ids []int, newIndexName string, bulk bulkAdder) {
			t.Fatal("handleDocs should not be called when bulk creation fails")
		},
		func(alias string) ([]string, error) { return []string{"old_series"}, nil },
		func(index, alias string) { actions = append(actions, "add:"+index+":"+alias) },
		func(index, alias string) { actions = append(actions, "remove:"+index+":"+alias) },
		func(index string) { actions = append(actions, "delete:"+index) },
		func(index string) { actions = append(actions, "count:"+index) },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantActions := []string{
		"add:series_20260603:series",
		"remove:old_series:series",
		"delete:old_series",
		"count:series",
	}
	if !reflect.DeepEqual(actions, wantActions) {
		t.Fatalf("unexpected alias actions: got=%v want=%v", actions, wantActions)
	}
}
