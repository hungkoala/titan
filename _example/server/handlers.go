package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Handler struct {
	userService *UserService
}

func NewHandler(userService *UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
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
	_, _ = w.Write(e)
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	_, _ = w.Write(body)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	_, _ = w.Write(body)
}

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello world"))
}
