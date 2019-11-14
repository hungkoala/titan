package main

import (
	"log"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	request, _ := nats.NewReqBuilder().
		Get("/hello").
		Subject("test").
		Build()
	client := nats.NewClient("nats://127.0.0.1:4222")
	msg, err := client.Request(request)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
