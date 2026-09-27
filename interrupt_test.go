package ezbus

import (
	"context"
	"errors"
	"testing"

	"github.com/zapote/go-ezbus/headers"
	"gotest.tools/v3/assert"
)

func TestInterruptedMessageIsNotPutOnTheErrorQueue(t *testing.T) {
	broker := &FakeBroker{}
	router := NewRouter()
	calls := 0
	router.Handle("FakeMessage", func(m Message) error {
		calls++
		<-m.Context().Done()
		return m.Context().Err()
	})
	bus := &bus{broker: broker, router: router}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := NewMessage(map[string]string{headers.MessageName: "FakeMessage"}, nil).WithContext(ctx)

	err := bus.handle(m)

	assert.Check(t, errors.Is(err, context.Canceled))
	assert.Equal(t, 1, calls)
	assert.Equal(t, "", broker.sentDst)
}

func TestHandlerGetsTheContextOfTheMessage(t *testing.T) {
	broker := &FakeBroker{}
	router := NewRouter()
	var got context.Context
	router.Handle("FakeMessage", func(m Message) error {
		got = m.Context()
		return nil
	})
	bus := &bus{broker: broker, router: router}

	ctx := context.WithValue(context.Background(), ctxKey{}, "from-the-broker")
	m := NewMessage(map[string]string{headers.MessageName: "FakeMessage"}, nil).WithContext(ctx)

	assert.NilError(t, bus.handle(m))
	assert.Equal(t, "from-the-broker", got.Value(ctxKey{}))
}

func TestRerunFromTheErrorQueueKeepsTheContextOfTheMessage(t *testing.T) {
	broker := &FakeBroker{}
	router := NewRouter()
	var got context.Context
	router.Handle("FakeMessage", func(m Message) error {
		got = m.Context()
		return nil
	})
	bus := &bus{broker: broker, router: router}

	ctx := context.WithValue(context.Background(), ctxKey{}, "from-the-broker")
	h := map[string]string{headers.MessageName: "FakeMessage", headers.Error: "failed before"}
	m := NewMessage(h, nil).WithContext(ctx)

	assert.NilError(t, bus.handle(m))
	assert.Equal(t, "from-the-broker", got.Value(ctxKey{}))
}
