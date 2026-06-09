package util

import "testing"

func TestArrayContainsString(t *testing.T) {
	arr := []string{"x", "y", "z"}

	if !ArrayContainsString(arr, "y") {
		t.Fatal("expected to find element y")
	}

	if ArrayContainsString(arr, "w") {
		t.Fatal("did not expect to find element w")
	}
}

func TestStringToInt(t *testing.T) {
	got := StringToInt("42")
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestStringToIntPanicOnInvalidValue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid integer conversion")
		}
	}()

	_ = StringToInt("abc")
}
