package ezbus

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
)

// Bus for publish, send and receive messages
type Bus struct {
	broker   Broker
	router   Router
	done     chan (struct{})
	messages chan (Message)
}

// NewBus creates a bus instance for sending and receiving messages.
func NewBus(b Broker, r Router) *Bus {
	bus := Bus{b, r, make(chan struct{}), make(chan Message)}
	return &bus
}

// NewSendOnlyBus creates a bus instance for sending messages.
func NewSendOnlyBus(b Broker) *Bus {
	return &Bus{broker: b}
}

//Go starts the bus and listens to incoming messages.
func (b *Bus) Go() {
	go b.handle()

	err := b.broker.Start(b.messages)

	if err != nil {
		log.Panicln("Failed to start broker: ", err)
	}
}

func (b *Bus) Stop() {
	go func() {
		b.done <- struct{}{}
	}()
}

// Send message to destination
func (b *Bus) Send(dst string, msg interface{}) error {
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	t := reflect.TypeOf(msg)

	return b.broker.Send(dst, NewMessage(getHeaders(t), json))
}

func (b *Bus) handle() {

	log.Println("starting handle")

	for m := range b.messages {
		n := m.Headers[MessageName]

		err := retry(func() {
			b.router.handle(n, m)
		}, 5)

		if err != nil {
			eq := fmt.Sprintf("%s.error", b.broker.QueueName())
			log.Println("Failed to handle message. Putting on error queue: ", eq)
			b.broker.Send(eq, m)
		}
	}

	log.Println("no more handle")
}

func recoverHandle(m Message) {
	if err := recover(); err != nil {
		log.Printf("Failed to handle '%s': %s", m.Headers[MessageName], err)
	}
}

func getHeaders(msgType reflect.Type) map[string]string {
	h := make(map[string]string)
	h[MessageName] = msgType.Name()
	h[MessageFullname] = msgType.String()
	h[TimeSent] = time.Now().Format("2006-01-02 15:04:05.000000")

	hostName, err := os.Hostname()
	if err == nil {
		h[SendingHost] = hostName
	}

	return h
}
