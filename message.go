package ezbus

import "context"

// Message in EzBus
type Message struct {
	Headers map[string]string
	Body    []byte

	ctx context.Context
}

// NewMessage creates a new Message instance
// Using h as headers and b as body
func NewMessage(h map[string]string, b []byte) Message {
	return Message{Headers: h, Body: b}
}

// Context of the message. Background when the message carries none,
// which is the case for messages built outside of Bus.
func (m Message) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

// WithContext returns a copy of the message with ctx attached,
// in the same way as http.Request.WithContext.
func (m Message) WithContext(ctx context.Context) Message {
	if ctx == nil {
		panic("ezbus: nil context")
	}
	m.ctx = ctx
	return m
}
