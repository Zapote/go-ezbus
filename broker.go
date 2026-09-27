package ezbus

import "context"

// Broker interface
type Broker interface {
	Send(dst string, m Message) error
	Publish(m Message) error
	Start(handle MessageHandler) error
	Stop() error
	Shutdown(ctx context.Context) error
	Endpoint() string
	Subscribe(endpoint string, messageName string) error
}
