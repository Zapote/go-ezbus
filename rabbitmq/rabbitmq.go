package rabbitmq

import (
	"github.com/streadway/amqp"
	ezbus "github.com/zapote/go-ezbus"
)

func declareQueue(c *amqp.Channel, name string) (amqp.Queue, error) {
	return c.QueueDeclare(name, true, false, false, false, nil)
}

func declareExchange(c *amqp.Channel, name string) error {
	return c.ExchangeDeclare(name, amqp.ExchangeFanout, true, false, false, false, nil)
}

func publish(c *amqp.Channel, m ezbus.Message, dst string, exchange string) error {
	headers := make(amqp.Table)

	for key, value := range m.Headers {
		headers[key] = value
	}

	return c.Publish(exchange, dst, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Headers:     headers,
			Body:        m.Body,
		})
}

func consume(c *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	return c.Consume(queueName, "", false, false, false, false, nil)
}
