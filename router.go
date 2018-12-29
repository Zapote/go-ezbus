package ezbus

import (
	"log"
)

//Router calls correct MessageHandler func for an incoming message.
type Router struct {
	handlers    map[string]MessageHandler
	middleWares []Middleware
}

//NewRouter creates a new router instance.
func NewRouter() *Router {
	r := Router{make(map[string]MessageHandler), []Middleware{}}
	return &r
}

//Handle registers a MessageHandle func for a specific message (name).
func (r *Router) Handle(messageName string, h MessageHandler) {
	r.handlers[messageName] = h
}

//Middleware registers a Middleware func.
func (r *Router) Middleware(mw Middleware) {
	r.middleWares = append(r.middleWares, mw)
}

func (r *Router) handle(n string, m Message) {
	handler, ok := r.handlers[n]
	if ok {
		l := len(r.middleWares) - 1

		for i := l; i >= 0; i-- {
			handler = r.middleWares[i](handler)
		}

		handler(m)
	} else {
		log.Printf("No handler found for message namned '%s'", n)
	}
}
