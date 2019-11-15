package main

import (
	"gitlab.com/silenteer/go-nats/_example/server"

	"gitlab.com/silenteer/go-nats/log"
	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	logger := log.DefaultLogger(nil)
	userService := server.NewUserService()
	handler := server.NewHandler(userService)
	router := server.NewRouter(handler)

	nats.NewServerAndStart(
		nats.RouterProvider(router),
		nats.Subject("test"),
		nats.Logger(logger),
	)
}
