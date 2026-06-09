package app

import (
	"errors"
	"testing"

	"moviedb/services"
)

type fakeRabbit struct {
	prefetchErr error
	consumeErr  error
	prefetchSet int
	closed      bool
}

func (f *fakeRabbit) Close() { f.closed = true }

func (f *fakeRabbit) SetPrefetch(count int) error {
	f.prefetchSet = count
	return f.prefetchErr
}

func (f *fakeRabbit) ConsumeJSON(queueName string, handler func([]byte) error) error {
	return f.consumeErr
}

func TestRunCatalogWorker_InitError(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRabbitFactory := rabbitFactory
	defer func() {
		bootstrapInitialize = oldInit
		rabbitFactory = oldRabbitFactory
	}()

	wantErr := errors.New("init failed")
	bootstrapInitialize = func() (*services.MongoService, error) {
		return nil, wantErr
	}

	err := RunCatalogWorker()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestRunCatalogWorker_RabbitFactoryError(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRabbitFactory := rabbitFactory
	defer func() {
		bootstrapInitialize = oldInit
		rabbitFactory = oldRabbitFactory
	}()

	bootstrapInitialize = func() (*services.MongoService, error) {
		return &services.MongoService{}, nil
	}
	wantErr := errors.New("rabbit error")
	rabbitFactory = func(config services.Config) (rabbitConsumer, error) {
		return nil, wantErr
	}

	err := RunCatalogWorker()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestRunCatalogWorker_PrefetchError(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRabbitFactory := rabbitFactory
	defer func() {
		bootstrapInitialize = oldInit
		rabbitFactory = oldRabbitFactory
	}()

	bootstrapInitialize = func() (*services.MongoService, error) {
		return &services.MongoService{}, nil
	}

	fake := &fakeRabbit{prefetchErr: errors.New("prefetch error")}
	rabbitFactory = func(config services.Config) (rabbitConsumer, error) {
		return fake, nil
	}

	err := RunCatalogWorker()
	if err == nil || err.Error() != "prefetch error" {
		t.Fatalf("expected prefetch error, got %v", err)
	}
	if !fake.closed {
		t.Fatal("expected rabbit close to be called")
	}
}

func TestRunCatalogWorker_ConsumeError(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRabbitFactory := rabbitFactory
	defer func() {
		bootstrapInitialize = oldInit
		rabbitFactory = oldRabbitFactory
	}()

	bootstrapInitialize = func() (*services.MongoService, error) {
		return &services.MongoService{}, nil
	}

	fake := &fakeRabbit{consumeErr: errors.New("consume error")}
	rabbitFactory = func(config services.Config) (rabbitConsumer, error) {
		return fake, nil
	}

	err := RunCatalogWorker()
	if err == nil || err.Error() != "consume error" {
		t.Fatalf("expected consume error, got %v", err)
	}
	if !fake.closed {
		t.Fatal("expected rabbit close to be called")
	}
	if fake.prefetchSet != 10 {
		t.Fatalf("expected prefetch 10, got %d", fake.prefetchSet)
	}
}

func TestRunCatalogWorker_Success(t *testing.T) {
	oldInit := bootstrapInitialize
	oldRabbitFactory := rabbitFactory
	defer func() {
		bootstrapInitialize = oldInit
		rabbitFactory = oldRabbitFactory
	}()

	bootstrapInitialize = func() (*services.MongoService, error) {
		return &services.MongoService{}, nil
	}

	fake := &fakeRabbit{}
	rabbitFactory = func(config services.Config) (rabbitConsumer, error) {
		return fake, nil
	}

	err := RunCatalogWorker()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fake.closed {
		t.Fatal("expected rabbit close to be called")
	}
	if fake.prefetchSet != 10 {
		t.Fatalf("expected prefetch 10, got %d", fake.prefetchSet)
	}
}
