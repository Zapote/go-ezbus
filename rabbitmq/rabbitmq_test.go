package rabbitmq

import (
	"fmt"
	"testing"

	"github.com/streadway/amqp"
)

func TestDeclareQueue(t *testing.T) {
	const queueName = "rabbitmq-test-queue"
	ch, err := channel()
	if err != nil {
		t.Errorf("Failed to get channel: %s", err.Error())
	}

	_, err = declareQueue(ch, queueName)
	if err != nil {
		t.Errorf("Failed to declare queue: %s", err.Error())
	}

	ch.QueueDelete(queueName, true, true, false)
}

func TestDeclareExchange(t *testing.T) {
	const exchangeName = "rabbitmq-test-exchange"

	ch, err := channel()
	if err != nil {
		t.Errorf("Failed to get channel: %s", err.Error())
	}

	err = declareExchange(ch, exchangeName)
	if err != nil {
		t.Errorf("Failed to declare exchange: %s", err.Error())
	}

	ch.ExchangeDelete(exchangeName, true, false)
}

func channel() (*amqp.Channel, error) {
	cn, err := amqp.Dial("amqp:localhost")

	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	ch, err := cn.Channel()

	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	return ch, nil
}
