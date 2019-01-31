package ezbus

// EzBus Message
type Message struct {
	Headers map[string]string
	Body    []byte
}

//NewMessage creates a new Message instance
// Using h as headers and b as body
func NewMessage(h map[string]string, b []byte) Message {
	return Message{h, b}
}

type subscription struct {
	endpoint    string
	messageName string
}

type subscriptions []subscription

//MessageHandler func for handling messsages
type MessageHandler = func(m Message)

//Middleware for router message handling
type Middleware = func(next MessageHandler) MessageHandler

//Constants for EzBus message headers
const (
	MessageFullname = "EzBus.MessageFullname"
	MessageName     = "EzBus.MessageName"
	UserPrincipal   = "EzBus.UserPrincipal"
	SendingHost     = "EzBus.SendingHost"
	SendingModule   = "EzBus.SendingModule"
	Destination     = "EzBus.Destination"
	TimeSent        = "EzBus.TimeSent"
)
