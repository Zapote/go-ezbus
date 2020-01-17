package main

import (
	"encoding/json"
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

type greeting struct {
	Text string `json:"text"`
}

type greeting2 struct {
	Text string `json:"text"`
}

func main() {
	//setup publisher
	bp := rabbitmq.NewBroker("sample-publisher")
	rp := ezbus.NewRouter()
	publisher := ezbus.NewBus(bp, rp)
	publisher.Go()

	//setup receiver
	br := rabbitmq.NewBroker("sample-receiver")
	rr := ezbus.NewRouter()
	rr.Handle("greeting", handler)
	receiver := ezbus.NewBus(br, rr)
	receiver.SubscribeMessage("sample-publisher", "greeting")
	//receiver.Subscribe("sample-publisher")
	receiver.Go()

	//publish messsage
	publisher.Publish(greeting{"hello ezbus"})
	publisher.Publish(greeting2{"hello ezbus"})

	forever := make(chan (struct{}))
	<-forever
}

func handler(m ezbus.Message) error {
	var g greeting
	json.Unmarshal(m.Body, &g)
	log.Println(g.Text)
	return nil
}
