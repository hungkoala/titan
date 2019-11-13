package nats

import "github.com/go-chi/chi"

type Router interface {
	Routes(r chi.Router) // side effect function
}
