package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/silenteer/go-nats/nats"
)

type GetResult struct {
	RequestId   string           `json:"RequestId"`
	QueryParams nats.QueryParams `json:"QueryParams"`
	PathParams  nats.PathParams  `json:"PathParams"`
}

var config = nats.DefaultConfig()

func TestGetRequest(t *testing.T) {
	//1. setup server
	server := nats.NewServer(
		nats.SetConfig(nats.DefaultConfig()),
		nats.Routes(func(r nats.Router) {
			r.Register("GET", "/api/test/get/{id}", func(c *nats.Context, rq *nats.Request) *nats.Response {
				return nats.
					NewResBuilder().
					BodyJSON(&GetResult{
						c.RequestId(),
						c.QueryParams(),
						c.PathParams(),
					}).
					Build()
			})
		}),
	)

	go func() { server.Start() }()

	// wait for server ready
	time.Sleep(1 * time.Millisecond)
	defer server.Stop()

	//2. client request it
	request, _ := nats.NewReqBuilder().
		Get("/api/test/get/10002?from=10&to=90").
		Build()

	msg, err := nats.NewClient(config).SendRequest(nats.NewBackgroundContext(), request)
	if err != nil {
		t.Errorf("Error = %v", err)
	}

	result := &GetResult{}
	err = json.Unmarshal(msg.Body, &result)
	if err != nil {
		t.Errorf("json Unmarshal error  = %v", err)
	}
	//3. assert it
	assert.NotEmpty(t, result.RequestId, "Request Id not found")
	assert.Equal(t, result.PathParams["id"], "10002")
	assert.Equal(t, result.QueryParams["from"][0], "10")
	assert.Equal(t, result.QueryParams["to"][0], "90")
}

type PostRequest struct {
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
}

type PostResponse struct {
	Id       string `json:"id"`
	FullName string `json:"FullName"`
}

func TestPostRequestUsingHandlerJson(t *testing.T) {

	//1. setup server
	server := nats.NewServer(
		nats.SetConfig(nats.DefaultConfig()),
		nats.Routes(func(r nats.Router) {
			r.RegisterJson("POST", "/api/test/post/{id}", func(c *nats.Context, rq *PostRequest) (*PostResponse, error) {
				return &PostResponse{
					Id:       c.PathParams()["id"],
					FullName: fmt.Sprintf("%s %s", rq.FirstName, rq.LastName),
				}, nil
			})
		}),
	)

	go func() { server.Start() }()

	// wait for server ready
	time.Sleep(1 * time.Millisecond)
	defer server.Stop()

	//2. client request it
	potsRequest := &PostRequest{FirstName: "", LastName: ""}
	request, _ := nats.NewReqBuilder().
		Post("/api/test/post/1111").
		BodyJSON(potsRequest).
		Build()

	msg, err := nats.NewClient(config).SendRequest(nats.NewBackgroundContext(), request)
	if err != nil {
		t.Errorf("Error = %v", err)
	}

	result := &PostResponse{}
	err = json.Unmarshal(msg.Body, &result)
	if err != nil {
		t.Errorf("json Unmarshal error  = %v", err)
	}

	//3. assert it
	assert.NotEmpty(t, result.Id, "Request Id not found")
	assert.Equal(t, result.FullName, fmt.Sprintf("%s %s", potsRequest.FirstName, potsRequest.LastName))
}
