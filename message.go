package ezbus

type Message struct {
	Headers map[string]string
	Body    []byte
}

//NewMessage creates a new Message instance
// Using h as headers and b as body
func NewMessage(h map[string]string, b []byte) Message {
	return Message{h, b}
}
