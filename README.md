# go-ezbus
This is a package for communication between services in a distrubuted architecture.
RabbitMQ as transport


pub/sub pattern
```code
                                       subscriber a
                                      /
                                     /
                                    /
command ----> publisher -- event --> subscriber b
                                    \
                                     \
                                      \
                                       subscriber c 
```

# code example
```go
type PlaceOrder struct {
	ID string
}

type OrderPlaced struct {
	ID string
}

r := ezbus.NewRouter()
r.Handle("placeOrder", func(message) {
    PlaceOrder po
    json.Unmarshal(m.Body, &po) 
    bus.publish(OrderPlaced {po.ID})
})

b := ezbus.NewBroker("my.queue");
bus := ezbus.NewBus(b, r)

bus.Go()

```