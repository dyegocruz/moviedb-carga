package util

import "testing"

func TestReverseArray(t *testing.T) {
	input := []string{"a", "b", "c"}
	got := ReverseArray(input)

	if len(got) != 3 {
		t.Fatalf("expected len 3, got %d", len(got))
	}

	if got[0] != "c" || got[1] != "b" || got[2] != "a" {
		t.Fatalf("unexpected reverse result: %#v", got)
	}

	if input[0] != "a" {
		t.Fatalf("input slice must not be changed, got %#v", input)
	}
}

func TestReverseArrayEmpty(t *testing.T) {
	got := ReverseArray([]string{})
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %#v", got)
	}
}
