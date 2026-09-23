package ezbus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/zapote/go-ezbus/headers"
	"github.com/zapote/go-ezbus/logger"
)

type subscription struct {
	endpoint    string
	messageName string
}

type subscriptions []subscription

// Bus for publishing, sending and receiving messages
type Bus interface {
	StarterStopper
	Sender
	Publisher
	Subscriber
}

// Sender sends a message to a destination. SendContext carries ctx
// along with the message; Send is SendContext with context.Background().
type Sender interface {
	Send(dst string, msg interface{}) error
	SendContext(ctx context.Context, dst string, msg interface{}) error
}

// Publisher publishes a message to subscribers. PublishContext carries
// ctx along with the message; Publish is PublishContext with
// context.Background().
type Publisher interface {
	Publish(msg interface{}) error
	PublishContext(ctx context.Context, msg interface{}) error
}

// Subscriber interface
type Subscriber interface {
	Subscribe(endpoint string)
	SubscribeMessage(endpoint string, messageName string)
}

// StarterStopper interface
type StarterStopper interface {
	Go() error
	Stop() error
}

type bus struct {
	broker      Broker
	router      Router
	subscribers []subscription
}

// NewBus creates a bus instance for sending and receiving messages.
func NewBus(b Broker, r Router) Bus {
	bus := bus{
		broker:      b,
		router:      r,
		subscribers: make([]subscription, 0),
	}

	return &bus
}

// Go starts the bus and listens to incoming messages.
func (b *bus) Go() error {
	err := b.broker.Start(b.handle)
	if err != nil {
		return err
	}
	for _, s := range b.subscribers {
		err = b.broker.Subscribe(s.endpoint, s.messageName)
		if err != nil {
			return err
		}
	}

	logger.Info("Bus is on the Go!")
	return nil
}

// Stop the bus and any incoming messages.
func (b *bus) Stop() error {
	logger.Info("Bus stopped.")
	return b.broker.Stop()
}

// Send message to destination.
func (b *bus) Send(dst string, msg interface{}) error {
	return b.SendContext(context.Background(), dst, msg)
}

// SendContext sends a message to given destination, carrying ctx along.
func (b *bus) SendContext(ctx context.Context, dst string, msg interface{}) error {
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	t := reflect.TypeOf(msg)
	h := b.getHeaders(t, dst)

	ctx, span := startPublish(ctx, dst, h)
	defer span.End()

	start := time.Now()
	err = b.broker.Send(dst, NewMessage(h, json).WithContext(ctx))
	recordSend(ctx, operationSend, dst, start, err)
	if err != nil {
		failSpan(span, err)
	}

	return err
}

// Publish message to subscribers
func (b *bus) Publish(msg interface{}) error {
	return b.PublishContext(context.Background(), msg)
}

// PublishContext publishes a message to subscribers, carrying ctx along.
func (b *bus) PublishContext(ctx context.Context, msg interface{}) error {
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	t := reflect.TypeOf(msg)
	h := b.getHeaders(t, "")

	ctx, span := startPublish(ctx, t.Name(), h)
	defer span.End()

	start := time.Now()
	err = b.broker.Publish(NewMessage(h, json).WithContext(ctx))
	recordSend(ctx, operationPublish, t.Name(), start, err)
	if err != nil {
		failSpan(span, err)
	}

	return err
}

// SubscribeMessage to a specific message from a publisher. Provide endpoint (queue) and name of the message to subscribe to.
func (b *bus) SubscribeMessage(endpoint string, messageName string) {
	logger.Infof("Subscribing to message '%s' from endpoint '%s'", messageName, endpoint)
	b.subscribers = append(b.subscribers, subscription{endpoint, messageName})
}

// Subscribe to all messages from a publisher. Provide endpoint (queue).
func (b *bus) Subscribe(endpoint string) {
	logger.Infof("Subscribing to all messages from endpoint '%s'", endpoint)
	b.subscribers = append(b.subscribers, subscription{endpoint, ""})
}

func (b *bus) handle(m Message) (err error) {
	n := m.Headers[headers.MessageName]
	queue := b.broker.Endpoint()
	start := time.Now()

	ctx, span := startProcess(queue, n, m.Headers)
	defer span.End()
	m = m.WithContext(ctx)

	attempts, err := receive(ctx, n, func() error {
		return b.router.Receive(n, m)
	}, 5)

	if err == nil {
		recordProcess(ctx, queue, n, start, attempts, outcomeOK)
		return nil
	}

	if IsHandlerNotFoundErr(err) {
		logger.Debugf("Message will be discarded: %s", err.Error())
		recordProcess(ctx, queue, n, start, attempts, outcomeDiscarded)
		return nil
	}

	failAfterAttempts(span, err)
	recordProcess(ctx, queue, n, start, attempts, outcomeErrorQueue)

	eq := fmt.Sprintf("%s-error", queue)
	m.Headers[headers.Error] = err.Error()
	markErrorQueued(ctx, m.Headers)

	logger.Errorf("Failed to handle message. Putting on error queue: %s\n", eq)

	return b.broker.Send(eq, m)
}

func (b *bus) getHeaders(msgType reflect.Type, dst string) map[string]string {
	h := make(map[string]string)
	h[headers.MessageName] = msgType.Name()
	h[headers.MessageFullname] = msgType.String()
	h[headers.TimeSent] = time.Now().Format("2006-01-02 15:04:05.000000")

	if dst != "" {
		h[headers.Destination] = dst
	}

	n, err := os.Hostname()
	if err == nil {
		h[headers.SendingHost] = n
	} else {
		logger.Warnf("Failed to get hostname: %v", err)
	}

	return h
}
