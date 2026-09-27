package ezbus

import (
	"context"
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

// lifecycleBroker records the order the bus starts and stops it in.
type lifecycleBroker struct {
	FakeBroker
	calls        chan string
	startErr     error
	subscribeErr error
	stopErr      error
}

func newLifecycleBroker() *lifecycleBroker {
	return &lifecycleBroker{calls: make(chan string, 10)}
}

func (b *lifecycleBroker) Start(handle MessageHandler) error {
	b.calls <- "start"
	return b.startErr
}

func (b *lifecycleBroker) Subscribe(queueName string, messageName string) error {
	return b.subscribeErr
}

func (b *lifecycleBroker) Stop() error {
	b.calls <- "stop"
	return b.stopErr
}

func TestRunStartsWaitsAndStops(t *testing.T) {
	broker := newLifecycleBroker()
	bus := NewBus(broker, NewRouter())
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- bus.Run(ctx) }()

	assert.Equal(t, "start", <-broker.calls)
	select {
	case <-done:
		t.Fatal("Run returned before the context had ended")
	case call := <-broker.calls:
		t.Fatalf("%s was called before the context had ended", call)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()

	assert.NilError(t, <-done)
	assert.Equal(t, "stop", <-broker.calls)
}

func TestRunReturnsTheErrorOfTheStart(t *testing.T) {
	broker := newLifecycleBroker()
	broker.startErr = errors.New("no broker to connect to")
	bus := NewBus(broker, NewRouter())

	err := bus.Run(context.Background())

	assert.ErrorContains(t, err, "no broker to connect to")
	assert.Equal(t, "start", <-broker.calls)
	assert.Equal(t, 0, len(broker.calls))
}

func TestRunStopsTheBrokerWhenASubscriptionFails(t *testing.T) {
	broker := newLifecycleBroker()
	broker.subscribeErr = errors.New("no exchange to bind to")
	bus := NewBus(broker, NewRouter())
	bus.SubscribeMessage("other-queue", "some-message")

	err := bus.Run(context.Background())

	assert.ErrorContains(t, err, "no exchange to bind to")
	assert.Equal(t, "start", <-broker.calls)
	assert.Equal(t, 1, len(broker.calls), "the broker was left running")
	assert.Equal(t, "stop", <-broker.calls)
}

func TestRunReturnsTheErrorOfTheStop(t *testing.T) {
	broker := newLifecycleBroker()
	broker.stopErr = context.DeadlineExceeded
	bus := NewBus(broker, NewRouter())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bus.Run(ctx)

	assert.Check(t, errors.Is(err, context.DeadlineExceeded))
}
