package main

import (
	"log"
	"net/http"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	addr := "nats://127.0.0.1:4222"
	subject := "test"
	path := "/user/2020?from=67&to=90"

	header := http.Header{}
	header.Set("Kaka", "value")
	header.Add("Kaka", "value2")
	header.Set("Data", "data value")
	msg, err := nats.SendJsonRequest(addr, subject, path, nil, header, "GET")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
