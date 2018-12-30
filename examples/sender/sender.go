package main

import (
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

func main() {

	broker, err := rabbitmq.NewBroker("")
	if err != nil {
		log.Fatalf("NewBroker: %s", err)
	}
	bus := ezbus.NewSendOnlyBus(broker)

	bus.Send("rabbitmq.example.receiver", PlaceOrder{"1337"})
}

type PlaceOrder struct {
	ID string
}
