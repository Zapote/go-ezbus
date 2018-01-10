package ezbus

// Bus for publish, send and receive messages
type Bus struct {
	endpoint string
	broker   Broker
}

// NewBus creates a bus instance.
func NewBus(endpoint string, b Broker) Bus {
	bus := Bus{
		endpoint: endpoint,
		broker:   b,
	}

	return bus
}

// Send message to destingation
func (b *Bus) Send(dest string, msg interface{}) error {
	return b.broker.send(dest, msg)
}
