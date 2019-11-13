package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Get(w http.ResponseWriter, r *http.Request) {
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

func Put(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	_, _ = w.Write(body)
}

func Post(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	_, _ = w.Write(body)
}

func Hello(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello world"))
}
