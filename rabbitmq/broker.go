package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	ezbus "github.com/zapote/go-ezbus"
)

type Broker struct {
	queueName string
	conn      *amqp.Connection
	channel   *amqp.Channel
	done      chan (struct{})
}

//NewBroker creates a RabbitMQ broker instance
func NewBroker(queueName string) (*Broker, error) {
	b := Broker{queueName: queueName}
	b.done = make(chan struct{})

	cn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	b.conn = cn

	b.channel, err = b.conn.Channel()

	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	return &b, err
}

func (b *Broker) Send(dst string, m ezbus.Message) error {
	headers := make(amqp.Table)

	for key, value := range m.Headers {
		headers[key] = value
	}

	err := b.channel.Publish(
		"",    // exchange
		dst,   // routing key
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

func (b *Broker) Start(messages chan ezbus.Message) error {
	queue, err := b.channel.QueueDeclare(
		b.queueName, // name of the queue
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // noWait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers)", queue.Name, queue.Messages, queue.Consumers)

	msgs, err := b.channel.Consume(
		queue.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	go func() {
		for d := range msgs {
			headers := extractHeaders(d.Headers)
			m := ezbus.Message{Headers: headers, Body: d.Body}
			messages <- m
		}
	}()
	<-b.done
	return nil
}

func (b *Broker) Stop() {
	b.done <- struct{}{}
	//TODO:close channel and connection...
}

func (b *Broker) QueueName() string {
	return b.queueName
}

func extractHeaders(h amqp.Table) map[string]string {
	headers := make(map[string]string)
	for k, v := range h {
		headers[k] = v.(string)
	}
	return headers
}
