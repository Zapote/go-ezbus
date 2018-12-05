package ezbus

import "log"

type Handler = func(m *Message)

type Router struct {
	handlers map[string]Handler
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	r := Router{make(map[string]Handler)}
	return &r
}

func (r *Router) Handle(m string, h Handler) {
	r.handlers[m] = h
}

func (r *Router) handle(n string, m *Message) {
	h, ok := r.handlers[n]
	if ok {
		h(m)
	} else {
		log.Printf("No handler found for message name %s", n)
	}
}
