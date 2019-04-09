package ezbus

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/zapote/go-ezbus/headers"
)

type subscription struct {
	endpoint    string
	messageName string
}

type subscriptions []subscription

// Bus for publishing, sending and receiving messages
type Bus struct {
	broker      Broker
	router      Router
	subscribers []subscription
}

// NewBus creates a bus instance for sending and receiving messages.
func NewBus(b Broker, r Router) *Bus {
	bus := Bus{
		broker:      b,
		router:      r,
		subscribers: make([]subscription, 0)}

	return &bus
}

// NewSendOnlyBus creates a bus instance for sending messages.
func NewSendOnlyBus(b Broker) (*Bus, error) {
	bus := Bus{broker: b}
	err := bus.broker.Start(bus.handle)
	return &bus, err
}

//Go starts the bus and listens to incoming messages.
func (b *Bus) Go() {
	b.startBroker()
	for _, s := range b.subscribers {
		b.broker.Subscribe(s.endpoint, s.messageName)
	}
	log.Println("Bus is on the Go!")
}

//Stop the bus and any incoming messages.
func (b *Bus) Stop() {
	defer log.Println("Bus stopped.")
	b.broker.Stop()
}

// Send message to destination.
func (b *Bus) Send(dst string, msg interface{}) error {
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	t := reflect.TypeOf(msg)

	return b.broker.Send(dst, NewMessage(getHeaders(t, dst), json))
}

//Publish message to subscribers
func (b *Bus) Publish(msg interface{}) error {
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	t := reflect.TypeOf(msg)

	return b.broker.Publish(NewMessage(getHeaders(t, ""), json))
}

//SubscribeMessage to a specific message from a publisher. Provide endpoint (queue) and name of the message to subscribe to.
func (b *Bus) SubscribeMessage(endpoint string, messageName string) {
	log.Printf("Subscribing to message '%s' from endpoint '%s'", messageName, endpoint)
	b.subscribers = append(b.subscribers, subscription{endpoint, messageName})
}

//Subscribe to all messages from a publisher. Provide endpoint (queue).
func (b *Bus) Subscribe(endpoint string) {
	log.Printf("Subscribing to all messages from endpoint '%s'", endpoint)
	b.subscribers = append(b.subscribers, subscription{endpoint, ""})
}

func (b *Bus) handle(m Message) {
	n := m.Headers[headers.MessageName]
	err := retry(func() {
		b.router.handle(n, m)
	}, 5)

	if err == nil {
		return
	}

	eq := fmt.Sprintf("%s.error", b.broker.Endpoint())
	log.Println("Failed to handle message. Putting on error queue: ", eq)
	err = b.broker.Send(eq, m)
	if err != nil {
		log.Println("Failed to put message on error queue: ", eq)
	}
}

func (b *Bus) startBroker() {
	err := b.broker.Start(b.handle)

	if err != nil {
		log.Panicln("Failed to start broker: ", err)
	}
}

func recoverHandle(m Message) {
	if err := recover(); err != nil {
		log.Printf("Failed to handle '%s': %s", m.Headers[headers.MessageName], err)
	}
}

func getHeaders(msgType reflect.Type, dst string) map[string]string {
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
	}
	if err != nil {
		log.Printf("Failed to get hostname: %v", err)
	}

	return h
}
