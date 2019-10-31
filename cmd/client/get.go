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

	request := &nats.Request{URL: "/hello", Headers: http.Header{}, Method: "GET"}

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
