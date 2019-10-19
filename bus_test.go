package ezbus

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/zapote/go-ezbus/headers"
	"gotest.tools/assert"
)

var broker = newFakeBroker()
var msg = FakeMessage{ID: "123-4"}
var rtr = NewRouter()
var b = NewBus(broker, rtr)

func TestSendCorrectDestination(t *testing.T) {
	b.Send("queue-name", msg)
	assert.Equal(t, broker.sentDst, "queue-name")
}

func TestSendHasCorrectMessageBody(t *testing.T) {
	b.Send("queue-name", msg)

	m := broker.sentMessage.(Message)

	sent := FakeMessage{}
	json.Unmarshal(m.Body, &sent)

	assert.Equal(t, msg.ID, msg.ID)
}

func TestSendHasCorrectHeaders(t *testing.T) {
	b.Send("queue-name", msg)

	m := broker.sentMessage.(Message)

	assert.Equal(t, m.Headers[headers.MessageName], "FakeMessage")
	assert.Equal(t, m.Headers[headers.MessageFullname], "ezbus.FakeMessage")
	assert.Equal(t, m.Headers[headers.Destination], "queue-name")
}

func TestHandle(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	handled := false

	rtr.Handle("FakeMessage", func(m Message) error {
		handled = true
		defer wg.Done()
		return nil
	})

	go b.Go()
	defer b.Stop()
	broker.invoke()

	wg.Wait()
	assert.Check(t, handled, "Message should be handled")
}

func TestHandleErrorShallRetryFiveTimes(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(5)
	n := 0

	rtr.Handle("FakeMessage", func(m Message) error {
		n++
		defer wg.Done()
		return errors.New("Error in message")
	})

	go b.Go()
	defer b.Stop()
	broker.invoke()
	wg.Wait()
	assert.Equal(t, n, 5)
}

func TestHandleErrorSendsToErrorQueue(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(5)
	n := 0
	rtr.Handle("FakeMessage", func(m Message) error {
		n++
		defer wg.Done()
		return errors.New("Error in message")
	})

	go b.Go()
	defer b.Stop()
	broker.invoke()
	wg.Wait()
	assert.Equal(t, broker.sentDst, fmt.Sprintf("%s-error", broker.Endpoint()))
}

func TestHandlePanicSendsToErrorQueue(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	n := 0
	rtr.Handle("FakeMessage", func(m Message) error {
		n++
		defer wg.Done()
		panic("Panicking")
	})

	go b.Go()
	defer b.Stop()
	broker.invoke()
	wg.Wait()
	assert.Equal(t, broker.sentDst, fmt.Sprintf("%s-error", broker.Endpoint()))
}

type FakeMessage struct {
	ID string
}

type FakeBroker struct {
	sentMessage interface{}
	sentDst     string
	handle      MessageHandler
	started     chan struct{}
}

func newFakeBroker() *FakeBroker {
	return &FakeBroker{
		started: make(chan struct{}),
	}
}

func (b *FakeBroker) Send(dst string, msg Message) error {
	b.sentDst = dst
	b.sentMessage = msg
	return nil
}

func (b *FakeBroker) Publish(msg Message) error {
	return nil
}

func (b *FakeBroker) Start(handle MessageHandler) error {
	b.handle = handle
	b.started <- struct{}{}
	return nil
}

func (b *FakeBroker) Stop() error {
	return nil
}

func (b *FakeBroker) Subscribe(queueName string, messageName string) error {
	return nil
}

func (b *FakeBroker) Endpoint() string {
	return "fake-broker"
}

func (b *FakeBroker) invoke() {
	<-b.started
	m := make(map[string]string)
	m[headers.MessageName] = "FakeMessage"
	b.handle(NewMessage(m, nil))
}
