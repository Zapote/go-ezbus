package ezbus

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"time"
)

// Bus for publish, send and receive messages
type Bus struct {
	broker  Broker
	router  Router
	done    chan (struct{})
	forward chan (Message)
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

func (b *Bus) Start() {
	err := b.broker.Start(b.forward)

	if err != nil {
		log.Panicf("Failed to start broker")
		panic(err)
	}

	go b.handle()
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
	for m := range b.forward {
		n := m.Headers[MessageName]

		retry(func() {
			b.router.handle(n, m)
		}, 5)
	}
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
