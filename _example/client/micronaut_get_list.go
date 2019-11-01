package main

import (
	"log"
	"net/http"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	addr := "nats://127.0.0.1:4222"
	subject := "nats-service"
	path := "http://localhost:7073/api/app/testService/getQueryParams/pathValue?param1=p1&param2=p2"
	header := http.Header{}
	header.Add("h1", "h1v1")
	header.Add("h1", "h1v2")
	msg, err := nats.SendJsonRequest(addr, subject, path, nil, header, "GET")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
