package ezbus

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/zapote/go-ezbus/assert"
)

var broker = FakeBroker{}
var msg = FakeMessage{ID: "123-4"}
var router = NewRouter()
var bus = NewBus(&broker, *router)

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

	assert.IsEqual(t, m.Headers[MessageName], "FakeMessage")
	assert.IsEqual(t, m.Headers[MessageFullname], "ezbus.FakeMessage")
	assert.IsEqual(t, m.Headers[Destination], "queueName")
}

func TestReceive(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	handled := false

	router.Handle("FakeMessage", func(m Message) {
		handled = true
		done <- struct{}{}
	})

	bus.Go()
	defer bus.Stop()
	broker.invoke()

	<-done

	assert.IsTrue(t, handled, "Message should be handled")
}

func TestReceiveErrorShallRetryFiveTimes(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	n := 0

	router.Handle("FakeMessage", func(m Message) {
		n++

		if n > 4 {
			done <- struct{}{}
		}

		panic("Error in message")
	})

	bus.Go()
	defer bus.Stop()
	broker.invoke()

	<-done

	assert.IsEqual(t, n, 5)
}

func TestReceiveErrorShallSendToErrorQueue(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	n := 0
	router.Handle("FakeMessage", func(m Message) {
		n++
		if n > 4 {
			done <- struct{}{}
		}
		panic("Error in message")
	})

	bus.Go()
	defer bus.Stop()

	broker.invoke()

	<-done
	time.Sleep(time.Millisecond * 10)
	assert.IsEqual(t, broker.sentDst, fmt.Sprintf("%s.error", broker.Endpoint()))
}

type FakeMessage struct {
	ID string
}

type FakeBroker struct {
	sentMessage interface{}
	sentDst     string
	messages    chan<- Message
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
	done := make(chan struct{})
	go func() {
		m := make(map[string]string)
		m[MessageName] = "FakeMessage"
		b.messages <- NewMessage(m, nil)
		done <- struct{}{}
	}()
	<-done
}
