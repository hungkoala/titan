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
	c.MethodFunc("GET", "/hello", r.h.Hello)
	c.MethodFunc("GET", "/user/{id}", r.h.Get)
	c.MethodFunc("PUT", "/user/{id}", r.h.Put)
	c.MethodFunc("POST", "/user/{id}", r.h.Post)
}
