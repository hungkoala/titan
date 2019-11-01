package main

import (
	"log"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	addr := "nats://127.0.0.1:4222"
	subject := "test"
	msg, err := nats.Get(addr, subject, "/hello")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
