package main

import (
	"log"

	"gitlab.com/silenteer/go-nats/nats"
)

func main() {
	client := nats.NewClient("nats://127.0.0.1:4222")
	request, _ := nats.NewReqBuilder().
		Post("/user/2020?from=67&to=90").
		Subject("test").
		SetHeader("Kaka", "value").
		SetHeader("Data", "data value").
		AddHeader("Kaka", "value2").
		BodyJSON(struct {
			Name  string
			Email string
		}{
			"hung",
			"hung@silentium.io",
		}).
		Build()
	msg, err := client.Request(request)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reply: %s", string(msg.Body))
}
