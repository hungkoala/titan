package server

import (
	"gitlab.com/silenteer/go-nats/nats"
)

type Router struct {
	h *Handler
}

func NewRouter(h *Handler) *Router {
	return &Router{h: h}
}

func (r *Router) Routes(c nats.Router) {
	c.Register("GET", "/hello", r.h.Hello)
	c.Register("GET", "/user/{id}", r.h.Get)
	c.Register("PUT", "/user/{id}", r.h.Put)
	c.RegisterJson("POST", "/user/{id}", r.h.Post)
}
