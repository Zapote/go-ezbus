package main

import (
	"encoding/json"
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

type OrderPlaced struct {
	ID string
}

func main() {
	b := rabbitmq.NewBroker("rabbitmq.example.subscriber")
	r := ezbus.NewRouter()
	r.Handle("OrderPlaced", func(m ezbus.Message) {
		log.Println("Orderplaced")
	})

	r.Handle("OrderPlaced", func(m ezbus.Message) {
		var po OrderPlaced
		json.Unmarshal(m.Body, &po)
		log.Printf(" %v OrderPlaced messages handled", po)
	})

	bus := ezbus.NewBus(b, r)
	bus.SubscribeMessage("rabbitmq.example.receiver", "OrderPlaced")
	bus.Go()
}
