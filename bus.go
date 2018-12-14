package ezbus

import (
	"encoding/json"
	"os"
	"reflect"
	"time"
)

// Bus for publish, send and receive messages
type Bus struct {
	endpoint string
	broker   Broker
	router   Router
}

// NewBus creates a bus instance for sending and receiving messages.
func NewBus(b Broker, r Router) *Bus {
	bus := Bus{broker: b, router: r}
	return &bus
}

// NewSendOnlyBus creates a bus instance for sending messages.
func NewSendOnlyBus(b Broker) *Bus {
	bus := Bus{broker: b}
	return &bus
}

func (b *Bus) listen() {
	c := make(chan Message)
	err := b.broker.Start(c)

	if err == nil {
		go func() {
			m := <-c
			n := m.Headers["message-name"]
			b.router.handle(n, m)
		}()
	}
}

// Send message to destination
func (b *Bus) Send(dest string, msg interface{}) error {
	n := reflect.TypeOf(msg).Name()

	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.broker.Send(dest, NewMessage(getHeaders(n), json))
}

func getHeaders(messageName string) map[string]string {
	h := make(map[string]string)
	h["message-name"] = messageName
	h["time-sent"] = time.Now().Format("2006-01-02 15:04:05.000000")
	hostName, err := os.Hostname()
	if err == nil {
		h["sending-host"] = hostName
	}
	return h
}
