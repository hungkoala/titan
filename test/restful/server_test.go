package restful

import (
	context2 "context"
	"crypto/tls"
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

const cert string = "-----BEGIN CERTIFICATE-----\nMIIDYDCCAkigAwIBAgIJANC1VYI4EBFQMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV\nBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX\naWRnaXRzIFB0eSBMdGQwHhcNMTkxMTE4MDkxODIwWhcNMjAxMTE3MDkxODIwWjBF\nMQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50\nZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAyzIneTixRI5D9khw1EtHljyRx6Ij2ra/87hi5zkwgJoMz7O+oiO/igwk\nsC4+Yg1kEVDBusI1JiMgXm3KdI++JQyNIorIpurFKlhI2GX13LD4ogOBBrzwp6RU\nYYe1MFsrNqzJUDMPM9/orwpYBR3cXFAO8KMkFHnv2yBkkaK6gpqXbIlcls+xAomv\nnnBazIyOC3Mt9vvgv921H0K12I3UYikGYUaUqG/XzN4taVQ/gGV8J239qgKD8r6a\nnFOVxghRcKTkKeeSSpBZGcVCmgOX+H5y+ahpcv8rkFhsUE6U5pA3A5NLgQSeFu2z\nSh8+XSJjV5PYyGXev+gNRXvVZWsdFwIDAQABo1MwUTAdBgNVHQ4EFgQU/IOu32Dk\nukM3JHO9ClfMM2aLPN0wHwYDVR0jBBgwFoAU/IOu32DkukM3JHO9ClfMM2aLPN0w\nDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAEeomIjOwn35UBHO5\nUWlKDGcnt6fS4g7Pp/6uJ773HY5qNGSTrb+5A0eVQ4Ml99JTAG6Kt9flgbg95h8f\nmHrydQIp1iIGoDMw1x45DTWBxuxLl/iFKVH9tYU/9si8Ojgo2FaEOU60wvnSdyuj\nbiMoo7AX0LEQ6/Hfn/Xh3++DbdpxDMiZdL5stKK0KsIj84e1haAXiBMJAfJrQ3om\nHVnnrjbIvs4249qo3g20mAajlukks74b26Kq2I2XMeUlHszOwuJynVitIupExjoS\nOefdgIDzwFAmAgVgA/i3QTesmh5KuY4nQbHM3XHe7cKDejR04JQvgUs1SA8fH/qF\nlFM6xg==\n-----END CERTIFICATE-----"
const key string = "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDLMid5OLFEjkP2\nSHDUS0eWPJHHoiPatr/zuGLnOTCAmgzPs76iI7+KDCSwLj5iDWQRUMG6wjUmIyBe\nbcp0j74lDI0iisim6sUqWEjYZfXcsPiiA4EGvPCnpFRhh7UwWys2rMlQMw8z3+iv\nClgFHdxcUA7woyQUee/bIGSRorqCmpdsiVyWz7ECia+ecFrMjI4Lcy32++C/3bUf\nQrXYjdRiKQZhRpSob9fM3i1pVD+AZXwnbf2qAoPyvpqcU5XGCFFwpOQp55JKkFkZ\nxUKaA5f4fnL5qGly/yuQWGxQTpTmkDcDk0uBBJ4W7bNKHz5dImNXk9jIZd6/6A1F\ne9Vlax0XAgMBAAECggEALSVikdNfx2yYewLTVse3CxFADovezXxnH55rExaoyRnx\nGMDF7T5mEyTpjd9oat6wygwYTwdRSbzqNzDLl6RMSe0E+pS9SiDFV8gvvyzAOJ11\nUIHYzAd0rLqdKOI/BpRrAIXZYKFHkm4AQ8vXjKN6z2PWPe7xWD9TZGtJDPaL2/JZ\npMnqmRXtpuno6aRi3oYGUWXUbROX70XP6P9HwbZKe4klMblS3wBfi1OPmU9Zt7yB\neki+jNOLHzphJYNWE+r3W0zu3+1Yw+pXU8H6MTvkn5CYX0enG/PP6I9gXIdbEMjY\n19UxpARrZ255ZIK3yHk6cP1dOiLyxkeEtfEajRGqcQKBgQDrzHxmOY21QaENzDhO\nbg96HoaLvmfCOC8QPS7e15miV+8RZRIXhlZHFOqbyzwLzOauDwRJ3Brjk2+eL256\nUJpSU/f9AlCVZcsxWMSgAd68miKbIA5wTZmuedN7DasckNzzQgDvlHtfINf2PkeC\nJ+qQCNch5wLHGo5LIMz+ukrfXwKBgQDcmqAyvPXqBFKh3Tyyxt6+hq/gSMnCY6Uq\nImvfkN9yO2PXFPNFBp27MzYh8aivESWBnuw+6hX4r83YcEMWBYrFq2UCovQIdRwp\nEh3oolbuQqWdcoo7MCPTQ7cw3uxDbjFCF3AMwId/fKdU4LtNG92io5Hme0W3Of6e\nF/4aguV1SQKBgCkipj0LI06QoXEPxG7iQm7yblRoph86v/McSVX01Md+gaVONYbH\nF7wUyQzeup3wY/nPgtcDv+kdqmY1LhfGgfWE0olf4wD9HiKAsuSbDulmFk1rnTk4\nQGwwspUQAF7eYr1JMXKaO5+P0j0SBlWNcx0nfahbbZ+gYVx332s8wp0PAoGBAMRd\nXvvK95q2/lbWd5ErNFqjySn7oJxH1l0LBrqaWkt0UgrBu0lV+lEH5MeSNHSg7qHS\ntLfL5oLW+oQOaajQhhYt2lvecRqWI9rrJXRODNNIv+LGcgT9dOY5AHef9u6Ox4nt\nEvBG8FWqv8ftwsuAYmjC8LwYPpY6KUrQUH+IxHcBAoGBANzq2sP524Nf8McYk/4d\n38WowBsPSW55lrQ/ol0IoEaGAlCisNS6wECkKbViJIIVaSpo26CE0tJ49r2GXduo\nTWM52KnHfZmSJS5wjjlpcmhDS8sEOLaZ9RglYHKDLybWYAOeT5+uNcwbDz4nCaQs\ny6WGdI9LmEnylt8kQaNWKauH\n-----END PRIVATE KEY-----"

func TestGetRequest(t *testing.T) {
	//1. setup server
	port := "6968"
	server := restful.NewServer(restful.Port(port),
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

func TestHttpsHealthEndPoint(t *testing.T) {
	//1. setup server
	port := "6968"

	server := restful.NewServer(
		restful.Port(port),
		restful.TlsEnable(true),
		restful.TlsCert(cert),
		restful.TlsKey(key))

	testServer := test.NewTestServer(t, server)
	testServer.Start()
	defer testServer.Stop()

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}

	resp, err := client.Get(fmt.Sprintf("https://localhost:%s/health", port))

	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)

	result := &titan.Health{}
	jsonErr := json.Unmarshal(body, &result)
	require.NoError(t, jsonErr, "Unmarshal response error")

	//3. assert it
	assert.NotEmpty(t, result.Status, "UP")
}

func TestHttpsClient(t *testing.T) {
	//1. setup server
	port := "6967"

	server := restful.NewServer(
		restful.Port(port),
		restful.TlsEnable(true),
		restful.TlsCert(cert),
		restful.TlsKey(key))

	testServer := test.NewTestServer(t, server)
	testServer.Start()
	defer testServer.Stop()

	ctx := titan.NewContext(context2.Background())
	var result titan.Health

	request, err := titan.NewReqBuilder().Get("/health").Build()
	require.Nil(t, err)

	err = restful.
		NewClient(fmt.Sprintf("https://localhost:%s", port)).
		SendAndReceiveJson(ctx, request, &result)
	require.Nil(t, err)
	assert.NotEmpty(t, result.Status, "UP")

}
