package rabbitmq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/headers"
)

const testURL = "amqp://guest:guest@localhost:5672/"

// freshQueue gives the test an empty queue and removes it, its error queue
// and its exchange when the test is done.
func freshQueue(t *testing.T, queue string) {
	t.Helper()
	purge(queue)
	t.Cleanup(func() {
		withChannel(t, func(ch *amqp.Channel) {
			ch.QueueDelete(queue, false, false, false)
			ch.QueueDelete(queue+"-error", false, false, false)
			ch.ExchangeDelete(queue, false, false)
		})
	})
}

func withChannel(t *testing.T, fn func(ch *amqp.Channel)) {
	t.Helper()
	cn, err := amqp.Dial(testURL)
	if err != nil {
		t.Fatalf("Dial: %s", err)
	}
	defer cn.Close()
	ch, err := cn.Channel()
	if err != nil {
		t.Fatalf("Channel: %s", err)
	}
	defer ch.Close()
	fn(ch)
}

func sendTo(t *testing.T, queue string, body string) {
	t.Helper()
	withChannel(t, func(ch *amqp.Channel) {
		m := ezbus.NewMessage(map[string]string{headers.MessageName: "test-message"}, []byte(body))
		if err := publish(ch, m, queue, ""); err != nil {
			t.Fatalf("publish: %s", err)
		}
	})
}

// ready is the number of messages on the queue that no consumer holds.
func ready(t *testing.T, queue string) int {
	t.Helper()
	n := 0
	withChannel(t, func(ch *amqp.Channel) {
		q, err := ch.QueueDeclarePassive(queue, true, false, false, false, nil)
		if err != nil {
			t.Fatalf("QueueDeclarePassive: %s", err)
		}
		n = q.Messages
	})
	return n
}

func waitFor(t *testing.T, ch <-chan struct{}, limit time.Duration, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(limit):
		t.Fatalf("timed out waiting for %s", what)
	}
}
