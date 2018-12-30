package main

import (
	"encoding/json"
	"log"

	ezbus "github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

type PlaceOrder struct {
	ID string
}

func main() {
	b, err := rabbitmq.NewBroker("rabbitmq.example.receiver")

	if err != nil {
		log.Fatal("Failed to create RabbitMQ broker: ", err)
	}

	r := ezbus.NewRouter()

	r.Handle("PlaceOrder", func(m ezbus.Message) {
		var po PlaceOrder
		json.Unmarshal(m.Body, &po)

		log.Println(po)
	})

	bus := ezbus.NewBus(b, *r)

	bus.Go()
}
