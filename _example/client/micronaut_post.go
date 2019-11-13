package main

import (
	"log"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	client := nats.NewClient("nats://127.0.0.1:4222")
	request, _ := nats.NewRequestBuilder().
		Post("http://localhost:7073/api/app/testService/post").
		Subject("nats-service").
		BodyJSON(struct {
			Param1 string `json:"param1"`
			Param2 string `json:"param2"`
		}{
			"value1",
			"value2",
		}).
		Build()
	msg, err := client.Request(request)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
