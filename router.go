package ezbus

import (
	"log"
)

//MessageHandler func for handling messsages
type MessageHandler = func(m Message)

//Middleware for router message handling
type Middleware = func(next MessageHandler) MessageHandler

//Router routes message to correct MessageHandler func.
type Router interface {
	Handle(messageName string, h MessageHandler)
	Middleware(mw Middleware)
	Receive(n string, m Message)
}

type router struct {
	handlers    map[string]MessageHandler
	middlewares []Middleware
}

//NewRouter creates a new router instance.
func NewRouter() Router {
	r := router{
		handlers:    make(map[string]MessageHandler),
		middlewares: []Middleware{},
	}
	return &r
}

//Handle registers a ezbus.MessageHandler h, for specific messagename, n.
func (r *router) Handle(n string, h MessageHandler) {
	r.handlers[n] = h
}

//Middleware registers a Middleware func.
func (r *router) Middleware(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

//Receive tries to find a registered handler for ezbus.Message m,  based on message name, n
func (r *router) Receive(n string, m Message) {
	handler, ok := r.handlers[n]
	if !ok {
		log.Printf("No handler found for message namned '%s'", n)
		return
	}

	l := len(r.middlewares) - 1
	for i := l; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	handler(m)
}
