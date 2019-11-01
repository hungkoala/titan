package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world"))
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
		_, _ = w.Write(e)
	})

	r.Put("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		_, _ = w.Write(body)
	})

	r.Post("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		_, _ = w.Write(body)
	})

	err := nats.ListenAndServe("test", r)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
