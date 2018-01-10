package ezbus

// Broker interface
type Broker interface {
	send(dest string, msg interface{}) error
}

// Sender interface
type Sender interface {
	send(dest string, msg interface{}) error
}

// Publisher interface
type Publisher interface {
	publish(msg interface{}) error
}
