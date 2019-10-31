package main

import (
	"encoding/json"
	"gitlab.com/silenteer/go-nats/nats"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	r.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		w.Write(e)
	})

	r.Put("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		w.Write(body)
	})

	r.Post("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		w.Write(body)
	})

	server := &nats.Server{Handler: r}
	server.ListenAndServe()
	//select {}
}
