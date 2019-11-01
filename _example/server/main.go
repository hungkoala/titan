package main

import (
	"log"

	"gitlab.com/silenteer/go-nats/_example/server/controller"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/hello", controller.Hello)

	r.Get("/user/{id}", controller.Get)

	r.Put("/user/{id}", controller.Put)

	r.Post("/user/{id}", controller.Post)

	err := nats.ListenAndServe("test", r)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
