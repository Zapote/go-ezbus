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
	b := rabbitmq.NewBroker("go-ezbus-receiver")
	r := ezbus.NewRouter()
	bus := ezbus.NewBus(b, r)

	r.Middleware(func(next func(m ezbus.Message) error) func(m ezbus.Message) error {
		return func(m ezbus.Message) error {
			t := time.Now()
			err := next(m)
			log.Println(fmt.Sprintf("Message handled in %.2f us", time.Since(t).Seconds()*1000000))
			return err
		}
	})

	r.Handle("PlaceOrder", func(m ezbus.Message) error {
		var po PlaceOrder
		json.Unmarshal(m.Body, &po)
		n++
		log.Println(fmt.Sprintf(" %d PlaceOrder messages handled", n))

		return bus.Publish(OrderPlaced{po.ID})
	})

	ch := make(chan (interface{}))

	bus.Go()

	<-ch
}

//PlaceOrder command
type PlaceOrder struct {
	ID string
}

//OrderPlaced event
type OrderPlaced struct {
	ID string
}
