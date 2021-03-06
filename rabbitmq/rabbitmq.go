package rabbitmq

import (
	"fmt"

	"github.com/streadway/amqp"
	"github.com/zapote/go-ezbus"
)

func declareQueue(c *amqp.Channel, name string) (amqp.Queue, error) {
	return c.QueueDeclare(name, true, false, false, false, nil)
}

func declareExchange(c *amqp.Channel, name string) error {
	return c.ExchangeDeclare(name, amqp.ExchangeTopic, true, false, false, false, nil)
}

func queueBind(c *amqp.Channel, queueName string, messageName string, exchange string) error {
	return c.QueueBind(queueName, messageName, exchange, false, nil)
}

func publish(c *amqp.Channel, m ezbus.Message, key string, exchange string) error {
	if c == nil {
		return fmt.Errorf("channel is nil")
	}

	headers := make(amqp.Table)

	for key, value := range m.Headers {
		headers[key] = value
	}

	return c.Publish(exchange, key, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Headers:      headers,
			Body:         m.Body,
			DeliveryMode: amqp.Persistent,
		})
}

func consume(c *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	return c.Consume(queueName, "", false, false, false, false, nil)
}
