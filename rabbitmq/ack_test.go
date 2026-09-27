package rabbitmq

/*
	Needs a running RabbmitMQ on localhost:5672
*/
import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zapote/go-ezbus"
	"gotest.tools/v3/assert"
)

func TestFailedHandlerPutsTheMessageBack(t *testing.T) {
	const queue = "ack-test-failed-handler"
	freshQueue(t, queue)

	var calls atomic.Int32
	handled := make(chan struct{})

	b := NewBroker(queue)
	err := b.Start(func(m ezbus.Message) error {
		if calls.Add(1) == 1 {
			return errors.New("error queue out of reach")
		}
		close(handled)
		return nil
	})
	assert.NilError(t, err)

	sendTo(t, queue, "put-me-back")
	waitFor(t, handled, 5*time.Second, "the second delivery")

	assert.NilError(t, b.Stop())
	assert.Equal(t, int32(2), calls.Load())
	assert.Equal(t, 0, ready(t, queue))
}

func TestSucceededHandlerAcksTheMessage(t *testing.T) {
	const queue = "ack-test-succeeded-handler"
	freshQueue(t, queue)

	var calls atomic.Int32
	handled := make(chan struct{})

	b := NewBroker(queue)
	err := b.Start(func(m ezbus.Message) error {
		calls.Add(1)
		close(handled)
		return nil
	})
	assert.NilError(t, err)

	sendTo(t, queue, "ack-me")
	waitFor(t, handled, 5*time.Second, "the delivery")

	// Closing the channel puts unacked messages back, so an empty queue
	// after Stop means the message was acked.
	time.Sleep(200 * time.Millisecond)
	assert.NilError(t, b.Stop())
	assert.Equal(t, int32(1), calls.Load())
	assert.Equal(t, 0, ready(t, queue))
}
