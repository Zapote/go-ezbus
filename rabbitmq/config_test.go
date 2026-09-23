package rabbitmq

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestNewBrokerDefaults(t *testing.T) {
	b := NewBroker("my-queue")

	assert.Equal(t, "my-queue", b.Endpoint())
	assert.Assert(t, !b.sendOnly)
	assert.Equal(t, "amqp://guest:guest@localhost:5672", b.cfg.url)
	assert.Equal(t, 100, b.cfg.prefetchCount)
	assert.Equal(t, "-", b.cfg.queueNameDelimiter)
}

func TestNewBrokerOptions(t *testing.T) {
	b := NewBroker("",
		WithURL("amqp://user:secret@rabbit:5672"),
		WithPrefetchCount(10),
		WithQueueNameDelimiter("."))

	assert.Assert(t, b.sendOnly)
	assert.Equal(t, "amqp://user:secret@rabbit:5672", b.cfg.url)
	assert.Equal(t, 10, b.cfg.prefetchCount)
	assert.Equal(t, ".", b.cfg.queueNameDelimiter)
}
