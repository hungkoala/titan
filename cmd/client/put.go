package main

import (
	"encoding/json"
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

	user := struct {
		Name  string
		Email string
	}{
		"hung",
		"hung@silentium.io",
	}

	body, _ := json.Marshal(user)

	req := &nats.Request{URL: "/user/1110", Headers: http.Header{}, Method: "PUT"}
	req.Headers.Set("Content-Type", "application/json; charset=utf-8")
	req.Body = body

	// Send the request
	msg := nats.Response{}

	err := conn.Enc.Request("test", req, &msg, time.Second)
	conn.Enc.Flush()
	if err != nil {
		log.Fatal(err)
	}

	// Use the response
	log.Printf("Reply: %s", string(msg.Body))
}
