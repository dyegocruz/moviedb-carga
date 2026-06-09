package services

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"moviedb/configs"

	"github.com/streadway/amqp"
)

// ---------------------------------------------------------------------------
// fakeChannel implements amqpChanneler for unit tests without a real broker.
// ---------------------------------------------------------------------------

type fakeChannel struct {
	mu            sync.Mutex
	prefetchCount int
	published     []amqp.Publishing
	publishErr    error
	queueDeclErr  error
	consumeErr    error
	deliveries    chan amqp.Delivery
	closed        bool
}

func (f *fakeChannel) Qos(count, size int, global bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prefetchCount = count
	return nil
}

func (f *fakeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return amqp.Queue{Name: name}, f.queueDeclErr
}

func (f *fakeChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.publishErr != nil {
		return f.publishErr
	}
	f.published = append(f.published, msg)
	return nil
}

func (f *fakeChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if f.consumeErr != nil {
		return nil, f.consumeErr
	}
	if f.deliveries == nil {
		f.deliveries = make(chan amqp.Delivery)
	}
	return f.deliveries, nil
}

func (f *fakeChannel) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

// fakeAcker satisfies amqp.Acknowledger so we can build amqp.Delivery structs.
type fakeAcker struct {
	mu    sync.Mutex
	acked bool
}

func (a *fakeAcker) Ack(tag uint64, multiple bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.acked = true
	return nil
}
func (a *fakeAcker) Nack(tag uint64, multiple bool, requeue bool) error { return nil }
func (a *fakeAcker) Reject(tag uint64, requeue bool) error              { return nil }

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newServiceWithFake(ch *fakeChannel) *RabbitMQService {
	return &RabbitMQService{channel: ch}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestSetPrefetch(t *testing.T) {
	ch := &fakeChannel{}
	svc := newServiceWithFake(ch)

	if err := svc.SetPrefetch(5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.prefetchCount != 5 {
		t.Fatalf("expected prefetchCount=5, got %d", ch.prefetchCount)
	}
}

func TestPublishJSON_MarshalAndSend(t *testing.T) {
	ch := &fakeChannel{}
	svc := newServiceWithFake(ch)

	payload := map[string]int{"id": 42}
	if err := svc.PublishJSON("q-test", payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ch.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(ch.published))
	}

	var got map[string]int
	if err := json.Unmarshal(ch.published[0].Body, &got); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if got["id"] != 42 {
		t.Fatalf("expected id=42, got %d", got["id"])
	}
}

func TestPublishJSON_PropagatesPublishError(t *testing.T) {
	wantErr := errors.New("publish failed")
	ch := &fakeChannel{publishErr: wantErr}
	svc := newServiceWithFake(ch)

	err := svc.PublishJSON("q-test", map[string]string{"k": "v"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestPublishJSON_PropagatesQueueDeclareError(t *testing.T) {
	wantErr := errors.New("declare failed")
	ch := &fakeChannel{queueDeclErr: wantErr}
	svc := newServiceWithFake(ch)

	err := svc.PublishJSON("q-test", map[string]string{"k": "v"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestPublishJSON_PropagatesMarshalError(t *testing.T) {
	ch := &fakeChannel{}
	svc := newServiceWithFake(ch)

	// Channels are not JSON-marshallable.
	err := svc.PublishJSON("q-test", make(chan int))
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestConsumeJSON_HandlerInvokedAndAcked(t *testing.T) {
	acker := &fakeAcker{}
	deliveries := make(chan amqp.Delivery, 1)
	body, _ := json.Marshal(map[string]string{"hello": "world"})
	deliveries <- amqp.Delivery{Body: body, Acknowledger: acker}

	ch := &fakeChannel{deliveries: deliveries}
	svc := newServiceWithFake(ch)

	handlerCalled := make(chan []byte, 1)
	go func() {
		_ = svc.ConsumeJSON("q-test", func(b []byte) error {
			handlerCalled <- b
			return nil
		})
	}()

	select {
	case received := <-handlerCalled:
		var got map[string]string
		if err := json.Unmarshal(received, &got); err != nil {
			t.Fatalf("handler received invalid JSON: %v", err)
		}
		if got["hello"] != "world" {
			t.Fatalf("unexpected handler payload: %v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler was not called within timeout")
	}

	// Give the goroutine time to call Ack.
	time.Sleep(50 * time.Millisecond)
	acker.mu.Lock()
	acked := acker.acked
	acker.mu.Unlock()
	if !acked {
		t.Fatal("expected delivery to be acknowledged")
	}
}

func TestConsumeJSON_ReturnsErrorOnConsumeFail(t *testing.T) {
	wantErr := errors.New("consume failed")
	ch := &fakeChannel{consumeErr: wantErr}
	svc := newServiceWithFake(ch)

	err := svc.ConsumeJSON("q-test", func([]byte) error { return nil })
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestConsumeJSON_ReturnsErrorOnQueueDeclareFail(t *testing.T) {
	wantErr := errors.New("declare failed")
	ch := &fakeChannel{queueDeclErr: wantErr}
	svc := newServiceWithFake(ch)

	err := svc.ConsumeJSON("q-test", func([]byte) error { return nil })
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestClose_ClosesChannel(t *testing.T) {
	ch := &fakeChannel{}
	svc := newServiceWithFake(ch)

	svc.Close()

	if !ch.closed {
		t.Fatal("expected channel close to be called")
	}
}

func TestClose_NilServiceSafe(t *testing.T) {
	var svc *RabbitMQService
	// Should not panic.
	svc.Close()
}

func TestNewRabbitMQService_SuccessWithHooks(t *testing.T) {
	oldDial := rabbitDialFn
	oldChannelFactory := rabbitChannelFactoryFn
	defer func() {
		rabbitDialFn = oldDial
		rabbitChannelFactoryFn = oldChannelFactory
	}()

	dialedURL := ""
	rabbitDialFn = func(url string) (*amqp.Connection, error) {
		dialedURL = url
		return nil, nil
	}
	expectedChannel := &fakeChannel{}
	rabbitChannelFactoryFn = func(conn *amqp.Connection) (amqpChanneler, error) {
		return expectedChannel, nil
	}

	svc, err := NewRabbitMQService(fakeConfig{rabbit: configs.RabbitMQConfig{Host: "h", Port: "5672", User: "u", Password: "p"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.channel != expectedChannel {
		t.Fatal("expected channel from hook")
	}
	if dialedURL != "amqp://u:p@h:5672/" {
		t.Fatalf("unexpected dial url: %s", dialedURL)
	}
}

func TestNewRabbitMQService_DialError(t *testing.T) {
	oldDial := rabbitDialFn
	oldChannelFactory := rabbitChannelFactoryFn
	defer func() {
		rabbitDialFn = oldDial
		rabbitChannelFactoryFn = oldChannelFactory
	}()

	wantErr := errors.New("dial failed")
	rabbitDialFn = func(url string) (*amqp.Connection, error) {
		return nil, wantErr
	}
	rabbitChannelFactoryFn = func(conn *amqp.Connection) (amqpChanneler, error) {
		t.Fatal("channel factory should not be called when dial fails")
		return nil, nil
	}

	svc, err := NewRabbitMQService(fakeConfig{rabbit: configs.RabbitMQConfig{Host: "h", Port: "5672", User: "u", Password: "p"}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if svc != nil {
		t.Fatal("expected nil service on dial error")
	}
}

func TestNewRabbitMQService_ChannelError(t *testing.T) {
	oldDial := rabbitDialFn
	oldChannelFactory := rabbitChannelFactoryFn
	defer func() {
		rabbitDialFn = oldDial
		rabbitChannelFactoryFn = oldChannelFactory
	}()

	wantErr := errors.New("channel failed")
	rabbitDialFn = func(url string) (*amqp.Connection, error) {
		return nil, nil
	}
	rabbitChannelFactoryFn = func(conn *amqp.Connection) (amqpChanneler, error) {
		return nil, wantErr
	}

	svc, err := NewRabbitMQService(fakeConfig{rabbit: configs.RabbitMQConfig{Host: "h", Port: "5672", User: "u", Password: "p"}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if svc != nil {
		t.Fatal("expected nil service on channel error")
	}
}
