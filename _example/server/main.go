package main

import (
	"log"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/hello", Hello)

	r.Get("/user/{id}", Get)

	r.Put("/user/{id}", Put)

	r.Post("/user/{id}", Post)

	err := nats.ListenAndServe("test", r)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
