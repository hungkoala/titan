package main

import (
	"fmt"
	"os"

	"gitlab.com/silenteer/go-nats/log"
	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	logger := log.DefaultLogger(nil)
	userService := NewUserService()
	handler := NewHandler(userService)
	router := NewRouter(handler)

	server, err := nats.NewServer(
		nats.Routes(router),
		nats.Subject("test"),
		nats.Logger(logger),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
		os.Exit(1)
	}

	err = server.Start()
	if err != nil {
		logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
		os.Exit(1)
	}
}
