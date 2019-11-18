package test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	oNats "github.com/nats-io/nats.go"
	"gitlab.com/silenteer/go-nats/nats"
)

type GetResult struct {
	RequestId   string           `json:"RequestId"`
	QueryParams nats.QueryParams `json:"QueryParams"`
	PathParams  nats.PathParams  `json:"PathParams"`
}

func TestGetRequest(t *testing.T) {
	//1. setup server
	server := nats.NewServerAndStartRoutine(
		nats.Subject("test"),
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

	defer server.Stop()

	//2. client request it
	request, _ := nats.NewReqBuilder().
		Get("/api/test/get/10002?from=10&to=90").
		Subject("test").
		Build()

	msg, err := nats.NewClient(oNats.DefaultURL).Request(request)
	if err != nil {
		t.Errorf("Error = %v", err)
	}

	result := &GetResult{}
	err = json.Unmarshal(msg.Body, &result)
	if err != nil {
		t.Errorf("json Unmarshal error  = %v", err)
	}
	//3. assert it
	assert.NotEmpty(t, result.RequestId, "RequestId not found")
	assert.Equal(t, result.PathParams["id"], "10002")
	assert.Equal(t, result.QueryParams["from"][0], "10")
	assert.Equal(t, result.QueryParams["to"][0], "90")
}
