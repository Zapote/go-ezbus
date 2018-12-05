package ezbus

import (
	"testing"
)

var r = NewRouter()

func TestInvokeCorrectHandler(t *testing.T) {
	var h = false

	r.Handle("TestMessage", func(m *Message) {
		h = true
	})

	r.handle("TestMessage", NewMessage(nil, nil))

	if !h {
		t.Error("Message not handled")
	}
}

func TestNoInvokationOfHandler(t *testing.T) {
	var h = false

	r.Handle("TestMessage", func(m *Message) {
		h = true
	})

	r.handle("NoMessageToMessage", NewMessage(nil, nil))

	if h {
		t.Error("Message not handled")
	}
}
