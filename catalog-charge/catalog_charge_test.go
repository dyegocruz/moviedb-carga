package catalogCharge

import "testing"

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
	after(1, nil, nil, nil)
	after(2, nil, nil, assertErr{})
}

type assertErr struct{}

func (assertErr) Error() string { return "err" }
