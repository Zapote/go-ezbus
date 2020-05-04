package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/headers"
)

//Broker RabbitMQ implementation of ezbus.broker interface.
type Broker struct {
	queueName      string
	conn           *amqp.Connection
	sendChannel    *amqp.Channel
	receiveChannel *amqp.Channel
	cfg            *config
}

//NewBroker creates a RabbitMQ broker instance
//Default url amqp://guest:guest@localhost:5672
//Default prefetchCount 100
func NewBroker(q ...string) *Broker {
	var queue string
	if len(q) > 0 {
		queue = q[0]
	}

	b := Broker{queueName: queue}
	b.cfg = &config{
		url:                "amqp://guest:guest@localhost:5672",
		prefetchCount:      100,
		queueNameDelimiter: "-",
	}
	return &b
}

//Send sends a message to given destination
func (b *Broker) Send(dst string, m ezbus.Message) error {
	err := publish(b.sendChannel, m, dst, "")
	if err != nil {
		return fmt.Errorf("Send: %s", err)
	}
	return err
}

//Publish publishes message on exhange
func (b *Broker) Publish(m ezbus.Message) error {
	key := m.Headers[headers.MessageName]
	err := publish(b.sendChannel, m, key, b.queueName)
	if err != nil {
		return fmt.Errorf("Publish: %s", err)
	}
	return err
}

//Start starts the RabbitMQ broker and declars queue, and exchange.
func (b *Broker) Start(handle ezbus.MessageHandler) error {
	cn, err := amqp.Dial(b.cfg.url)

	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	b.conn = cn
	b.sendChannel, err = b.conn.Channel()
	if err != nil {
		return fmt.Errorf("Send channel: %s", err)
	}

	b.receiveChannel, err = b.conn.Channel()
	if err != nil {
		return fmt.Errorf("Receive channel: %s", err)
	}

	if b.Endpoint() == "" {
		return nil
	}

	err = b.receiveChannel.Qos(b.cfg.prefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("Qos: %s", err)
	}

	queue, err := declareQueue(b.receiveChannel, b.queueName)
	if err != nil {
		return fmt.Errorf("Declare Queue : %s", err)
	}
	log.Printf("Queue declared. (%q %d messages, %d consumers)", queue.Name, queue.Messages, queue.Consumers)

	queueErr, err := declareQueue(b.receiveChannel, fmt.Sprintf("%s%serror", b.queueName, b.cfg.queueNameDelimiter))
	if err != nil {
		return fmt.Errorf("Declare Error Queue : %s", err)
	}
	log.Printf("Queue declared. (%q %d messages, %d consumers)", queueErr.Name, queue.Messages, queue.Consumers)

	err = declareExchange(b.receiveChannel, b.queueName)
	if err != nil {
		return fmt.Errorf("Declare Exchange : %s", err)
	}
	log.Printf("Exchange declared. (%q)", b.queueName)

	msgs, err := b.receiveChannel.Consume(queue.Name, "", false, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}
	go func() {
		for d := range msgs {
			headers := extractHeaders(d.Headers)
			m := ezbus.Message{Headers: headers, Body: d.Body}
			handle(m)
			b.receiveChannel.Ack(d.DeliveryTag, false)
		}
	}()
	log.Print("RabbitMQ broker started")
	return nil
}

//Stop stops the RabbitMQ broker
func (b *Broker) Stop() error {
	err := b.sendChannel.Close()
	if err != nil {
		return fmt.Errorf("Send channel Close: %s", err)
	}
	err = b.receiveChannel.Close()
	if err != nil {
		return fmt.Errorf("Receive channel Close: %s", err)
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
	if messageName == "" {
		messageName = "#"
	}
	return queueBind(b.receiveChannel, b.Endpoint(), messageName, endpoint)
}

//Configure RabbitMQ.
func (b *Broker) Configure() Configurer {
	return b.cfg
}

func extractHeaders(h amqp.Table) map[string]string {
	headers := make(map[string]string)
	for k, v := range h {
		headers[k] = v.(string)
	}
	return headers
}
