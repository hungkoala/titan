package server

import (
	"encoding/json"

	"gitlab.com/silenteer/go-nats/nats"

	"github.com/go-chi/chi"
)

type Handler struct {
	userService *UserService
}

func NewHandler(userService *UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) Get(c *nats.Context, r *nats.SRequest) *nats.Response {
	data := struct {
		RequestId     interface{}         `json:"RequestId"`
		RequestParams map[string][]string `json:"RequestParams"`
		RouteParams   chi.RouteParams     `json:"RouteParams"`
	}{
		"",
		r.RequestParams,
		r.RouteParams,
	}

	e, _ := json.Marshal(data)
	return nats.
		NewResBuilder().
		Body(e).
		Build()
}

func (h *Handler) Put(c *nats.Context, r *nats.SRequest) *nats.Response {
	return nats.
		NewResBuilder().
		Body(r.Body).
		Build()
}

func (h *Handler) Post(c *nats.Context, r *nats.SRequest) *nats.Response {
	return nats.
		NewResBuilder().
		Body(r.Body).
		Build()
}

func (h *Handler) Hello(c *nats.Context, r *nats.SRequest) *nats.Response {
	return nats.
		NewResBuilder().
		Body([]byte("hello world")).
		Build()
}
