package rabbitmq

/*
	Needs a running RabbmitMQ on localhost:5672
*/
import (
	"fmt"
	"testing"

	"github.com/streadway/amqp"
	ezbus "github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/assert"
	"github.com/zapote/go-ezbus/headers"
)

const (
	queueName    = "rabbitmq-test-queue"
	exchangeName = "rabbitmq-test-exchange"
)

var cn *amqp.Connection
var channel *amqp.Channel

func TestPublishQueue(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	h := make(map[string]string)
	h["header-one"] = "test"
	b := []byte("test-message")

	cch, _ := consume(channel, queueName)

	publish(channel, ezbus.NewMessage(h, b), queueName, "")

	delivery := <-cch

	assert.IsEqual(t, delivery.ContentType, "application/json")
	assert.IsEqual(t, "test-message", string(delivery.Body))
	assert.IsEqual(t, "test", delivery.Headers["header-one"])

	delivery.Ack(false)
}

func TestPublishExchange(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	queueBind(channel, queueName, "", exchangeName)

	h := make(map[string]string)
	h[headers.MessageName] = "test-message"
	b := []byte("test-message")

	cch, _ := consume(channel, queueName)

	publish(channel, ezbus.NewMessage(h, b), "", exchangeName)

	delivery := <-cch

	assert.IsEqual(t, delivery.ContentType, "application/json")
	assert.IsEqual(t, string(delivery.Body), "test-message")

	delivery.Ack(false)
}

func setup(t *testing.T) func(t *testing.T) {
	cn, err := amqp.Dial("amqp:localhost")
	if err != nil {
		panic(fmt.Sprintf("Dial: %s", err))
	}

	channel, err = cn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Channel: %s", err))
	}

	declareQueue(channel, queueName)
	if err != nil {
		panic(fmt.Sprintf("DeclareQueue: %s", err))
	}

	declareExchange(channel, exchangeName)
	if err != nil {
		panic(fmt.Sprintf("DeclareExchange: %s", err))
	}

	channel.QueuePurge(queueName, true)

	return func(t *testing.T) {
		channel.QueueDelete(queueName, false, false, true)
		channel.ExchangeDelete(exchangeName, false, true)
		channel.Close()
		cn.Close()
	}
}
