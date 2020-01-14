package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

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
			}, titan.IsAnonymous())
		}),
	)

	testServer := NewTestServer(t, server)
	testServer.Start()
	defer server.Stop()

	//2. client request it
	request, _ := titan.NewReqBuilder().
		Get("/api/service/test/get/10002?from=10&to=90").
		Build()

	msg, err := titan.GetDefaultClient().SendRequest(titan.NewBackgroundContext(), request)
	require.NoError(t, err, "Sending Nats request error")

	result := &GetResult{}
	jsonErr := json.Unmarshal(msg.Body, &result)
	require.NoError(t, jsonErr, "Unmarshal response error")

	//3. assert it
	assert.NotEmpty(t, result.RequestId, "Request Id not found")
	assert.Equal(t, result.PathParams["id"], "10002")
	assert.Equal(t, result.QueryParams["from"][0], "10")
	assert.Equal(t, result.QueryParams["to"][0], "90")
}

func TestRegisterTopic(t *testing.T) {
	//1. setup server
	server := titan.NewServer("api.service.test",
		titan.Routes(func(r titan.Router) {
			r.RegisterTopic("GET_DATA", func(c *titan.Context) (*titan.Response, error) {
				return titan.NewResBuilder().
					BodyJSON(&GetResult{
						c.RequestId(),
						c.QueryParams(),
						c.PathParams(),
					}).
					Build(), nil
			})
		}),
	)

	testServer := NewTestServer(t, server)
	testServer.Start()
	defer server.Stop()

	//2. client request it
	request, _ := titan.NewReqBuilder().
		Post("/api/service/test/GET_DATA").
		Build()

	_, err := titan.GetDefaultClient().SendRequest(titan.NewBackgroundContext(), request)
	require.NoError(t, err, "Sending Nats request error")
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
			}, titan.IsAnonymous())
		}),
	)

	testServer := NewTestServer(t, server)
	testServer.Start()
	defer server.Stop()

	//2. client request it
	potsRequest := &PostRequest{FirstName: "", LastName: ""}
	request, _ := titan.NewReqBuilder().
		Post("/api/service/test/post/1111").
		BodyJSON(potsRequest).
		Build()

	msg, err := titan.GetDefaultClient().SendRequest(titan.NewBackgroundContext(), request)
	require.NoError(t, err, "Sending Nats request error")

	result := &PostResponse{}
	jsonErr := json.Unmarshal(msg.Body, &result)
	require.NoError(t, jsonErr, "Unmarshal response error")

	//3. assert it
	assert.NotEmpty(t, result.Id, "Request Id not found")
	assert.Equal(t, result.FullName, fmt.Sprintf("%s %s", potsRequest.FirstName, potsRequest.LastName))
}

type TestValidationRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type TestValidationResponse struct {
	FullName string `json:"fullName"`
}

func TestValidator(t *testing.T) {
	//1. setup server
	server := titan.NewServer("api.service.test",
		titan.Routes(func(r titan.Router) {
			r.RegisterJson("POST", "/api/service/test/validation", func(c *titan.Context, rq *TestValidationRequest) (*TestValidationResponse, error) {
				return &TestValidationResponse{
					FullName: fmt.Sprintf("%s %s", rq.FirstName, rq.LastName),
				}, nil
			}, titan.IsAnonymous())
		}),
	)

	testServer := NewTestServer(t, server)
	testServer.Start()
	defer server.Stop()

	//2. client request it
	potsRequest := &TestValidationRequest{FirstName: "", LastName: ""}
	request, _ := titan.NewReqBuilder().
		Post("/api/service/test/validation").
		BodyJSON(potsRequest).
		Build()

	_, err := titan.GetDefaultClient().SendRequest(titan.NewBackgroundContext(), request)
	require.Error(t, err, "Sending Nats request error")
	require.IsType(t, &titan.ClientResponseError{}, err)
	cerr, _ := err.(*titan.ClientResponseError)
	assert.NotNil(t, cerr.Response)
	assert.Equal(t, 400, cerr.Response.StatusCode)
}

type TestAuthorizationResponse struct {
}

func TestAuthorization(t *testing.T) {
	//1. setup server
	server := titan.NewServer("api.service.test",
		titan.Routes(func(r titan.Router) {
			r.RegisterJson("GET", "/api/service/test/authorization", func(c *titan.Context) (*TestAuthorizationResponse, error) {
				return &TestAuthorizationResponse{}, nil
			}, titan.Secured("admin"))
		}),
	)

	testServer := NewTestServer(t, server)
	testServer.Start()
	defer server.Stop()

	//2. client request it
	potsRequest := &TestValidationRequest{FirstName: "", LastName: ""}
	request, _ := titan.NewReqBuilder().
		Get("/api/service/test/authorization").
		BodyJSON(potsRequest).
		Build()

	_, err := titan.GetDefaultClient().SendRequest(titan.NewBackgroundContext(), request)
	require.Error(t, err, "Sending Nats request error")
	require.IsType(t, &titan.ClientResponseError{}, err)
	cerr, _ := err.(*titan.ClientResponseError)
	assert.NotNil(t, cerr.Response)
	assert.Equal(t, 401, cerr.Response.StatusCode)
}

func TestDefaultHandlers(t *testing.T) {
	ctx := titan.NewContext(context.Background())
	client := titan.GetDefaultClient()
	//1. setup server
	server := titan.NewServer("api.service.test")

	testServer := NewTestServer(t, server)
	testServer.Start()
	defer server.Stop()

	//2. test health endPoint
	request, _ := titan.NewReqBuilder().
		Get("/api/service/test/health").
		Build()

	var result titan.Health
	err := client.SendAndReceiveJson(ctx, request, &result)
	require.NoError(t, err, "Sending Nats request error")

	assert.Equal(t, result.Status, "UP")

	//3. test info endPoint
	_ = os.Setenv("BUILD_VERSION", "1.0")
	_ = os.Setenv("BUILD_DATE", "15/12/2019")
	_ = os.Setenv("BUILD_TAG", "mytag")

	request, _ = titan.NewReqBuilder().
		Get("/api/service/test/info").
		Build()

	var info titan.AppInfo
	err = client.SendAndReceiveJson(ctx, request, &info)
	require.NoError(t, err, "Sending Nats request error")

	assert.Equal(t, info.Build.Version, "1.0")
	assert.Equal(t, info.Build.Date, "15/12/2019")
	assert.Equal(t, info.Build.Tag, "mytag")
}

// todo: should create more test cases relate to exception/validation handling

type TestBody struct {
	Msg string `json:"msg"`
}

func TestMessageSubscriber(t *testing.T) {
	ctx := titan.NewContext(context.Background())
	client := titan.GetDefaultClient()
	var tb TestBody
	var perr error
	var serr error

	//1. setup server
	messageReceived := make(chan interface{})
	server := titan.NewServer("api.service.test",
		titan.Subscribe(func (ms *titan.MessageSubscriber) {
			ms.Register("test", "api.service.test", func(p *titan.Message) error {
				close(messageReceived)

				_, serr = p.Parse(&tb)
				if serr != nil {
					return serr
				}

				return nil
			})
		}),
	)

	testServer := NewTestServer(t, server)
	testServer.Start()
	defer testServer.Stop()

	//2. test publish
	perr = client.Publish(ctx, "test", TestBody{Msg: "test msg"})
	WaitOrTimedOut(t, messageReceived, "Message not received")

	require.Nil(t, perr)
	require.Nil(t, perr)
	require.EqualValues(t, "test msg", tb.Msg)
}
