package rabbitmq

import (
	"github.com/streadway/amqp"
	ezbus "github.com/zapote/go-ezbus"
)

type Broker struct {
	cn *amqp.Connection
	ch *amqp.Channel
}

//NewBroker creates a RabbitMQ broker instance
func NewBroker() (*Broker, error) {
	b := Broker{}
	cn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	b.cn = cn
	b.ch, err = b.cn.Channel()
	return &b, err
}

func (b *Broker) Send(n string, m ezbus.Message) error {
	headers := make(amqp.Table)

	for key, value := range m.Headers {
		headers[key] = value
	}

	err := b.ch.Publish(
		"",    // exchange
		n,     // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Headers:     headers,
			Body:        m.Body,
		})

	return err
}

func (b *Broker) Publish(m ezbus.Message) error {
	return nil
}

func (b *Broker) Start(chan ezbus.Message) error {
	return nil
}
