package rabbitmq

/*
	Needs a running RabbmitMQ on localhost:5672
*/
import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/headers"
	"gotest.tools/v3/assert"
)

// blockingHandler tells when it has a message and holds it until released.
type blockingHandler struct {
	calls   atomic.Int32
	started chan struct{}
	release chan struct{}
	then    func() error
}

func newBlockingHandler() *blockingHandler {
	return &blockingHandler{
		started: make(chan struct{}, 100),
		release: make(chan struct{}),
	}
}

func (h *blockingHandler) handle(m ezbus.Message) error {
	h.calls.Add(1)
	h.started <- struct{}{}
	<-h.release
	if h.then != nil {
		return h.then()
	}
	return nil
}

// shutdownWhileHandling begins the shutdown while the handler holds a
// message, and lets the handler go once the shutdown is waiting for it.
func shutdownWhileHandling(t *testing.T, b *Broker, h *blockingHandler) (error, time.Duration) {
	t.Helper()
	waitFor(t, h.started, 5*time.Second, "the handler to start")

	const held = 500 * time.Millisecond
	time.AfterFunc(held, func() { close(h.release) })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()
	err := b.Shutdown(ctx)
	return err, time.Since(start)
}

func TestShutdownWaitsForTheRunningHandler(t *testing.T) {
	const queue = "drain-test-running-handler"
	freshQueue(t, queue)

	h := newBlockingHandler()
	b := NewBroker(queue)
	assert.NilError(t, b.Start(h.handle))
	sendTo(t, queue, "finish-me")

	err, took := shutdownWhileHandling(t, b, h)

	assert.NilError(t, err)
	assert.Assert(t, took >= 400*time.Millisecond, "shutdown returned after %s, before the handler was done", took)
	assert.Equal(t, int32(1), h.calls.Load())
	assert.Equal(t, 0, ready(t, queue))
}

func TestShutdownPutsBufferedMessagesBackInOrder(t *testing.T) {
	const queue = "drain-test-buffered"
	freshQueue(t, queue)
	for i := 1; i <= 10; i++ {
		sendTo(t, queue, fmt.Sprint(i))
	}

	h := newBlockingHandler()
	b := NewBroker(queue, WithPrefetchCount(10))
	assert.NilError(t, b.Start(h.handle))

	err, _ := shutdownWhileHandling(t, b, h)

	assert.NilError(t, err)
	assert.Equal(t, int32(1), h.calls.Load())
	assert.DeepEqual(t, []string{"2", "3", "4", "5", "6", "7", "8", "9", "10"}, drain(t, queue))
}

func TestHandlerCanPublishDuringShutdown(t *testing.T) {
	const queue = "drain-test-publisher"
	const subscriber = "drain-test-subscriber"
	freshQueue(t, queue)
	freshQueue(t, subscriber)

	h := newBlockingHandler()
	b := NewBroker(queue)
	h.then = func() error {
		event := ezbus.NewMessage(map[string]string{headers.MessageName: "test-event"}, []byte("published-late"))
		return b.Publish(event)
	}
	assert.NilError(t, b.Start(h.handle))

	received := make(chan struct{})
	s := NewBroker(subscriber)
	assert.NilError(t, s.Start(func(m ezbus.Message) error {
		close(received)
		return nil
	}))
	defer s.Stop()
	assert.NilError(t, s.Subscribe(queue, "test-event"))

	sendTo(t, queue, "publish-when-released")
	err, _ := shutdownWhileHandling(t, b, h)

	assert.NilError(t, err)
	waitFor(t, received, 5*time.Second, "the event published during shutdown")
	assert.Equal(t, 0, ready(t, queue))
}

func TestShutdownReturnsWhenTheDeadlinePasses(t *testing.T) {
	const queue = "drain-test-deadline"
	freshQueue(t, queue)

	h := newBlockingHandler()
	defer close(h.release)
	b := NewBroker(queue)
	assert.NilError(t, b.Start(h.handle))
	sendTo(t, queue, "too-slow")
	waitFor(t, h.started, 5*time.Second, "the handler to start")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := b.Shutdown(ctx)
	took := time.Since(start)

	assert.Assert(t, errors.Is(err, context.DeadlineExceeded), "got %v", err)
	assert.Assert(t, took < 2*time.Second, "shutdown took %s", took)
	assert.Equal(t, 1, ready(t, queue))
}

func TestShutdownReturnsWhenTheServerStopsAnswering(t *testing.T) {
	for name, queue := range map[string]string{
		"consumer":  "drain-test-silent-server",
		"send only": "",
	} {
		t.Run(name, func(t *testing.T) {
			if queue != "" {
				freshQueue(t, queue)
			}
			p := newProxy(t, "localhost:5672")
			b := NewBroker(queue, WithURL(p.url()))
			assert.NilError(t, b.Start(func(m ezbus.Message) error { return nil }))
			p.mute()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			stopped := make(chan error, 1)
			go func() { stopped <- b.Shutdown(ctx) }()

			select {
			case err := <-stopped:
				assert.Assert(t, errors.Is(err, context.DeadlineExceeded), "got %v", err)
			case <-time.After(time.Second):
				t.Fatal("Shutdown was still running a second after its deadline of 100ms")
			}
		})
	}
}

func TestShutdownInterruptsTheHandlerWhenTheDeadlinePasses(t *testing.T) {
	const queue = "drain-test-interrupt"
	freshQueue(t, queue)

	var calls atomic.Int32
	started := make(chan struct{}, 10)
	interrupted := make(chan struct{}, 10)

	router := ezbus.NewRouter()
	router.Handle("test-message", func(m ezbus.Message) error {
		calls.Add(1)
		started <- struct{}{}
		select {
		case <-m.Context().Done():
			interrupted <- struct{}{}
			return m.Context().Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	})
	bus := ezbus.NewBus(NewBroker(queue), router)
	assert.NilError(t, bus.Go())
	sendTo(t, queue, "interrupt-me")
	waitFor(t, started, 5*time.Second, "the handler to start")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err := bus.Shutdown(ctx)

	assert.Assert(t, errors.Is(err, context.DeadlineExceeded), "got %v", err)
	waitFor(t, interrupted, time.Second, "the handler to be interrupted")
	eventually(t, 2*time.Second, "the message to be back on the queue", func() bool {
		return ready(t, queue) == 1
	})
	assert.Equal(t, int32(1), calls.Load())
	assert.Equal(t, 0, ready(t, queue+"-error"))
}

func TestShutdownOfASendOnlyBroker(t *testing.T) {
	b := NewBroker("")
	assert.NilError(t, b.Start(func(m ezbus.Message) error { return nil }))

	assert.NilError(t, b.Shutdown(context.Background()))
	assert.NilError(t, b.Shutdown(context.Background()))
}

func TestStopDrainsWithinTheDrainTimeout(t *testing.T) {
	const queue = "drain-test-stop"
	freshQueue(t, queue)

	h := newBlockingHandler()
	defer close(h.release)
	b := NewBroker(queue, WithDrainTimeout(300*time.Millisecond))
	assert.NilError(t, b.Start(h.handle))
	sendTo(t, queue, "too-slow")
	waitFor(t, h.started, 5*time.Second, "the handler to start")

	start := time.Now()
	err := b.Stop()
	took := time.Since(start)

	assert.Assert(t, errors.Is(err, context.DeadlineExceeded), "got %v", err)
	assert.Assert(t, took >= 300*time.Millisecond && took < 2*time.Second, "stop took %s", took)
}

func TestReconnectsWhenTheConnectionIsLost(t *testing.T) {
	const queue = "drain-test-reconnect"
	freshQueue(t, queue)
	p := newProxy(t, "localhost:5672")

	handled := make(chan struct{}, 1)
	b := NewBroker(queue, WithURL(p.url()))
	b.cfg.reconnectDelay = 50 * time.Millisecond
	assert.NilError(t, b.Start(func(m ezbus.Message) error {
		handled <- struct{}{}
		return nil
	}))
	defer b.Stop()

	p.stop()
	time.Sleep(200 * time.Millisecond)
	p.listen(t)

	sendTo(t, queue, "after-the-reconnect")
	waitFor(t, handled, 5*time.Second, "a delivery on the new connection")
	assert.Equal(t, int32(2), p.accepted.Load())
}

func TestShutdownEndsTheReconnecting(t *testing.T) {
	const queue = "drain-test-no-reconnect"
	freshQueue(t, queue)
	p := newProxy(t, "localhost:5672")

	b := NewBroker(queue, WithURL(p.url()))
	b.cfg.reconnectDelay = 50 * time.Millisecond
	assert.NilError(t, b.Start(func(m ezbus.Message) error { return nil }))

	p.stop()
	time.Sleep(200 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	assert.NilError(t, b.Shutdown(ctx))

	p.listen(t)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, int32(1), p.accepted.Load())
}

// drain takes every message off the queue and returns the bodies in the
// order they came.
func drain(t *testing.T, queue string) []string {
	t.Helper()
	var bodies []string
	withChannel(t, func(ch *amqp.Channel) {
		for {
			d, ok, err := ch.Get(queue, true)
			if err != nil {
				t.Fatalf("Get: %s", err)
			}
			if !ok {
				return
			}
			bodies = append(bodies, string(d.Body))
		}
	})
	return bodies
}
