package ezbus

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
)

// Bus for publishing, sending and receiving messages
type Bus struct {
	broker      Broker
	router      Router
	done        chan (struct{})
	messages    chan (Message)
	subscribers []subscription
}

// NewBus creates a bus instance for sending and receiving messages.
func NewBus(b Broker, r Router) *Bus {
	bus := Bus{
		b,
		r,
		make(chan struct{}),
		make(chan Message),
		make([]subscription, 0)}

	return &bus
}

// NewSendOnlyBus creates a bus instance for sending messages.
func NewSendOnlyBus(b Broker) (*Bus, error) {
	bus := Bus{broker: b}
	err := bus.broker.Start(bus.messages)
	return &bus, err
}

//Go starts the bus and listens to incoming messages.
func (b *Bus) Go() {
	go b.handle()
	go b.startBroker()

	log.Println("Bus is on the Go!")

	for _, s := range b.subscribers {
		b.broker.Subscribe(s.endpoint, s.messageName)
	}

	<-b.done
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

//Subscribe to a publisher. Provide endpoint (queue) and name of the message to subscribe to.
func (b *Bus) Subscribe(endpoint string, messageName string) {
	b.subscribers = append(b.subscribers, subscription{endpoint, messageName})
}

func (b *Bus) handle() {
	for m := range b.messages {
		n := m.Headers[MessageName]

		err := retry(func() {
			b.router.handle(n, m)
		}, 5)

		if err != nil {
			eq := fmt.Sprintf("%s.error", b.broker.Endpoint())
			log.Println("Failed to handle message. Putting on error queue: ", eq)
			b.broker.Send(eq, m)
		}
	}
}

func (b *Bus) startBroker() {
	err := b.broker.Start(b.messages)

	if err != nil {
		log.Panicln("Failed to start broker: ", err)
	}
}

func recoverHandle(m Message) {
	if err := recover(); err != nil {
		log.Printf("Failed to handle '%s': %s", m.Headers[MessageName], err)
	}
}

func getHeaders(msgType reflect.Type, dst string) map[string]string {
	h := make(map[string]string)
	h[MessageName] = msgType.Name()
	h[MessageFullname] = msgType.String()
	h[TimeSent] = time.Now().Format("2006-01-02 15:04:05.000000")

	if dst != "" {
		h[Destination] = dst
	}

	hostName, err := os.Hostname()
	if err == nil {
		h[SendingHost] = hostName
	}

	return h
}
