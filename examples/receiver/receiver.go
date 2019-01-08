package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	ezbus "github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/rabbitmq"
)

func main() {
	n := 0
	b := rabbitmq.NewBroker("rabbitmq.example.receiver")
	r := ezbus.NewRouter()

	r.Middleware(func(next func(m ezbus.Message)) func(m ezbus.Message) {
		return func(m ezbus.Message) {
			t := time.Now()
			next(m)
			log.Println(fmt.Sprintf("Message handled in %v us", time.Since(t).Seconds()*1000000))
		}
	})

	r.Handle("PlaceOrder", func(m ezbus.Message) {
		var po PlaceOrder
		json.Unmarshal(m.Body, &po)
		n++
		log.Println(fmt.Sprintf(" %d OrderPlaced messages handled", n))
	})

	bus := ezbus.NewBus(b, *r)

	bus.Go()
}

type PlaceOrder struct {
	ID string
}
