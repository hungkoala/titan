package main

import (
	"encoding/json"
	"fmt"
	"gitlab.com/silenteer/go-nats/nats"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := "nats://127.0.0.1:4222"

	conn := nats.NewConnection(addr)

	defer conn.Enc.Close()
	defer conn.Conn.Close()

	request := &nats.Request{URL: "/user/2020?from=67&to=90", Headers: http.Header{}, Method: "GET"}
	request.Headers.Set("Kaka", "value")
	request.Headers.Add("Kaka", "value2")
	request.Headers.Set("Data", "data value")

	b, err1 := json.Marshal(request)
	if err1 != nil {
		fmt.Println(err1)
		return
	} else {
		fmt.Println(string(b))
	}

	// Send the request
	msg := nats.Response{}

	err := conn.Enc.Request("test", request, &msg, time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Use the response
	log.Printf("Reply: %s", string(msg.Body))
}
