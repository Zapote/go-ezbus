# go-ezbus
Message bus for go

```go
r := ezbus.NewRouter()
r.Handle("placeOrder", func(message) {

})

b := ezbus.rabbitmq.NewBroker();

bus := ezbus.NewBus(r,b)

```