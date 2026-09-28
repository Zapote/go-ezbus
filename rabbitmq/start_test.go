package rabbitmq

/*
	Needs a running RabbmitMQ on localhost:5672
*/
import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zapote/go-ezbus"
	"gotest.tools/v3/assert"
)

func TestFailedStartClosesTheConnection(t *testing.T) {
	t.Run("declare", func(t *testing.T) {
		const queue = "start-test-declare"
		freshQueue(t, queue)
		withChannel(t, func(ch *amqp.Channel) {
			ch.QueueDelete(queue, false, false, false)
			_, err := ch.QueueDeclare(queue, false, false, false, false, nil)
			assert.NilError(t, err)
		})

		assertStartFailsAndCloses(t, NewBroker(queue), "PRECONDITION_FAILED")
	})

	t.Run("consume", func(t *testing.T) {
		const queue = "start-test-consume"
		freshQueue(t, queue)
		cn, err := amqp.Dial(testURL)
		assert.NilError(t, err)
		defer cn.Close()
		ch, err := cn.Channel()
		assert.NilError(t, err)
		_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
		assert.NilError(t, err)
		_, err = ch.Consume(queue, "exclusive", false, true, false, false, nil)
		assert.NilError(t, err)

		assertStartFailsAndCloses(t, NewBroker(queue), "ACCESS_REFUSED")
	})
}

func assertStartFailsAndCloses(t *testing.T, b *Broker, reason string) {
	t.Helper()
	err := b.Start(func(m ezbus.Message) error { return nil })
	assert.ErrorContains(t, err, reason)

	b.mu.Lock()
	conn := b.conn
	b.mu.Unlock()
	assert.Assert(t, conn.IsClosed(), "the connection was left open")
}
