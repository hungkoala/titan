package main

import "github.com/go-chi/chi"

type Router struct {
	h *Handler
}

func NewRouter(h *Handler) *Router {
	return &Router{h: h}
}

func (r *Router) Routes(c chi.Router) {
	c.Get("/hello", r.h.Hello)
	c.Get("/user/{id}", r.h.Get)
	c.Put("/user/{id}", r.h.Put)
	c.Post("/user/{id}", r.h.Post)
}
