package rabbitmq

import (
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/headers"
	"gotest.tools/v3/assert"
)

func TestExclusiveQueueIsTheInstancesOwn(t *testing.T) {
	host, _ := os.Hostname()
	b := NewBroker("exclusive-test-own", WithURL(testURL), WithExclusiveQueue())
	assert.NilError(t, b.Start(func(m ezbus.Message) error { return nil }))
	queue := b.Endpoint()
	assert.Equal(t, queue, "exclusive-test-own@"+host)

	withChannel(t, func(ch *amqp.Channel) {
		_, err := ch.Consume(queue, "", false, false, false, false, nil)
		assert.ErrorContains(t, err, "RESOURCE_LOCKED")
	})

	assert.NilError(t, b.Stop())

	for _, name := range []string{queue, queue + "-error"} {
		withChannel(t, func(ch *amqp.Channel) {
			_, err := ch.QueueDeclarePassive(name, false, false, true, false, nil)
			assert.ErrorContains(t, err, "NOT_FOUND", "%s outlived the connection", name)
		})
	}
}

func TestExclusiveQueueHasNoExchange(t *testing.T) {
	b := NewBroker("exclusive-test-exchange", WithURL(testURL), WithExclusiveQueue())
	assert.NilError(t, b.Start(func(m ezbus.Message) error { return nil }))
	defer b.Stop()

	withChannel(t, func(ch *amqp.Channel) {
		err := ch.ExchangeDeclarePassive(b.Endpoint(), amqp.ExchangeTopic, true, false, false, false, nil)
		assert.ErrorContains(t, err, "NOT_FOUND")
	})
	err := b.Publish(ezbus.NewMessage(map[string]string{headers.MessageName: "Saved"}, []byte("{}")))
	assert.ErrorContains(t, err, "no exchange")
}

func TestExclusiveQueueIsBoundAgainAfterAReconnect(t *testing.T) {
	const publisher = "exclusive-test-publisher"
	withChannel(t, func(ch *amqp.Channel) {
		assert.NilError(t, declareExchange(ch, publisher))
	})
	t.Cleanup(func() {
		withChannel(t, func(ch *amqp.Channel) { ch.ExchangeDelete(publisher, false, false) })
	})
	p := newProxy(t, "localhost:5672")

	handled := make(chan struct{}, 10)
	b := NewBroker("exclusive-test-reconnect", WithURL(p.url()), WithExclusiveQueue())
	b.cfg.reconnectDelay = 50 * time.Millisecond
	assert.NilError(t, b.Start(func(m ezbus.Message) error {
		handled <- struct{}{}
		return nil
	}))
	defer b.Stop()
	assert.NilError(t, b.Subscribe(publisher, "Saved"))

	publishSaved := func() {
		withChannel(t, func(ch *amqp.Channel) {
			m := ezbus.NewMessage(map[string]string{headers.MessageName: "Saved"}, []byte("{}"))
			assert.NilError(t, publish(ch, m, "Saved", publisher))
		})
	}
	publishSaved()
	waitFor(t, handled, 5*time.Second, "the event before the reconnect")

	p.stop()
	time.Sleep(200 * time.Millisecond)
	p.listen(t)

	// The queue went with the old connection; an event published before it is
	// back is lost, so publish until one arrives.
	eventually(t, 5*time.Second, "the event after the reconnect", func() bool {
		publishSaved()
		select {
		case <-handled:
			return true
		case <-time.After(100 * time.Millisecond):
			return false
		}
	})
	assert.Equal(t, int32(2), p.accepted.Load())
}
