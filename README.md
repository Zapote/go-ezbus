
# go-ezbus [![CircleCI](https://circleci.com/gh/Zapote/go-ezbus/tree/master.svg?style=shield)](https://circleci.com/gh/Zapote/go-ezbus/tree/master)

This is a package for communication between services in a distrubuted architecture. 

Using RabbitMQ as transport for messages. More transports can and will (hopefully) be added

###### Install
`go get github.com/google/uuid`

#### pub/sub pattern
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

#### code example
```go
//PlaceOrder command
type PlaceOrder struct {
	ID string
}

//OrderPlaced event
type OrderPlaced struct {
	ID string
}

//create message router
r := ezbus.NewRouter()

//register handler for message PlaceOrder
r.Handle("PlaceOrder", func(message) {
    PlaceOrder po
    json.Unmarshal(m.Body, &po) 
    bus.Publish(OrderPlaced {po.ID})
})

//create a rabbitmq broker
b := rabbitmq.NewBroker("my-queue");

//create the bus with router and broker
bus := ezbus.NewBus(b, r)

//Go!
bus.Go()
```