package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/logger"
	"github.com/zapote/go-ezbus/rabbitmq"
)

type greeting struct {
	Text string `json:"text"`
}

type greeting2 struct {
	Text string `json:"text"`
}

func main() {

	logger.SetLevel(logger.DebugLevel)

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
	receiver.Go()

	//publish messsage
	publisher.Publish(greeting{"hello ezbus"})

	forever := make(chan (struct{}))
	<-forever
}

func handler(m ezbus.Message) error {
	var g greeting
	json.Unmarshal(m.Body, &g)
	log.Println(g.Text)
	return fmt.Errorf("Did not work: %d", 1337)
}
