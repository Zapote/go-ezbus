package ezbus

import (
	"log"
	"testing"

	"gotest.tools/assert"
)

var r = NewRouter()

func TestInvokeCorrectHandler(t *testing.T) {
	var handled = false

	r.Handle("TestMessage", func(m Message) error {
		handled = true
		return nil
	})

	r.Receive("TestMessage", NewMessage(nil, nil))

	assert.Check(t, handled, "Message should be handled")
}

func TestNoInvokationOfHandler(t *testing.T) {
	handled := false

	r.Handle("TestMessage", func(m Message) error {
		handled = true
		return nil
	})

	r.Receive("NoMessageToHandle", NewMessage(nil, nil))

	assert.Check(t, !handled, "Message should not be handled")
}

func TestMiddlewareCalledInCorrectOrder(t *testing.T) {
	var c1, c2, c3, c4, h int
	idx := 0

	r.Handle("TestMessage", func(m Message) error {
		log.Printf("handler")
		h = idx
		idx++
		return nil
	})

	r.Middleware(func(next MessageHandler) MessageHandler {
		return func(m Message) error {
			log.Printf("bmw1")
			c1 = idx
			idx++
			next(m)
			c2 = idx
			idx++
			log.Printf("amw1")
			return nil
		}
	})

	r.Middleware(func(next MessageHandler) MessageHandler {
		return func(m Message) error {
			log.Printf("bmw2")
			c3 = idx
			idx++
			next(m)
			c4 = idx
			idx++
			log.Printf("amw2")
			return nil
		}
	})

	r.Receive("TestMessage", NewMessage(nil, nil))

	assert.Equal(t, c1, 0)
	assert.Equal(t, c3, 1)
	assert.Equal(t, h, 2)
	assert.Equal(t, c4, 3)
	assert.Equal(t, c2, 4)
}
