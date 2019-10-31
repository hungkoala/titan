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

	rq := struct {
		Param1 string `json:"param1"`
		Param2 string `json:"param2"`
	}{
		"value1",
		"value2",
	}

	body, err2 := json.Marshal(rq)
	if err2 != nil {
		fmt.Println(err2)
		return
	} else {
		fmt.Println(string(body))
	}

	request := &nats.Request{URL: "http://localhost:7073/api/app/testService/post", Headers: http.Header{}, Method: "POST"}
	request.Body = body
	request.Headers.Add("Content-Type", "application/json; charset=utf-8")

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
