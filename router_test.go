package ezbus

import (
	"log"
	"testing"

	"github.com/zapote/go-ezbus/assert"
)

var r = NewRouter()

func TestInvokeCorrectHandler(t *testing.T) {
	var h = false

	r.Handle("TestMessage", func(m Message) {
		h = true
	})

	r.Receive("TestMessage", NewMessage(nil, nil))

	assert.IsTrue(t, h, "Message should be handled")
}

func TestNoInvokationOfHandler(t *testing.T) {
	h := false

	r.Handle("TestMessage", func(m Message) {
		h = true
	})

	r.Receive("NoMessageToHandle", NewMessage(nil, nil))

	assert.IsFalse(t, h, "Message should not be handled")
}

func TestMiddlewareCalledInCorrectOrder(t *testing.T) {
	var c1, c2, c3, c4, h int
	idx := 0

	r.Handle("TestMessage", func(m Message) {
		log.Printf("handler")
		h = idx
		idx++
	})

	r.Middleware(func(next MessageHandler) MessageHandler {
		return func(m Message) {
			log.Printf("bmw1")
			c1 = idx
			idx++
			next(m)
			c2 = idx
			idx++
			log.Printf("amw1")

		}
	})

	r.Middleware(func(next MessageHandler) MessageHandler {
		return func(m Message) {
			log.Printf("bmw2")
			c3 = idx
			idx++
			next(m)
			c4 = idx
			idx++
			log.Printf("amw2")
		}
	})

	r.Receive("TestMessage", NewMessage(nil, nil))

	assert.IsEqual(t, c1, 0)
	assert.IsEqual(t, c3, 1)
	assert.IsEqual(t, h, 2)
	assert.IsEqual(t, c4, 3)
	assert.IsEqual(t, c2, 4)

}
