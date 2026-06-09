package main

import (
	"errors"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	old := runCatalog
	defer func() { runCatalog = old }()

	runCatalog = func() error { return nil }
	if err := run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunError(t *testing.T) {
	old := runCatalog
	defer func() { runCatalog = old }()

	wantErr := errors.New("run catalog error")
	runCatalog = func() error { return wantErr }

	err := run()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestMainSuccess(t *testing.T) {
	oldRun := runCatalog
	oldHandler := mainErrorHandler
	defer func() {
		runCatalog = oldRun
		mainErrorHandler = oldHandler
	}()

	runCatalog = func() error { return nil }
	called := false
	mainErrorHandler = func(err error) { called = true }

	main()

	if called {
		t.Fatal("error handler should not be called on success")
	}
}

func TestMainError(t *testing.T) {
	oldRun := runCatalog
	oldHandler := mainErrorHandler
	defer func() {
		runCatalog = oldRun
		mainErrorHandler = oldHandler
	}()

	wantErr := errors.New("main error")
	runCatalog = func() error { return wantErr }

	var gotErr error
	mainErrorHandler = func(err error) { gotErr = err }

	main()

	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, gotErr)
	}
}
