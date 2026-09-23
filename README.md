
# go-ezbus [![CircleCI](https://circleci.com/gh/Zapote/go-ezbus/tree/master.svg?style=shield)](https://circleci.com/gh/zapote/go-ezbus/tree/master) 

<img src="logo.png" align="right" width="140" />

This is a package for communication between software components. It makes sending, publishing and receiving messages super easy!

Using RabbitMQ as transport for messages. More transports can and will (hopefully) be added

#### install
`go get github.com/zapote/go-ezbus`

## idea
Ezbus is great to use when working in a distrubuted system. Publish events when a software executes a command and let rest of the system know. 

Plugin new software components in your architecture without touching the existing ones.

Ezbus is super easy to use and will get you started in no time.

## code example
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
r.Handle("PlaceOrder", func(message ezbus.Message) {
    var po PlaceOrder
    json.Unmarshal(m.Body, &po) 
    bus.Publish(OrderPlaced {po.ID})
})

//create a rabbitmq broker
b := rabbitmq.NewBroker("my-queue", rabbitmq.WithURL("amqp://guest:guest@localhost:5672"))

//create the bus with router and broker
bus := ezbus.NewBus(b, r)

//Go!
bus.Go()
```
## Traces and metrics

The bus is instrumented with the OpenTelemetry API. Install a tracer provider and a meter provider in the service and it starts reporting. Without them everything is a no-op.

Traces: `SendContext` and `PublishContext` start a producer span and write `traceparent` into the message headers. A handler gets the consumer span through `m.Context()`. Pass it on to database calls, HTTP requests and new messages, or the trace stops there.

The library's own log lines about a message, failed attempts and the move to the error queue, are logged with the message context, so a slog handler that adds `trace_id` from the context puts them on the trace too.

Metrics:

| Name | Kind | What |
| --- | --- | --- |
| `messaging.client.sent.messages` | counter | sends and publishes, by destination and outcome |
| `messaging.client.operation.duration` | histogram, s | time to send or publish |
| `messaging.client.consumed.messages` | counter | handled messages, by queue, message name and outcome: `ok`, `error_queue` or `discarded` |
| `messaging.process.duration` | histogram, s | time to handle a message, all attempts included |
| `ezbus.process.attempts` | histogram | attempts it took; five means the message ended on the error queue |

Queue lengths are not here. RabbitMQ reports them itself through `rabbitmq_prometheus`.
