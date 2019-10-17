package main

import (
	"encoding/json"
	"log"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

//OrderPlaced event
type OrderPlaced struct {
	ID string
}

func main() {
	b := rabbitmq.NewBroker("rabbitmq.example.subscriber")
	r := ezbus.NewRouter()
	r.Handle("OrderPlaced", func(m ezbus.Message) error {
		log.Println("Orderplaced")
		return nil
	})

	r.Handle("OrderPlaced", func(m ezbus.Message) error {
		var po OrderPlaced
		json.Unmarshal(m.Body, &po)
		log.Printf(" %v OrderPlaced messages handled", po)
		return nil
	})

	bus := ezbus.NewBus(b, r)
	bus.SubscribeMessage("rabbitmq.example.receiver", "OrderPlaced")
	bus.Go()
}
