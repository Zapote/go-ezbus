package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	ezbus "github.com/zapote/go-ezbus"
)

//Broker RabbitMQ implementation of ezbus.broker interaface.
type Broker struct {
	queueName string
	conn      *amqp.Connection
	channel   *amqp.Channel
	cfg       config
	done      chan (struct{})
}

//NewBroker creates a RabbitMQ broker instance
//Default url amqp://guest:guest@localhost:5672
//Default prefetchCount 100
func NewBroker(queueName string) *Broker {
	b := Broker{queueName: queueName, done: make(chan (struct{}))}
	b.cfg = config{
		url:           "amqp://guest:guest@localhost:5672",
		prefetchCount: 1}
	return &b
}

func (b *Broker) Send(dst string, m ezbus.Message) error {
	err := publish(b.channel, m, dst, "")
	if err != nil {
		return fmt.Errorf("Send: %s", err)
	}
	return err
}

func (b *Broker) Publish(m ezbus.Message) error {
	msgName := m.Headers[ezbus.MessageName]
	err := publish(b.channel, m, "", msgName)
	if err != nil {
		return fmt.Errorf("Publish: %s", err)
	}
	return err
}

func (b *Broker) Start(handle ezbus.MessageHandler) error {
	cn, err := amqp.Dial(b.cfg.url)

	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	b.conn = cn
	b.channel, err = b.conn.Channel()

	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	if b.Endpoint() == "" {
		return nil
	}

	err = b.channel.Qos(b.cfg.prefetchCount, 0, false)

	if err != nil {
		return fmt.Errorf("Qos: %s", err)
	}

	queue, err := declareQueue(b.channel, b.queueName)

	if err != nil {
		return fmt.Errorf("Declare Queue : %s", err)
	}

	log.Printf("Queue declared. (%q %d messages, %d consumers)", queue.Name, queue.Messages, queue.Consumers)

	err = declareExchange(b.channel, b.queueName)

	if err != nil {
		return fmt.Errorf("Declare Exchange : %s", err)
	}

	log.Printf("Exchange declared. (%q)", b.queueName)

	msgs, err := b.channel.Consume(queue.Name, "", false, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	for d := range msgs {
		headers := extractHeaders(d.Headers)
		m := ezbus.Message{Headers: headers, Body: d.Body}
		handle(m)
		d.Ack(false)
	}

	<-b.done
	log.Printf("Hm")
	return nil
}

func (b *Broker) Stop() error {
	b.done <- struct{}{}

	err := b.channel.Close()
	if err != nil {
		return fmt.Errorf("Channel Close: %s", err)
	}
	err = b.conn.Close()
	if err != nil {
		return fmt.Errorf("Connection Close: %s", err)
	}

	return nil
}

//Endpoint returns name of the queue
func (b *Broker) Endpoint() string {
	return b.queueName
}

//Subscribe to messages from specific endpoint
func (b *Broker) Subscribe(endpoint string, messageName string) error {
	log.Printf("Subscribing to message '%s' from endpoint '%s'", messageName, endpoint)
	return queueBind(b.channel, b.Endpoint(), messageName, endpoint)
}

//Configure RabbitMQ.
//url to broker
//prefetchCount
func (b *Broker) Configure(url string, prefetchCount int) {
	b.cfg.url = url
	b.cfg.prefetchCount = prefetchCount
}

func extractHeaders(h amqp.Table) map[string]string {
	headers := make(map[string]string)
	for k, v := range h {
		headers[k] = v.(string)
	}
	return headers
}
