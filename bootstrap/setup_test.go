package bootstrap

import (
	"errors"
	"testing"

	"moviedb/services"
)

type fakeMongoInitializer struct{ checked bool }

func (f *fakeMongoInitializer) CheckCreateCollections() { f.checked = true }

func TestInitializeWith_LoadError(t *testing.T) {
	wantErr := errors.New("load failed")
	_, err := initializeWith(func() error { return wantErr }, func() mongoInitializer {
		return &fakeMongoInitializer{}
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestInitializeWith_Success(t *testing.T) {
	fakeMongo := &fakeMongoInitializer{}
	svc, err := initializeWith(func() error { return nil }, func() mongoInitializer {
		return fakeMongo
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if !fakeMongo.checked {
		t.Fatal("expected CheckCreateCollections to be called")
	}
}

func TestInitialize_LoadError(t *testing.T) {
	old := initializeFn
	defer func() { initializeFn = old }()

	wantErr := errors.New("init failed")
	initializeFn = func(load func() error, newMongo func() mongoInitializer) (mongoInitializer, error) {
		return nil, wantErr
	}

	svc, err := Initialize()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if svc != nil {
		t.Fatal("expected nil service on error")
	}
}

func TestInitialize_SuccessWithMongoService(t *testing.T) {
	old := initializeFn
	defer func() { initializeFn = old }()

	expected := &services.MongoService{}
	initializeFn = func(load func() error, newMongo func() mongoInitializer) (mongoInitializer, error) {
		return expected, nil
	}

	svc, err := Initialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc != expected {
		t.Fatal("expected same mongo service pointer")
	}
}

func TestInitialize_SuccessWithUnexpectedType(t *testing.T) {
	old := initializeFn
	defer func() { initializeFn = old }()

	initializeFn = func(load func() error, newMongo func() mongoInitializer) (mongoInitializer, error) {
		return &fakeMongoInitializer{}, nil
	}

	svc, err := Initialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service for unexpected initializer type")
	}
}
