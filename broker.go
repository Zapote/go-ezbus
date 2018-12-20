package ezbus

// Broker interface
type Broker interface {
	Sender
	Publisher
	Receiver
}

// Sender interface
type Sender interface {
	Send(dest string, m Message) error
}

// Publisher interface
type Publisher interface {
	Publish(m Message) error
}

type Receiver interface {
	Start(chan Message) error
	Stop()
}
