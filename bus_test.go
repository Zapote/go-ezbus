package ezbus

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/zapote/go-ezbus/assert"
	"github.com/zapote/go-ezbus/headers"
)

var broker = newFakeBroker()
var msg = FakeMessage{ID: "123-4"}
var rtr = NewRouter()
var bus = NewBus(broker, rtr)

func TestSendCorrectDestination(t *testing.T) {
	bus.Send("queue.name", msg)
	assert.IsEqual(t, broker.sentDst, "queue.name")
}

func TestSendHasCorrectMessageBody(t *testing.T) {
	bus.Send("queueName", msg)

	m := broker.sentMessage.(Message)

	sent := FakeMessage{}
	json.Unmarshal(m.Body, &sent)

	assert.IsEqual(t, msg.ID, msg.ID)
}

func TestSendHasCorrectHeaders(t *testing.T) {
	bus.Send("queueName", msg)

	m := broker.sentMessage.(Message)

	assert.IsEqual(t, m.Headers[headers.MessageName], "FakeMessage")
	assert.IsEqual(t, m.Headers[headers.MessageFullname], "ezbus.FakeMessage")
	assert.IsEqual(t, m.Headers[headers.Destination], "queueName")
}

func TestReceive(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	handled := false

	rtr.Handle("FakeMessage", func(m Message) {
		handled = true
		wg.Done()
	})

	go bus.Go()
	defer bus.Stop()
	broker.invoke()

	wg.Wait()
	assert.IsTrue(t, handled, "Message should be handled")
}

func TestReceiveErrorShallRetryFiveTimes(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(5)
	n := 0

	rtr.Handle("FakeMessage", func(m Message) {
		n++
		wg.Done()
		panic("Error in message")
	})

	go bus.Go()
	defer bus.Stop()
	broker.invoke()
	wg.Wait()
	assert.IsEqual(t, n, 5)
}

func TestReceiveErrorShallSendToErrorQueue(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(5)
	n := 0
	rtr.Handle("FakeMessage", func(m Message) {
		n++
		wg.Done()
		panic("Error in message")
	})

	bus.Go()
	defer bus.Stop()
	broker.invoke()
	wg.Wait()
	assert.IsEqual(t, broker.sentDst, fmt.Sprintf("%s.error", broker.Endpoint()))
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
	return "fake.broker.queue"
}

func (b *FakeBroker) invoke() {
	<-b.started
	m := make(map[string]string)
	m[headers.MessageName] = "FakeMessage"
	b.handle(NewMessage(m, nil))
}
