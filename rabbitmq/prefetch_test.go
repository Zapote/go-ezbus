package rabbitmq

/*
	Needs a running RabbmitMQ on localhost:5672
*/
import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestBrokerTakesOneMessageAtATime(t *testing.T) {
	const queue = "prefetch-test"
	freshQueue(t, queue)
	for i := 1; i <= 3; i++ {
		sendTo(t, queue, fmt.Sprint(i))
	}

	h := newBlockingHandler()
	b := NewBroker(queue)
	assert.NilError(t, b.Start(h.handle))
	waitFor(t, h.started, 5*time.Second, "the handler to start")

	// The two messages that wait their turn stay on the queue, for any
	// consumer to take.
	assert.Equal(t, 2, ready(t, queue))

	close(h.release)
	eventually(t, 5*time.Second, "the rest to be handled", func() bool {
		return h.calls.Load() == 3
	})
	assert.NilError(t, b.Stop())
	assert.Equal(t, 0, ready(t, queue))
}
