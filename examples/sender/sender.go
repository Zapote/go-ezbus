package main

import (
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

func main() {
	b := rabbitmq.NewBroker("")
	bus := ezbus.NewSendOnlyBus(b)

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
