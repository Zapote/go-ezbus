# go-ezbus
Message bus for go

```go
r := ezbus.NewRouter()
r.Handle("placeOrder", func(message) {
    bus.publish(OrderPlaced {ID:"123", Number=1000})
})

b := ezbus.rabbitmq.NewBroker("my.queue");
bus := ezbus.NewBus(r,b)



```