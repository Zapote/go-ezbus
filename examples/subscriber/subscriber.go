package main

import (
	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

func main() {
	b := rabbitmq.NewBroker("rabbitmq.example.subscriber")
	r := ezbus.NewRouter()

	bus := ezbus.NewBus(b, *r)
	bus.Subscribe("rabbitmq.example.receiver", "OrderPlaced")
	bus.Go()
}
