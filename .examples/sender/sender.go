package main

import (
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

func main() {
	b := rabbitmq.NewBroker("")
	r := ezbus.NewRouter()
	bus := ezbus.NewBus(b, r)
	bus.Go()

	for i := 0; i < 1; i++ {
		err := bus.Send("go-ezbus-receiver", PlaceOrder{"1337"})
		if err != nil {
			log.Println(err)
		}
	}
}

type PlaceOrder struct {
	ID string
}
