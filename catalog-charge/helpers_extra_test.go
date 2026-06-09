package catalogCharge

import (
	"reflect"
	"testing"

	"moviedb/common"
)

func TestMediaFilePrefix(t *testing.T) {
	cases := map[string]string{
		common.MEDIA_TYPE_MOVIE:  "movie_ids_",
		common.MEDIA_TYPE_TV:     "tv_series_ids_",
		common.MEDIA_TYPE_PERSON: "person_ids_",
		"unknown":                "",
	}
	for in, want := range cases {
		if got := mediaFilePrefix(in); got != want {
			t.Fatalf("mediaFilePrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOldIndexToRetire(t *testing.T) {
	if _, ok := oldIndexToRetire(nil); ok {
		t.Fatal("expected ok=false for nil")
	}
	if _, ok := oldIndexToRetire([]string{}); ok {
		t.Fatal("expected ok=false for empty")
	}
	idx, ok := oldIndexToRetire([]string{"a", "b"})
	if !ok || idx != "a" {
		t.Fatalf("expected (a,true), got (%q,%v)", idx, ok)
	}
}

func TestProcessIDsInBatchesWithInterval(t *testing.T) {
	ids := []int{1, 2, 3}
	batches := make([][]int, 0)

	processIDsInBatchesWithInterval(ids, 2, func(batch []int) {
		cp := append([]int{}, batch...)
		batches = append(batches, cp)
	})

	// shouldFlushBatch uses index modulo, so with interval=2 and 3 items:
	// i=0 flush [1], i=2 flush [2,3]
	want := [][]int{{1}, {2, 3}}
	if !reflect.DeepEqual(batches, want) {
		t.Fatalf("unexpected batches: got=%v want=%v", batches, want)
	}
}

func TestProcessIDsInBatches_Wrapper(t *testing.T) {
	ids := []int{9, 8}
	batches := make([][]int, 0)

	processIDsInBatches(ids, 1000, func(batch []int) {
		cp := append([]int{}, batch...)
		batches = append(batches, cp)
	})

	// interval=1000 keeps current behavior: flush at i=0, then remainder.
	want := [][]int{{9}, {8}}
	if !reflect.DeepEqual(batches, want) {
		t.Fatalf("unexpected batches: got=%v want=%v", batches, want)
	}
}

func TestRotateAliasAndCleanup_WithOldIndex(t *testing.T) {
	actions := make([]string, 0)

	rotateAliasAndCleanup(
		"new_idx",
		"catalog_alias",
		"catalog_search",
		[]string{"old_idx"},
		func(index, alias string) { actions = append(actions, "add:"+index+":"+alias) },
		func(index, alias string) { actions = append(actions, "remove:"+index+":"+alias) },
		func(index string) { actions = append(actions, "delete:"+index) },
		func(index string) { actions = append(actions, "count:"+index) },
	)

	want := []string{
		"add:new_idx:catalog_alias",
		"remove:old_idx:catalog_alias",
		"delete:old_idx",
		"count:catalog_search",
	}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("unexpected actions: got=%v want=%v", actions, want)
	}
}

func TestRotateAliasAndCleanup_WithoutOldIndex(t *testing.T) {
	actions := make([]string, 0)

	rotateAliasAndCleanup(
		"new_idx",
		"catalog_alias",
		"catalog_search",
		nil,
		func(index, alias string) { actions = append(actions, "add:"+index+":"+alias) },
		func(index, alias string) { actions = append(actions, "remove:"+index+":"+alias) },
		func(index string) { actions = append(actions, "delete:"+index) },
		func(index string) { actions = append(actions, "count:"+index) },
	)

	want := []string{
		"add:new_idx:catalog_alias",
		"count:catalog_search",
	}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("unexpected actions: got=%v want=%v", actions, want)
	}
}
