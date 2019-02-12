package ezbus

// Broker interface
type Broker interface {
	Sender
	Publisher
	Receiver
	Subscriber
}

// Sender interface
type Sender interface {
	Send(dst string, m Message) error
}

// Publisher interface
type Publisher interface {
	Publish(m Message) error
}

//Receiver interface
type Receiver interface {
	Start(chan<- Message) error
	Stop() error
	Endpoint() string
}

//Subscriber interface
type Subscriber interface {
	Subscribe(endpoint string, messageName string) error
}
