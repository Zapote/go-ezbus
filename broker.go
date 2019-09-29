package ezbus

// Broker interface
type Broker interface {
	Send(dst string, m Message) error
	Publish(m Message) error
	Start(handle MessageHandler) error
	Stop() error
	Endpoint() string
	Subscribe(endpoint string, messageName string) error
}
