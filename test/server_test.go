package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gitlab.com/silenteer/titan"

	"github.com/stretchr/testify/assert"
)

type GetResult struct {
	RequestId   string            `json:"RequestId"`
	QueryParams titan.QueryParams `json:"QueryParams"`
	PathParams  titan.PathParams  `json:"PathParams"`
}

func TestGetRequest(t *testing.T) {
	//1. setup server
	server := titan.NewServer("api.service.test",
		titan.Routes(func(r titan.Router) {
			r.Register("GET", "/api/service/test/get/{id}", func(c *titan.Context, rq *titan.Request) *titan.Response {
				return titan.NewResBuilder().
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
	time.Sleep(5 * time.Millisecond)
	defer server.Stop()

	//2. client request it
	request, _ := titan.NewReqBuilder().
		Get("/api/service/test/get/10002?from=10&to=90").
		Build()

	msg, err := titan.GetDefaultClient().SendRequest(titan.NewBackgroundContext(), request)
	if err != nil {
		t.Errorf("Error = %v", err)
	}

	result := &GetResult{}
	jsonErr := json.Unmarshal(msg.Body, &result)
	if jsonErr != nil {
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
	server := titan.NewServer("api.service.test",
		titan.Routes(func(r titan.Router) {
			r.RegisterJson("POST", "/api/service/test/post/{id}", func(c *titan.Context, rq *PostRequest) (*PostResponse, error) {
				return &PostResponse{
					Id:       c.PathParams()["id"],
					FullName: fmt.Sprintf("%s %s", rq.FirstName, rq.LastName),
				}, nil
			})
		}),
	)

	go func() { server.Start() }()

	// wait for server ready
	time.Sleep(5 * time.Millisecond)
	defer server.Stop()

	//2. client request it
	potsRequest := &PostRequest{FirstName: "", LastName: ""}
	request, _ := titan.NewReqBuilder().
		Post("/api/service/test/post/1111").
		BodyJSON(potsRequest).
		Build()

	msg, err := titan.GetDefaultClient().SendRequest(titan.NewBackgroundContext(), request)
	if err != nil {
		t.Errorf("Error = %v", err)
	}

	result := &PostResponse{}
	jsonErr := json.Unmarshal(msg.Body, &result)
	if jsonErr != nil {
		t.Errorf("json Unmarshal error  = %v", err)
	}

	//3. assert it
	assert.NotEmpty(t, result.Id, "Request Id not found")
	assert.Equal(t, result.FullName, fmt.Sprintf("%s %s", potsRequest.FirstName, potsRequest.LastName))
}
