package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/headers"
	"github.com/zapote/go-ezbus/logger"
)

// Broker RabbitMQ implementation of ezbus.broker interface.
type Broker struct {
	queueName   string
	consumerTag string
	handler     ezbus.MessageHandler
	cfg         *config
	sendOnly    bool
	draining    atomic.Bool

	// mu guards what a reconnect replaces while Shutdown reads it.
	mu             sync.Mutex
	socket         net.Conn
	conn           *amqp.Connection
	sendChannel    *amqp.Channel
	receiveChannel *amqp.Channel
	drained        chan struct{}
}

// NewBroker creates a RabbitMQ broker for queue. An empty queue gives a
// broker that only sends. Options override the defaults: URL
// amqp://guest:guest@localhost:5672, prefetch count 100, queue name
// delimiter "-" and drain timeout 20 seconds.
func NewBroker(queue string, opts ...Option) *Broker {
	cfg := &config{
		url:                "amqp://guest:guest@localhost:5672",
		prefetchCount:      100,
		queueNameDelimiter: "-",
		drainTimeout:       20 * time.Second,
		reconnectAttempts:  60,
		reconnectDelay:     5 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Broker{
		queueName:   queue,
		consumerTag: consumerTag(queue),
		sendOnly:    queue == "",
		cfg:         cfg,
	}
}

// consumerTag names the consumer after the queue and the host, so that it
// can be cancelled and told apart from the other consumers of the queue.
func consumerTag(queue string) string {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}
	return fmt.Sprintf("%s@%s", queue, host)
}

// Send sends a message to given destination
func (b *Broker) Send(dst string, m ezbus.Message) error {

	err := publish(b.sendChannel, m, dst, "")
	if err != nil {
		return fmt.Errorf("Send: %s", err)
	}
	return err
}

// Publish publishes message on exhange
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

// Start starts the RabbitMQ broker and declars queue, and exchange.
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

// Stop shuts the broker down within the drain timeout.
func (b *Broker) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), b.cfg.drainTimeout)
	defer cancel()
	return b.Shutdown(ctx)
}

// Shutdown stops taking messages, lets the handler that is running finish
// and ack, puts the deliveries it has buffered back on the queue and closes
// the connection. When ctx ends first the connection is cut without
// waiting for the server, and the message in the handler is delivered
// again.
func (b *Broker) Shutdown(ctx context.Context) error {
	if !b.draining.CompareAndSwap(false, true) {
		return nil
	}
	stopCutting := context.AfterFunc(ctx, b.cutConnection)

	b.mu.Lock()
	receive, drained := b.receiveChannel, b.drained
	b.mu.Unlock()

	var errs []error
	if drained != nil {
		if err := receive.Cancel(b.consumerTag, false); open(err) {
			errs = append(errs, fmt.Errorf("Cancel consumer: %w", err))
		}

		select {
		case <-drained:
		case <-ctx.Done():
		}
	}

	errs = append(errs, b.closeConnection())
	if stopCutting() {
		return errors.Join(errs...)
	}
	return fmt.Errorf("Shutdown cut the connection: %w", ctx.Err())
}

// cutConnection closes the socket under the connection, which ends every
// call that is still waiting for the server.
func (b *Broker) cutConnection() {
	b.mu.Lock()
	socket := b.socket
	b.mu.Unlock()

	if socket != nil {
		socket.Close()
	}
}

// closeConnection closes the send channel last: a handler that is about to
// finish may still send or publish.
func (b *Broker) closeConnection() error {
	b.mu.Lock()
	conn, send, receive := b.conn, b.sendChannel, b.receiveChannel
	b.mu.Unlock()

	var errs []error
	if receive != nil {
		if err := receive.Close(); open(err) {
			errs = append(errs, fmt.Errorf("Receive channel Close: %w", err))
		}
	}
	if send != nil {
		if err := send.Close(); open(err) {
			errs = append(errs, fmt.Errorf("Send channel Close: %w", err))
		}
	}
	if conn != nil {
		if err := conn.Close(); open(err) {
			errs = append(errs, fmt.Errorf("Connection Close: %w", err))
		}
	}
	return errors.Join(errs...)
}

// open tells an error apart from the connection already being closed,
// which is what shutting down is after.
func open(err error) bool {
	return err != nil && !errors.Is(err, amqp.ErrClosed)
}

// Endpoint returns name of the queue
func (b *Broker) Endpoint() string {
	return b.queueName
}

// Subscribe to messages from specific endpoint
func (b *Broker) Subscribe(endpoint string, messageName string) error {
	if messageName == "" {
		messageName = "#"
	}
	return queueBind(b.receiveChannel, b.Endpoint(), messageName, endpoint)
}

func (b *Broker) connect() error {
	var socket net.Conn
	cn, err := amqp.DialConfig(b.cfg.url, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Dial: func(network, addr string) (net.Conn, error) {
			conn, err := amqp.DefaultDial(30*time.Second)(network, addr)
			socket = conn
			return conn, err
		},
	})
	if err != nil {
		return fmt.Errorf("amqp.Dial: %s", err.Error())
	}

	notifyClose := make(chan *amqp.Error)
	cn.NotifyClose(notifyClose)

	go func() {
		closeErr := <-notifyClose
		if closeErr == nil {
			return
		}

		logger.Warnf("Connection closed: %s", closeErr.Error())
		b.reconnect()
	}()

	send, err := cn.Channel()
	if err != nil {
		cn.Close()
		return fmt.Errorf("Send channel: %s", err)
	}

	var receive *amqp.Channel
	if !b.sendOnly {
		receive, err = cn.Channel()
		if err != nil {
			cn.Close()
			return fmt.Errorf("Receive channel: %s", err)
		}

		err = receive.Qos(b.cfg.prefetchCount, 0, false)
		if err != nil {
			cn.Close()
			return fmt.Errorf("Qos: %s", err)
		}
	}

	b.mu.Lock()
	b.socket, b.conn, b.sendChannel, b.receiveChannel = socket, cn, send, receive
	b.mu.Unlock()

	return nil
}

func (b *Broker) reconnect() {
	attempts := b.cfg.reconnectAttempts
	for i := 0; i < attempts; i++ {
		if b.draining.Load() {
			return
		}
		logger.Infof("Reconnecting... attempt %d of %d", i+1, attempts)

		if err := b.connect(); err != nil {
			logger.Warnf("Failed to connect: %s", err.Error())
			time.Sleep(b.cfg.reconnectDelay)
			continue
		}
		logger.Infof("Reconnect succeeded")

		if b.draining.Load() {
			b.closeConnection()
			return
		}

		if err := b.consume(); err != nil {
			logger.Warnf("Failed to consume: %s", err.Error())
			continue
		}
		logger.Infof("Consume succeeded")

		return
	}

	if b.draining.Load() {
		return
	}
	panic("Unable to reconnect to broker. Giving up after many attempts.")
}

func (b *Broker) consume() error {
	if b.sendOnly {
		return nil
	}

	b.mu.Lock()
	receive := b.receiveChannel
	b.mu.Unlock()

	deliveries, err := receive.Consume(b.queueName, b.consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	drained := make(chan struct{})
	b.mu.Lock()
	b.drained = drained
	b.mu.Unlock()

	go func() {
		defer close(drained)
		for d := range deliveries {
			if b.draining.Load() {
				b.putBack(d)
				continue
			}
			b.process(d)
		}
	}()

	return nil
}

// process acks the delivery when the handler succeeded. A handler that
// failed has neither handled the message nor moved it to the error queue,
// so the message goes back on the queue.
func (b *Broker) process(d amqp.Delivery) {
	m := ezbus.NewMessage(extractHeaders(d.Headers), d.Body)
	name := m.Headers[headers.MessageName]

	if err := b.handler(m); err != nil {
		logger.ErrorContext(m.Context(), "message put back on the queue", "message", name, "queue", b.queueName, "err", err)
		if err := d.Nack(false, true); err != nil {
			logger.ErrorContext(m.Context(), "nack failed", "message", name, "queue", b.queueName, "err", err)
		}
		return
	}

	if err := d.Ack(false); err != nil {
		logger.ErrorContext(m.Context(), "ack failed, the message will be delivered again", "message", name, "queue", b.queueName, "err", err)
	}
}

// putBack returns a delivery that was buffered when the shutdown began.
func (b *Broker) putBack(d amqp.Delivery) {
	if err := d.Nack(false, true); open(err) {
		logger.Warnf("Failed to put a buffered message back on %q: %s", b.queueName, err)
	}
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
	headers := make(map[string]string, len(h))
	for k, v := range h {
		switch t := v.(type) {
		case string:
			headers[k] = t
		case []byte:
			headers[k] = string(t)
		case nil:
			headers[k] = ""
		default:
			headers[k] = fmt.Sprint(t)
		}
	}
	return headers
}
