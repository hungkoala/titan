package restful

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"gitlab.com/silenteer-oss/titan/restful"

	"gitlab.com/silenteer-oss/titan/test"

	"gitlab.com/silenteer-oss/titan"
)

type GetResult struct {
	RequestId   string            `json:"RequestId"`
	QueryParams titan.QueryParams `json:"QueryParams"`
	PathParams  titan.PathParams  `json:"PathParams"`
}

var H = func(c *titan.Context, rq *titan.Request) *titan.Response {
	return titan.NewResBuilder().
		BodyJSON(&GetResult{
			c.RequestId(),
			c.QueryParams(),
			c.PathParams(),
		}).
		Build()
}

func TestGetRequest(t *testing.T) {
	//1. setup server
	port := "6968"
	server := restful.NewServer(port,
		restful.Routes(func(r titan.Router) {

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

	testServer := test.NewTestServer(t, server)
	testServer.Start()
	defer testServer.Stop()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/service/test/get/10002?from=10&to=90", port))

	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)

	result := &GetResult{}
	jsonErr := json.Unmarshal(body, &result)
	require.NoError(t, jsonErr, "Unmarshal response error")

	//3. assert it
	assert.NotEmpty(t, result.RequestId, "Request Id not found")
	assert.Equal(t, result.PathParams["id"], "10002")
	assert.Equal(t, result.QueryParams["from"][0], "10")
	assert.Equal(t, result.QueryParams["to"][0], "90")
}
