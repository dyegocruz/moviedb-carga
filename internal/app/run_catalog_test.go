package app

import (
	"errors"
	"testing"

	"moviedb/services"
)

type fakeCatalogRuntime struct {
	generalCalls int
	elasticCalls int
}

func (f *fakeCatalogRuntime) GeneralCatalogHandler() { f.generalCalls++ }
func (f *fakeCatalogRuntime) ElasticGeneralCharge()  { f.elasticCalls++ }

type fakeScheduler struct {
	addErr   error
	entries  map[string]func()
	started  bool
	addCalls int
}

func (f *fakeScheduler) AddFunc(spec string, cmd func()) error {
	f.addCalls++
	if f.addErr != nil {
		return f.addErr
	}
	if f.entries == nil {
		f.entries = map[string]func(){}
	}
	f.entries[spec] = cmd
	return nil
}

func (f *fakeScheduler) Start() { f.started = true }

func TestRunCatalog_InitError(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRuntimeFactory := catalogRuntimeFactory
	oldSchedulerFactory := schedulerFactory
	oldBlockForever := blockForever
	defer func() { bootstrapInitialize = oldInit }()
	defer func() { catalogRuntimeFactory = oldRuntimeFactory }()
	defer func() { schedulerFactory = oldSchedulerFactory }()
	defer func() { blockForever = oldBlockForever }()

	wantErr := errors.New("init catalog failed")
	bootstrapInitialize = func() (*services.MongoService, error) {
		return nil, wantErr
	}
	blockForever = func() {}

	err := RunCatalog()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestRunCatalog_SchedulerAddError(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRuntimeFactory := catalogRuntimeFactory
	oldSchedulerFactory := schedulerFactory
	oldBlockForever := blockForever
	defer func() { bootstrapInitialize = oldInit }()
	defer func() { catalogRuntimeFactory = oldRuntimeFactory }()
	defer func() { schedulerFactory = oldSchedulerFactory }()
	defer func() { blockForever = oldBlockForever }()

	bootstrapInitialize = func() (*services.MongoService, error) {
		return &services.MongoService{}, nil
	}
	catalogRuntimeFactory = func(_ *services.MongoService, _ services.Config) catalogRuntime {
		return &fakeCatalogRuntime{}
	}
	wantErr := errors.New("add failed")
	schedulerFactory = func() scheduler {
		return &fakeScheduler{addErr: wantErr}
	}
	blockForever = func() {}

	err := RunCatalog()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestRunCatalog_RegistersAndRunsJobs(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRuntimeFactory := catalogRuntimeFactory
	oldSchedulerFactory := schedulerFactory
	oldBlockForever := blockForever
	defer func() { bootstrapInitialize = oldInit }()
	defer func() { catalogRuntimeFactory = oldRuntimeFactory }()
	defer func() { schedulerFactory = oldSchedulerFactory }()
	defer func() { blockForever = oldBlockForever }()

	bootstrapInitialize = func() (*services.MongoService, error) {
		return &services.MongoService{}, nil
	}
	runtime := &fakeCatalogRuntime{}
	catalogRuntimeFactory = func(_ *services.MongoService, _ services.Config) catalogRuntime {
		return runtime
	}
	fs := &fakeScheduler{}
	schedulerFactory = func() scheduler { return fs }
	blockForever = func() {}

	err := RunCatalog()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fs.started {
		t.Fatal("expected scheduler to start")
	}
	if fs.addCalls != 2 {
		t.Fatalf("expected 2 cron entries, got %d", fs.addCalls)
	}

	daily, ok := fs.entries["@daily"]
	if !ok {
		t.Fatal("expected daily entry")
	}
	elastic, ok := fs.entries["0 0 3 * * *"]
	if !ok {
		t.Fatal("expected elastic entry")
	}

	daily()
	elastic()

	if runtime.generalCalls != 1 {
		t.Fatalf("expected GeneralCatalogHandler call count 1, got %d", runtime.generalCalls)
	}
	if runtime.elasticCalls != 1 {
		t.Fatalf("expected ElasticGeneralCharge call count 1, got %d", runtime.elasticCalls)
	}
}
