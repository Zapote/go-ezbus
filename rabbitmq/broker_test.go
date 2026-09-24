package rabbitmq

/*
	Needs a running RabbmitMQ on localhost:5672
*/
import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"testing"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/headers"
	"gotest.tools/v3/assert"
)

func TestSend(t *testing.T) {
	v := &validator{}
	v.queue = "send-validator"
	v.start()

	b := NewBroker("")
	b.Start(func(m ezbus.Message) error {
		return nil
	})
	h := make(map[string]string)
	h[headers.MessageName] = "validation-message"
	m := ezbus.NewMessage(h, []byte("message-body-sent"))

	err := b.Send("send-validator", m)
	if err != nil {
		t.Errorf("Failed to send message: %s", err.Error())
	}

	v.waitOne()

	assert.Equal(t, "message-body-sent", string(v.m.Body))
}

func TestPublish(t *testing.T) {
	b := NewBroker("test-publisher")
	b.Start(func(m ezbus.Message) error {
		return nil
	})
	v := &validator{}
	v.queue = "publish-validator"
	v.start()
	v.b.Subscribe("test-publisher", "validation-message")

	h := make(map[string]string)
	h[headers.MessageName] = "validation-message"
	m := ezbus.NewMessage(h, []byte("message-body-published"))

	err := b.Publish(m)
	if err != nil {
		t.Errorf("Failed to publish message: %s", err.Error())
	}

	v.waitOne()

	assert.Equal(t, "message-body-published", string(v.m.Body))
}

type validator struct {
	b     *Broker
	m     ezbus.Message
	wg    sync.WaitGroup
	queue string
}

// start expects exactly one message. The queue is purged first: a message
// left behind by an earlier run would otherwise be delivered before the
// test has counted it, and the wait group would go negative.
func (v *validator) start() {
	v.wg.Add(1)
	purge(v.queue)
	v.b = NewBroker(v.queue)
	err := v.b.Start(v.handler())
	if err != nil {
		panic(err)
	}
}

func (v *validator) waitOne() {
	v.wg.Wait()
}

func purge(queue string) {
	cn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(fmt.Sprintf("Dial: %s", err))
	}
	defer cn.Close()
	ch, err := cn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Channel: %s", err))
	}
	defer ch.Close()
	if _, err := declareQueue(ch, queue); err != nil {
		panic(fmt.Sprintf("DeclareQueue: %s", err))
	}
	if _, err := ch.QueuePurge(queue, false); err != nil {
		panic(fmt.Sprintf("QueuePurge: %s", err))
	}
}

func (v *validator) handler() ezbus.MessageHandler {
	return func(m ezbus.Message) error {
		v.m = m
		log.Println(string(m.Body))
		v.wg.Done()
		return nil
	}
}
