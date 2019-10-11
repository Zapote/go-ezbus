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

func TestQueueBind(t *testing.T) {
	const queueName = "rabbitmq-test-queue"
	const exchangeName = "rabbitmq-test-exchange"
	cn, err := connection()
	if err != nil {
		t.Errorf("Failed to get connection: %s", err.Error())
	}

	ch, _ := cn.Channel()

	_, err = declareQueue(ch, queueName)
	if err != nil {
		t.Errorf("Failed to declare exchange: %s", err.Error())
	}

	err = declareExchange(ch, exchangeName)
	if err != nil {
		t.Errorf("Failed to declare exchange: %s", err.Error())
	}

	err = queueBind(ch, queueName, "", exchangeName)
	if err != nil {
		t.Errorf("Failed to bind to queue: %s", err.Error())
	}

}

func TestPublishQueue(t *testing.T) {
	const queueName = "rabbitmq-test-queue"
	cn, err := connection()

	if err != nil {
		t.Errorf("Failed to get connection: %s", err.Error())
	}

	ch, _ := cn.Channel()
	ch.QueuePurge(queueName, true)
	_, err = declareQueue(ch, queueName)
	if err != nil {
		t.Errorf("Failed to declare exchange: %s", err.Error())
	}

	h := make(map[string]string)
	h["header-one"] = "test"
	b := []byte("test-message")

	cch, _ := consume(ch, queueName)

	publish(ch, ezbus.NewMessage(h, b), queueName, "")

	delivery := <-cch

	assert.IsEqual(t, delivery.ContentType, "application/json")
	assert.IsEqual(t, "test-message", string(delivery.Body))
	assert.IsEqual(t, "test", delivery.Headers["header-one"])

	delivery.Ack(false)
}

func connection() (*amqp.Connection, error) {
	cn, err := amqp.Dial("amqp:localhost")

	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	return cn, nil
}
