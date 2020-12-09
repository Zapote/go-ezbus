package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/logger"
	"github.com/zapote/go-ezbus/rabbitmq"
)

type greeting struct {
	Text string `json:"text"`
}

func main() {
	logger.SetLevel(logger.DebugLevel)

	//setup publisher
	bp := rabbitmq.NewBroker("sample-publisher")
	rp := ezbus.NewRouter()
	publisher := ezbus.NewBus(bp, rp)
	err := publisher.Go()
	if err != nil {
		log.Fatalf("Start publisher: %s", err.Error())
	}

	//setup receiver
	br := rabbitmq.NewBroker("sample-receiver")
	rr := ezbus.NewRouter()
	rr.Handle("greeting", handler)
	receiver := ezbus.NewBus(br, rr)
	receiver.SubscribeMessage("sample-publisher", "greeting")
	receiver.Go()

	for {
		err := publisher.Publish(greeting{"hello ezbus"})
		if err != nil {
			logger.Error(err.Error())
		} else {
			logger.Info("Message published")
		}
		time.Sleep(time.Second * 3)
	}
	receiver.Stop()
	//publish messsage
}

func handler(m ezbus.Message) error {
	var g greeting
	json.Unmarshal(m.Body, &g)
	logger.Info(g.Text)
	return nil
}
