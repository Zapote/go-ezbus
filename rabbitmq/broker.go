package rabbitmq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/headers"
	"github.com/zapote/go-ezbus/logger"
)

//Broker RabbitMQ implementation of ezbus.broker interface.
type Broker struct {
	queueName      string
	handler        ezbus.MessageHandler
	cfg            *config
	sendOnly       bool
	conn           *amqp.Connection
	sendChannel    *amqp.Channel
	receiveChannel *amqp.Channel
}

//NewBroker creates a RabbitMQ broker instance
//Default url amqp://guest:guest@localhost:5672
//Default prefetchCount 100
func NewBroker(q ...string) *Broker {
	var queue string

	if len(q) > 0 {
		queue = q[0]
	}

	b := Broker{queueName: queue, sendOnly: queue == ""}
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
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("Publish: %s", err)
	}

	err = publish(ch, m, key, b.queueName)
	defer ch.Close()
	if err != nil {
		return fmt.Errorf("Publish: %s", err)
	}
	return err
}

//Start starts the RabbitMQ broker and declars queue, and exchange.
func (b *Broker) Start(h ezbus.MessageHandler) error {
	b.handler = h

	if err := b.connect(); err != nil {
		return err
	}

	if err := b.declareQueues(); err != nil {
		return err
	}

	if err := b.consume(); err != nil {
		return err
	}

	logger.Info("RabbitMQ broker started")
	return nil
}

//Stop stops the RabbitMQ broker
func (b *Broker) Stop() error {
	err := b.sendChannel.Close()
	if err != nil {
		return fmt.Errorf("Send channel Close: %s", err)
	}

	if b.receiveChannel != nil {
		err = b.receiveChannel.Close()
		if err != nil {
			return fmt.Errorf("Receive channel Close: %s", err)
		}
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

func (b *Broker) connect() error {
	cn, err := amqp.Dial(b.cfg.url)
	if err != nil {
		return fmt.Errorf("amqp.Dial: %s", err.Error())
	}

	b.conn = cn
	notifyClose := make(chan *amqp.Error)
	b.conn.NotifyClose(notifyClose)

	go func() {
		closeErr := <-notifyClose
		if closeErr == nil {
			return
		}

		logger.Warnf("Connection closed: %s", closeErr.Error())
		attempts := 60
		for i := 0; i < attempts; i++ {
			logger.Infof("Reconnecting... attempt %d of %d", i+1, attempts)

			if err := b.connect(); err != nil {
				logger.Warnf("Failed to connect: %s", err.Error())
				time.Sleep(time.Second * 5)
				continue
			}
			logger.Infof("Reconnect succeeded")

			if err = b.consume(); err != nil {
				logger.Warnf("Failed to consume: %s", err.Error())
				continue
			}
			logger.Infof("Consume succeeded")

			return
		}

		panic("Unable to reconnect to broker. Giving up after many attempts.")
	}()

	//create sending channel
	b.sendChannel, err = b.conn.Channel()
	if err != nil {
		return fmt.Errorf("Send channel: %s", err)
	}

	if b.sendOnly {
		return nil
	}

	//create receiving channel
	b.receiveChannel, err = b.conn.Channel()
	if err != nil {
		return fmt.Errorf("Receive channel: %s", err)
	}

	err = b.receiveChannel.Qos(b.cfg.prefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("Qos: %s", err)
	}

	return nil
}

func (b *Broker) consume() error {
	if b.sendOnly {
		return nil
	}

	deliveries, err := b.receiveChannel.Consume(b.queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	go func() {
		for d := range deliveries {
			headers := extractHeaders(d.Headers)
			m := ezbus.Message{Headers: headers, Body: d.Body}
			b.handler(m)
			b.receiveChannel.Ack(d.DeliveryTag, false)
		}
	}()

	return nil
}

func (b *Broker) declareQueues() error {
	if b.sendOnly {
		return nil
	}
	//declare queues
	queue, err := declareQueue(b.receiveChannel, b.queueName)
	if err != nil {
		return fmt.Errorf("Declare Queue : %s", err)
	}
	logger.Infof("Queue declared. (%q %d messages, %d consumers)", queue.Name, queue.Messages, queue.Consumers)

	queueErr, err := declareQueue(b.receiveChannel, fmt.Sprintf("%s%serror", b.queueName, b.cfg.queueNameDelimiter))
	if err != nil {
		return fmt.Errorf("Declare Error Queue : %s", err)
	}
	logger.Infof("Queue declared. (%q %d messages)", queueErr.Name, queueErr.Messages)

	//declare exchange
	err = declareExchange(b.receiveChannel, b.queueName)
	if err != nil {
		return fmt.Errorf("Declare Exchange : %s", err)
	}
	logger.Infof("Exchange declared. (%q)", b.queueName)

	return nil
}

func extractHeaders(h amqp.Table) map[string]string {
	headers := make(map[string]string)
	for k, v := range h {
		headers[k] = v.(string)
	}
	return headers
}
