package ezbus

import (
	"log"
	"testing"
	"time"

	"github.com/zapote/go-ezbus/assert"
)

var r = NewRouter()

func TestInvokeCorrectHandler(t *testing.T) {
	var h = false

	r.Handle("TestMessage", func(m Message) {
		h = true
	})

	r.handle("TestMessage", NewMessage(nil, nil))

	assert.IsTrue(t, h, "Message should be handled")
}

func TestNoInvokationOfHandler(t *testing.T) {
	h := false

	r.Handle("TestMessage", func(m Message) {
		h = true
	})

	r.handle("NoMessageToHandle", NewMessage(nil, nil))

	assert.IsFalse(t, h, "Message should not be handled")
}

func TestMiddlewareCalledInCorrectOrder(t *testing.T) {
	var c1, c2, c3, c4, h time.Time

	r.Handle("TestMessage", func(m Message) {
		log.Printf("handler")
		h = time.Now()
	})

	r.Middleware(func(next MessageHandler) MessageHandler {
		return func(m Message) {
			log.Printf("mw1")
			c1 = time.Now()
			log.Printf("%v", c1)
			next(m)
			c2 = time.Now()

		}
	})

	r.Middleware(func(next MessageHandler) MessageHandler {
		return func(m Message) {
			log.Printf("mw2")
			c3 = time.Now()
			next(m)
			c4 = time.Now()
		}
	})

	r.handle("TestMessage", NewMessage(nil, nil))

	assert.IsTrue(t, c1.Before(h), "First middleware 'before' should be called first")
	assert.IsTrue(t, c2.After(h), "First middleware 'after' should be called after handling")
	assert.IsTrue(t, c3.Before(h), "Second middleware 'before' should be called before handle")
	assert.IsTrue(t, c3.After(c1), "Second middleware 'before' should be called after first middleware")
	assert.IsTrue(t, c4.After(h), "Second middleware 'after' should be called after handling")
	assert.IsTrue(t, c2.After(c4), "First middleware 'after' should be called after last")
}
