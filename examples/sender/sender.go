package main

import (
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

func main() {
	b := rabbitmq.NewBroker("")
	bus, err := ezbus.NewSendOnlyBus(b)

	if err != nil {
		log.Panicf("Failed to create bus: %s", err)
	}

	for i := 0; i < 1000; i++ {
		err := bus.Send("rabbitmq.example.receiver", PlaceOrder{"1337"})
		if err != nil {
			log.Println(err)
		}
	}
}

type PlaceOrder struct {
	ID string
}
