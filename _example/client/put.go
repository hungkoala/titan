package main

import (
	"log"
	"net/http"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	addr := "nats://127.0.0.1:4222"
	subject := "test"
	path := "/user/1110"

	body := struct {
		Name  string
		Email string
	}{
		"hung",
		"hung@silentium.io",
	}
	header := http.Header{}
	header.Add("Content-Type", "application/json; charset=utf-8")

	msg, err := nats.SendJsonRequest(addr, subject, path, body, header, "POST")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
