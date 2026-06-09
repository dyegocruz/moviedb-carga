package main

import (
	"errors"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	old := runCatalogWorker
	defer func() { runCatalogWorker = old }()

	runCatalogWorker = func() error { return nil }
	if err := run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunError(t *testing.T) {
	old := runCatalogWorker
	defer func() { runCatalogWorker = old }()

	wantErr := errors.New("run worker error")
	runCatalogWorker = func() error { return wantErr }

	err := run()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestMainSuccess(t *testing.T) {
	oldRun := runCatalogWorker
	oldHandler := mainErrorHandler
	defer func() {
		runCatalogWorker = oldRun
		mainErrorHandler = oldHandler
	}()

	runCatalogWorker = func() error { return nil }
	called := false
	mainErrorHandler = func(err error) { called = true }

	main()

	if called {
		t.Fatal("error handler should not be called on success")
	}
}

func TestMainError(t *testing.T) {
	oldRun := runCatalogWorker
	oldHandler := mainErrorHandler
	defer func() {
		runCatalogWorker = oldRun
		mainErrorHandler = oldHandler
	}()

	wantErr := errors.New("main error")
	runCatalogWorker = func() error { return wantErr }

	var gotErr error
	mainErrorHandler = func(err error) { gotErr = err }

	main()

	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, gotErr)
	}
}
