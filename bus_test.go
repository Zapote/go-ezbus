package ezbus

import (
	"encoding/json"
	"testing"
)

var b = FakeBroker{}
var ep = "service.address"
var msg = FakeMessage{ID: "12300-1"}
var router = NewRouter()
var bus = NewBus(&b, router)

func TestSendCorrectDestination(t *testing.T) {
	bus.Send(ep, msg)

	if b.sentDest != ep {
		t.Errorf("'%s' should be '%s'", b.sentDest, ep)
	}
}

func TestSendCorrectMessageWithCorrectHeaders(t *testing.T) {
	bus.Send(ep, msg)

	m := b.sentMessage.(Message)
	mn := m.Headers["message-name"]

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
	handled := false
	router.Handle("FakeMessage", func(m Message) {
		handled = true
	})

	bus.listen()
	b.invoke()

	if !handled {
		t.Errorf("Message should be handled")
	}
}

type FakeMessage struct {
	ID string
}

type FakeBroker struct {
	sentMessage interface{}
	sentDest    string
	rc          chan Message
}

func (b *FakeBroker) Send(dest string, msg Message) error {
	b.sentDest = dest
	b.sentMessage = msg
	return nil
}

func (b *FakeBroker) Publish(msg Message) error {
	return nil
}

func (b *FakeBroker) Start(c chan Message) error {
	b.rc = c
	return nil
}

func (b *FakeBroker) invoke() {
	m := make(map[string]string)
	m["message-name"] = "FakeMessage"
	b.rc <- NewMessage(m, nil)
}
