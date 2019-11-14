package main

import (
	"log"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	client := nats.NewClient("nats://127.0.0.1:4222")
	request, _ := nats.NewReqBuilder().
		Get("http://localhost:7073/api/app/testService/getQueryParams/pathValue?param1=p1&param2=p2").
		Subject("nats-service").
		AddHeader("h1", "h1v1").
		AddHeader("h1", "h1v2").
		Build()
	msg, err := client.Request(request)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
