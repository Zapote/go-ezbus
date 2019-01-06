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
r := ezbus.NewRouter()
r.Handle("placeOrder", func(message) {
    bus.publish(OrderPlaced {ID:"123", Number=1000})
})

b := ezbus.rabbitmq.NewBroker("my.queue");
bus := ezbus.NewBus(r,b)

```