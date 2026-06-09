package parameter

import "testing"

func TestNewService(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestGetByType_UsesHookWhenProvided(t *testing.T) {
	svc := NewService(nil)
	svc.getByTypeFn = func(paramType string) Parameter {
		if paramType != "tmdb" {
			t.Fatalf("unexpected param type: %s", paramType)
		}
		return Parameter{ParamType: "tmdb"}
	}

	got := svc.GetByType("tmdb")
	if got.ParamType != "tmdb" {
		t.Fatalf("unexpected parameter: %+v", got)
	}
}
