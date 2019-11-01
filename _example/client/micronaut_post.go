package main

import (
	"log"
	"net/http"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	addr := "nats://127.0.0.1:4222"
	subject := "nats-service"
	path := "http://localhost:7073/api/app/testService/post"

	body := struct {
		Param1 string `json:"param1"`
		Param2 string `json:"param2"`
	}{
		"value1",
		"value2",
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json; charset=utf-8")

	msg, err := nats.SendJsonRequest(addr, subject, path, body, header, "POST")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
