package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
	ezbus "github.com/zapote/go-ezbus"
)

type Broker struct {
	queueName string
	cn        *amqp.Connection
	ch        *amqp.Channel
}

//NewBroker creates a RabbitMQ broker instance
func NewBroker(q string) (*Broker, error) {
	b := Broker{queueName: q}
	defer recoverDial()
	cn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	b.cn = cn
	b.ch, err = b.cn.Channel()

	if b.ch == nil {
		log.Panicln("channel nil")
	}

	return &b, err
}

func recoverDial() {
	if err := recover(); err != nil {
		log.Println("Failed to connect amqp host.")
	}
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

func (b *Broker) Start(c chan ezbus.Message) error {
	msgs, err := b.ch.Consume(
		b.queueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			log.Println("message received")
			headers := extractHeaders(d.Headers)
			m := ezbus.Message{Headers: headers, Body: d.Body}
			c <- m
		}
	}()

	return nil
}

func (b *Broker) Stop() {

}

func extractHeaders(h amqp.Table) map[string]string {
	headers := make(map[string]string)
	for k, v := range h {
		headers[k] = v.(string)
	}
	return headers
}
