package rabbitmq

import (
	"fmt"
	"testing"

	"github.com/streadway/amqp"
)

func TestDeclareQueue(t *testing.T) {
	const queueName = "rabbitmq-test-queue"
	cn, err := connection()
	if err != nil {
		t.Errorf("Failed to get connection: %s", err.Error())
	}
	ch, _ := cn.Channel()

	q, err := declareQueue(ch, queueName)
	if err != nil {
		t.Errorf("Failed to declare queue: %s", err.Error())
	}

	if q.Name != queueName {
		t.Errorf("Queue name should be '%s' not %s", queueName, q.Name)
	}

	ch.QueueDelete(queueName, true, true, false)
	ch.Close()
	cn.Close()
}

func TestDeclareExchange(t *testing.T) {
	const exchangeName = "rabbitmq-test-exchange"
	cn, err := connection()
	if err != nil {
		t.Errorf("Failed to get connection: %s", err.Error())
	}
	ch, _ := cn.Channel()

	err = declareExchange(ch, exchangeName)
	if err != nil {
		t.Errorf("Failed to declare exchange: %s", err.Error())
	}

	ch.ExchangeDelete(exchangeName, true, false)
	ch.Close()
	cn.Close()
}

func connection() (*amqp.Connection, error) {
	cn, err := amqp.Dial("amqp:localhost")

	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	return cn, nil
}
