package ezbus

import (
	"encoding/json"
	"testing"

	"github.com/zapote/go-ezbus/assert"
)

var broker = FakeBroker{}
var ep = "service.address"
var msg = FakeMessage{ID: "12300-1"}
var router = NewRouter()
var bus = NewBus(&broker, router)

func TestSendCorrectDestination(t *testing.T) {
	bus.Send(ep, msg)

	if broker.sentDest != ep {
		t.Errorf("'%s' should be '%s'", broker.sentDest, ep)
	}
}

func TestSendCorrectMessageWithCorrectHeaders(t *testing.T) {
	bus.Send(ep, msg)

	m := broker.sentMessage.(Message)
	mn := m.Headers[MessageName]

	if mn != "FakeMessage" {
		t.Errorf("'%s' should be '%s'", mn, "FakeMessage")
	}

	sent := FakeMessage{}
	json.Unmarshal(m.Body, &sent)

	if sent.ID != msg.ID {
		t.Errorf("'%s' should be '%s'", sent.ID, msg.ID)
	}
}

func TestReceive(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	handled := false

	router.Handle("FakeMessage", func(m Message) {
		handled = true
		done <- struct{}{}
	})

	bus.Start()
	defer bus.Stop()
	broker.invoke()

	<-done

	assert.IsTrue(t, handled, "Message should be handled")
}

func TestReceiveError(t *testing.T) {
	done := make(chan struct{}, 5)
	defer close(done)
	n := 0

	router.Handle("FakeMessage", func(m Message) {
		n++

		if n > 4 {
			defer func() {
				done <- struct{}{}
			}()
		}

		panic("Error in message")
	})

	bus.Start()
	defer bus.Stop()
	broker.invoke()

	<-done

	assert.IsEqual(t, n, 5)
}

type FakeMessage struct {
	ID string
}

type FakeBroker struct {
	sentMessage interface{}
	sentDest    string
	forward     chan Message
}

func (b *FakeBroker) Send(dst string, msg Message) error {
	b.sentDest = dst
	b.sentMessage = msg
	return nil
}

func (b *FakeBroker) Publish(msg Message) error {
	return nil
}

func (b *FakeBroker) Start(forward chan Message) error {
	b.forward = forward
	return nil
}

func (b *FakeBroker) invoke() {
	go func() {
		m := make(map[string]string)
		m[MessageName] = "FakeMessage"
		b.forward <- NewMessage(m, nil)
	}()
}
