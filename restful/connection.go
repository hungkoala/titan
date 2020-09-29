package restful

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/prometheus/common/log"

	"gitlab.com/silenteer-oss/titan"
)

type Connection struct {
	client    *http.Client
	discovery Discovery
	//add    string // example : https://192.168.1.10:8080/
}

func (c *Connection) Close() {

}

func (c *Connection) Drain() {

}

func NewConnection(discovery Discovery) titan.IConnection {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}

	return &Connection{
		discovery: discovery,
		client:    client,
	}
}

func (c *Connection) SendRequest(rq *titan.Request, subject string) (*titan.Response, error) {
	request, err := titan.NatsRequestToHttpRequest(rq)
	if err != nil {
		return nil, err
	}
	request.Header.Del("application/json")

	add, err := c.discovery.LookupService(subject)
	if err != nil {
		return nil, err
	}

	urlString := fmt.Sprintf("%s/%s", strings.TrimSuffix(add, "/"), strings.TrimPrefix(rq.URL, "/"))
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	request.URL = u

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &titan.Response{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Headers:    response.Header,
		Body:       body,
	}, nil
}

func (c *Connection) Publish(subject string, v interface{}) error {
	log.Error("Not implemented http Publish")
	return nil
}

func (c *Connection) Flush() error {
	return nil
}
