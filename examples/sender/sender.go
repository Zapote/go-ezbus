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

	for i := 0; i < 1; i++ {
		err := bus.Send("rabbitmq.example.receiver", placeOrder{"1337"})
		if err != nil {
			log.Println(err)
		}
	}
}

type placeOrder struct {
	ID string
}
