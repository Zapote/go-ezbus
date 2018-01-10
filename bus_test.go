package ezbus

import (
	"testing"
)

var b = FakeBroker{}
var ep = "service.address"
var msg = FakeMessage{ID: "12300-1"}
var bus = NewBus(ep, &b)

// TestSend test start func
func TestSendCorrectDestination(t *testing.T) {
	bus.Send(ep, msg)

	if b.sentDest != ep {
		t.Errorf("'%s' should be '%s'", b.sentDest, ep)
	}
}

func TestSendCorrectMessage(t *testing.T) {
	bus.Send(ep, msg)

	sent := b.sentMessage.(FakeMessage)

	if sent.ID != msg.ID {
		t.Errorf("'%s' should be '%s'", sent.ID, msg.ID)
	}
}

type FakeMessage struct {
	ID string
}

type FakeBroker struct {
	sentMessage interface{}
	sentDest    string
}

func (b *FakeBroker) send(dest string, msg interface{}) error {
	b.sentDest = dest
	b.sentMessage = msg
	return nil
}
