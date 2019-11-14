package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"gitlab.com/silenteer/go-nats/nats"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Handler struct {
	userService *UserService
}

func NewHandler(userService *UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) Get(r *http.Request) *nats.Response {
	data := struct {
		RequestId     interface{}         `json:"RequestId"`
		RequestParams map[string][]string `json:"RequestParams"`
		URLParams     chi.RouteParams     `json:"URLParams"`
	}{
		r.Context().Value(middleware.RequestIDKey),
		r.URL.Query(),
		chi.RouteContext(r.Context()).URLParams,
	}

	e, _ := json.Marshal(data)
	return nats.
		NewResBuilder().
		Body(e).
		Build()
}

func (h *Handler) Put(r *http.Request) *nats.Response {
	body, _ := ioutil.ReadAll(r.Body)
	return nats.
		NewResBuilder().
		Body(body).
		Build()
}

func (h *Handler) Post(r *http.Request) *nats.Response {
	body, _ := ioutil.ReadAll(r.Body)
	return nats.
		NewResBuilder().
		Body(body).
		Build()
}

func (h *Handler) Hello(r *http.Request) *nats.Response {
	return nats.
		NewResBuilder().
		Body([]byte("hello world")).
		Build()
}
