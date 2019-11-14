package main

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
	c.Get("/hello", r.h.Hello)
	c.Get("/user/{id}", r.h.Get)
	c.Put("/user/{id}", r.h.Put)
	c.Post("/user/{id}", r.h.Post)
}
