package rabbitmq

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"gotest.tools/v3/assert"
)

func TestExtractHeadersHandlesAllValueTypes(t *testing.T) {
	h := amqp.Table{
		"string": "a value",
		"bytes":  []byte("another value"),
		"number": 42,
		"empty":  nil,
	}

	headers := extractHeaders(h)

	assert.Equal(t, "a value", headers["string"])
	assert.Equal(t, "another value", headers["bytes"])
	assert.Equal(t, "42", headers["number"])
	assert.Equal(t, "", headers["empty"])
}
