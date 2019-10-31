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

	header := http.Header{}
	header.Add("h1", "h1v1")
	header.Add("h1", "h1v2")
	request := &nats.Request{URL: "http://localhost:7073/api/app/testService/getQueryParams/pathValue?param1=p1&param2=p2", Headers: header, Method: "GET"}

	b, err1 := json.Marshal(request)
	if err1 != nil {
		fmt.Println(err1)
		return
	} else {
		fmt.Println(string(b))
	}

	// Send the request
	msg := nats.Response{}

	err := conn.Enc.Request("nats-service", request, &msg, time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Use the response
	log.Printf("Reply: %s", string(msg.Body))
}
