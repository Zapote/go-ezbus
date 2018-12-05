package ezbus

type Message struct {
	Headers map[string]string
	Body    []byte
}

func NewMessage(h map[string]string, b []byte) *Message {
	m := Message{h, b}
	return &m
}
